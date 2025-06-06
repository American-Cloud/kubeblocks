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

package model

import "sigs.k8s.io/controller-runtime/pkg/client"

type GraphOptions struct {
	replaceIfExisting     bool
	haveDifferentTypeWith bool
	clientOpt             any
	propagationPolicy     client.PropagationPolicy
	subResource           string
}

type GraphOption interface {
	ApplyTo(*GraphOptions)
}

// ReplaceIfExistingOption tells the GraphWriter methods to replace Obj and OriObj with the given ones if already existing.
// used in Action methods: Create, Update, Patch, Status and Delete
type ReplaceIfExistingOption struct{}

var _ GraphOption = &ReplaceIfExistingOption{}

func (o *ReplaceIfExistingOption) ApplyTo(opts *GraphOptions) {
	opts.replaceIfExisting = true
}

// HaveDifferentTypeWithOption is used in FindAll method to find all objects have different type with the given one.
type HaveDifferentTypeWithOption struct{}

var _ GraphOption = &HaveDifferentTypeWithOption{}

func (o *HaveDifferentTypeWithOption) ApplyTo(opts *GraphOptions) {
	opts.haveDifferentTypeWith = true
}

type clientOption struct {
	opt any
}

var _ GraphOption = &clientOption{}

func (o *clientOption) ApplyTo(opts *GraphOptions) {
	opts.clientOpt = o.opt
}

func WithClientOption(opt any) GraphOption {
	return &clientOption{
		opt: opt,
	}
}

type propagationPolicyOption struct {
	propagationPolicy client.PropagationPolicy
}

func (o *propagationPolicyOption) ApplyTo(opts *GraphOptions) {
	opts.propagationPolicy = o.propagationPolicy
}

var _ GraphOption = &propagationPolicyOption{}

func WithPropagationPolicy(policy client.PropagationPolicy) GraphOption {
	return &propagationPolicyOption{
		propagationPolicy: policy,
	}
}

type subResourceOption struct {
	subResource string
}

func (o *subResourceOption) ApplyTo(opts *GraphOptions) {
	opts.subResource = o.subResource
}

var _ GraphOption = &subResourceOption{}

func WithSubResource(subResource string) GraphOption {
	return &subResourceOption{
		subResource: subResource,
	}
}
