/*
Copyright (C) 2022-2025 ApeCloud Co., Ltd

This file is part of KubeBlocks project

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package dataprotection

import (
	"context"
	"reflect"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	dpv1alpha1 "github.com/apecloud/kubeblocks/apis/dataprotection/v1alpha1"
	intctrlutil "github.com/apecloud/kubeblocks/pkg/controllerutil"
	dpbackup "github.com/apecloud/kubeblocks/pkg/dataprotection/backup"
	dptypes "github.com/apecloud/kubeblocks/pkg/dataprotection/types"
	dputils "github.com/apecloud/kubeblocks/pkg/dataprotection/utils"
)

// BackupScheduleReconciler reconciles a BackupSchedule object
type BackupScheduleReconciler struct {
	client.Client
	Scheme   *k8sruntime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=dataprotection.kubeblocks.io,resources=backupschedules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dataprotection.kubeblocks.io,resources=backupschedules/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dataprotection.kubeblocks.io,resources=backupschedules/finalizers,verbs=update

// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=cronjobs/status,verbs=get
// +kubebuilder:rbac:groups=batch,resources=cronjobs/finalizers,verbs=update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the backupschedule closer to the desired state.
func (r *BackupScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqCtx := intctrlutil.RequestCtx{
		Ctx:      ctx,
		Req:      req,
		Log:      log.FromContext(ctx).WithValues("backupSchedule", req.NamespacedName),
		Recorder: r.Recorder,
	}

	backupSchedule := &dpv1alpha1.BackupSchedule{}
	if err := r.Client.Get(reqCtx.Ctx, reqCtx.Req.NamespacedName, backupSchedule); err != nil {
		return intctrlutil.CheckedRequeueWithError(err, reqCtx.Log, "")
	}

	reqCtx.Log.V(1).Info("reconcile", "backupSchedule", req.NamespacedName)

	original := backupSchedule.DeepCopy()

	// handle finalizer
	res, err := intctrlutil.HandleCRDeletion(reqCtx, r, backupSchedule, dptypes.DataProtectionFinalizerName, func() (*ctrl.Result, error) {
		return nil, r.deleteExternalResources(reqCtx, backupSchedule)
	})
	if res != nil {
		return *res, err
	}

	if err = r.handleSchedule(reqCtx, backupSchedule); err != nil {
		if intctrlutil.IsTargetError(err, intctrlutil.ErrorTypeRequeue) {
			return intctrlutil.RequeueAfter(reconcileInterval, reqCtx.Log, "")
		}
		return r.patchStatusFailed(reqCtx, backupSchedule, "HandleBackupScheduleFailed", err)
	}

	return r.patchStatusAvailable(reqCtx, original, backupSchedule)
}

// SetupWithManager sets up the controller with the Manager.
func (r *BackupScheduleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	b := intctrlutil.NewControllerManagedBy(mgr).
		For(&dpv1alpha1.BackupSchedule{})

	// Compatible with kubernetes versions prior to K8s 1.21, only supports batch v1beta1.
	if dputils.SupportsCronJobV1() {
		b.Owns(&batchv1.CronJob{})
	} else {
		b.Owns(&batchv1beta1.CronJob{})
	}
	b.Watches(&dpv1alpha1.Backup{}, handler.EnqueueRequestsFromMapFunc(r.parseBackup))
	return b.Complete(r)
}

func (r *BackupScheduleReconciler) deleteExternalResources(
	reqCtx intctrlutil.RequestCtx,
	backupSchedule *dpv1alpha1.BackupSchedule) error {
	// delete cronjob resource
	cronJobList := &batchv1.CronJobList{}
	if err := r.Client.List(reqCtx.Ctx, cronJobList,
		client.InNamespace(backupSchedule.Namespace),
		client.MatchingLabels{
			dptypes.BackupScheduleLabelKey: backupSchedule.Name,
		},
	); err != nil {
		return err
	}
	for _, cronjob := range cronJobList.Items {
		if err := dputils.RemoveDataProtectionFinalizer(reqCtx.Ctx, r.Client, &cronjob); err != nil {
			return err
		}
		if err := intctrlutil.BackgroundDeleteObject(r.Client, reqCtx.Ctx, &cronjob); err != nil {
			// failed delete k8s job, return error info.
			return err
		}
	}
	return nil
}

// patchStatusAvailable patches backup policy status phase to available.
func (r *BackupScheduleReconciler) patchStatusAvailable(reqCtx intctrlutil.RequestCtx,
	origin, backupSchedule *dpv1alpha1.BackupSchedule) (ctrl.Result, error) {
	if !reflect.DeepEqual(origin.Spec, backupSchedule.Spec) {
		if err := r.Client.Update(reqCtx.Ctx, backupSchedule); err != nil {
			return intctrlutil.CheckedRequeueWithError(err, reqCtx.Log, "")
		}
	}
	// update status phase
	if backupSchedule.Status.Phase != dpv1alpha1.BackupSchedulePhaseAvailable ||
		backupSchedule.Status.ObservedGeneration != backupSchedule.Generation {
		patch := client.MergeFrom(backupSchedule.DeepCopy())
		backupSchedule.Status.ObservedGeneration = backupSchedule.Generation
		backupSchedule.Status.Phase = dpv1alpha1.BackupSchedulePhaseAvailable
		backupSchedule.Status.FailureReason = ""
		if err := r.Client.Status().Patch(reqCtx.Ctx, backupSchedule, patch); err != nil {
			return intctrlutil.CheckedRequeueWithError(err, reqCtx.Log, "")
		}
	}
	return intctrlutil.Reconciled()
}

// patchStatusFailed patches backup policy status phase to failed.
func (r *BackupScheduleReconciler) patchStatusFailed(reqCtx intctrlutil.RequestCtx,
	backupSchedule *dpv1alpha1.BackupSchedule,
	reason string,
	err error) (ctrl.Result, error) {
	if intctrlutil.IsTargetError(err, intctrlutil.ErrorTypeRequeue) {
		return intctrlutil.RequeueAfter(reconcileInterval, reqCtx.Log, "")
	}
	backupScheduleDeepCopy := backupSchedule.DeepCopy()
	backupSchedule.Status.Phase = dpv1alpha1.BackupSchedulePhaseFailed
	backupSchedule.Status.FailureReason = err.Error()
	if !reflect.DeepEqual(backupSchedule.Status, backupScheduleDeepCopy.Status) {
		if patchErr := r.Client.Status().Patch(reqCtx.Ctx, backupSchedule, client.MergeFrom(backupScheduleDeepCopy)); patchErr != nil {
			return intctrlutil.RequeueWithError(patchErr, reqCtx.Log, "")
		}
	}
	r.Recorder.Event(backupSchedule, corev1.EventTypeWarning, reason, err.Error())
	return intctrlutil.RequeueWithError(err, reqCtx.Log, "")
}

// handleSchedule handles backup schedules for different backup method.
func (r *BackupScheduleReconciler) handleSchedule(
	reqCtx intctrlutil.RequestCtx,
	backupSchedule *dpv1alpha1.BackupSchedule) error {
	backupPolicy, err := dputils.GetBackupPolicyByName(reqCtx, r.Client, backupSchedule.Spec.BackupPolicyName)
	if err != nil {
		return err
	}
	if err = r.patchScheduleMetadata(reqCtx, backupSchedule); err != nil {
		return err
	}
	// TODO: update the mcMgr param
	saName, err := EnsureWorkerServiceAccount(reqCtx, r.Client, backupSchedule.Namespace, nil)
	if err != nil {
		return err
	}
	scheduler := dpbackup.Scheduler{
		RequestCtx:           reqCtx,
		BackupSchedule:       backupSchedule,
		BackupPolicy:         backupPolicy,
		Client:               r.Client,
		Scheme:               r.Scheme,
		WorkerServiceAccount: saName,
	}
	return scheduler.Schedule()
}

func (r *BackupScheduleReconciler) patchScheduleMetadata(
	reqCtx intctrlutil.RequestCtx,
	backupSchedule *dpv1alpha1.BackupSchedule) error {
	if backupSchedule.Labels[dptypes.BackupPolicyLabelKey] == backupSchedule.Spec.BackupPolicyName {
		return nil
	}
	patch := client.MergeFrom(backupSchedule.DeepCopy())
	if backupSchedule.Labels == nil {
		backupSchedule.Labels = map[string]string{}
	}
	backupSchedule.Labels[dptypes.BackupPolicyLabelKey] = backupSchedule.Spec.BackupPolicyName
	return r.Client.Patch(reqCtx.Ctx, backupSchedule, patch)
}

func (r *BackupScheduleReconciler) parseBackup(ctx context.Context, object client.Object) []reconcile.Request {
	backup := object.(*dpv1alpha1.Backup)
	backupScheduleName := dptypes.BackupScheduleLabelKey
	if backup.Labels[dptypes.BackupTypeLabelKey] == string(dpv1alpha1.BackupTypeContinuous) &&
		backupScheduleName != "" {
		return []reconcile.Request{
			{
				NamespacedName: types.NamespacedName{
					Namespace: backup.Namespace,
					Name:      backupScheduleName,
				},
			},
		}
	}
	return []reconcile.Request{}
}
