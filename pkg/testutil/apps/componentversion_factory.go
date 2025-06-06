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

package apps

import (
	appsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	"github.com/apecloud/kubeblocks/pkg/constant"
)

type MockComponentVersionFactory struct {
	BaseFactory[appsv1.ComponentVersion, *appsv1.ComponentVersion, MockComponentVersionFactory]
}

func NewComponentVersionFactory(name string) *MockComponentVersionFactory {
	f := &MockComponentVersionFactory{}
	f.Init("", name, &appsv1.ComponentVersion{
		Spec: appsv1.ComponentVersionSpec{
			CompatibilityRules: []appsv1.ComponentVersionCompatibilityRule{},
			Releases:           []appsv1.ComponentVersionRelease{},
		},
	}, f)
	f.AddAnnotations(constant.CRDAPIVersionAnnotationKey, appsv1.GroupVersion.String())
	return f
}

func (f *MockComponentVersionFactory) SetSpec(spec appsv1.ComponentVersionSpec) *MockComponentVersionFactory {
	f.Get().Spec = spec
	return f
}

func (f *MockComponentVersionFactory) SetDefaultSpec(compDef string) *MockComponentVersionFactory {
	f.Get().Spec = defaultComponentVerSpec(compDef)
	return f
}

func (f *MockComponentVersionFactory) AddRelease(name, changes, serviceVersion string, images map[string]string) *MockComponentVersionFactory {
	release := appsv1.ComponentVersionRelease{
		Name:           name,
		Changes:        changes,
		ServiceVersion: serviceVersion,
		Images:         images,
	}
	f.Get().Spec.Releases = append(f.Get().Spec.Releases, release)
	return f
}

func (f *MockComponentVersionFactory) AddCompatibilityRule(compDefs, releases []string) *MockComponentVersionFactory {
	rule := appsv1.ComponentVersionCompatibilityRule{
		CompDefs: compDefs,
		Releases: releases,
	}
	f.Get().Spec.CompatibilityRules = append(f.Get().Spec.CompatibilityRules, rule)
	return f
}
