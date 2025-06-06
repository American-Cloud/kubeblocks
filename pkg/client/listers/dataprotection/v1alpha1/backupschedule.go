/*
Copyright (C) 2022-2025 ApeCloud Co., Ltd

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/apecloud/kubeblocks/apis/dataprotection/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// BackupScheduleLister helps list BackupSchedules.
// All objects returned here must be treated as read-only.
type BackupScheduleLister interface {
	// List lists all BackupSchedules in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.BackupSchedule, err error)
	// BackupSchedules returns an object that can list and get BackupSchedules.
	BackupSchedules(namespace string) BackupScheduleNamespaceLister
	BackupScheduleListerExpansion
}

// backupScheduleLister implements the BackupScheduleLister interface.
type backupScheduleLister struct {
	indexer cache.Indexer
}

// NewBackupScheduleLister returns a new BackupScheduleLister.
func NewBackupScheduleLister(indexer cache.Indexer) BackupScheduleLister {
	return &backupScheduleLister{indexer: indexer}
}

// List lists all BackupSchedules in the indexer.
func (s *backupScheduleLister) List(selector labels.Selector) (ret []*v1alpha1.BackupSchedule, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.BackupSchedule))
	})
	return ret, err
}

// BackupSchedules returns an object that can list and get BackupSchedules.
func (s *backupScheduleLister) BackupSchedules(namespace string) BackupScheduleNamespaceLister {
	return backupScheduleNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// BackupScheduleNamespaceLister helps list and get BackupSchedules.
// All objects returned here must be treated as read-only.
type BackupScheduleNamespaceLister interface {
	// List lists all BackupSchedules in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.BackupSchedule, err error)
	// Get retrieves the BackupSchedule from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.BackupSchedule, error)
	BackupScheduleNamespaceListerExpansion
}

// backupScheduleNamespaceLister implements the BackupScheduleNamespaceLister
// interface.
type backupScheduleNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all BackupSchedules in the indexer for a given namespace.
func (s backupScheduleNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.BackupSchedule, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.BackupSchedule))
	})
	return ret, err
}

// Get retrieves the BackupSchedule from the indexer for a given namespace and name.
func (s backupScheduleNamespaceLister) Get(name string) (*v1alpha1.BackupSchedule, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("backupschedule"), name)
	}
	return obj.(*v1alpha1.BackupSchedule), nil
}
