// Copyright 2018 Yipee.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	//	"fmt"
)

// Tracks ownership relationships between Kubernetes objects. If an object
// has no owner, we treat it as its own owner
type ownerRef struct {
	ctx         context.Context
	ref         map[string]interface{}
	cachedOwner resource // cached info for on-demand lookup
}

// resource method implementations
func (r *ownerRef) Kind() string {
	if r.cachedOwner == nil {
		r.cachedOwner = getOwner(r.ctx, r.ref)
	}
	return r.cachedOwner.Kind()
}

func (r *ownerRef) Metadata() *metadataResolver {
	if r.cachedOwner == nil {
		r.cachedOwner = getOwner(r.ctx, r.ref)
	}
	return r.cachedOwner.Metadata()
}

func (r *ownerRef) Owner() *resourceResolver {
	if r.cachedOwner == nil {
		r.cachedOwner = getOwner(r.ctx, r.ref)
	}
	return r.cachedOwner.Owner()
}

func (r *ownerRef) RootOwner() *resourceResolver {
	if r.cachedOwner == nil {
		r.cachedOwner = getOwner(r.ctx, r.ref)
	}
	return r.cachedOwner.RootOwner()
}

// Fetch owners by getting ownerReferences and doing lookups based
// on their contents
func getRawOwner(val map[string]interface{}) map[string]interface{} {
	if orefs := getMetadataField(val, "ownerReferences"); orefs != nil {
		oArray := orefs.([]interface{})
		if len(oArray) > 0 {
			owner := oArray[0].(map[string]interface{})
			if res := getK8sResource(owner["kind"].(string),
				getNamespace(val),
				owner["name"].(string)); res != nil {
				return res
			}
		}
	}

	return val
}

func getOwner(ctx context.Context, val map[string]interface{}) resource {
	return mapToResource(ctx, getRawOwner(val))
}

func getRootOwner(ctx context.Context, val map[string]interface{}) resource {
	result := getRawOwner(val)

	if getUid(result) == getUid(val) {
		return mapToResource(ctx, result)
	}

	return getRootOwner(ctx, getRawOwner(result))
}

func hasMatchingOwner(jsonObj map[string]interface{}, name, kind string) bool {
	if orefs := getMetadataField(jsonObj, "ownerReferences"); orefs != nil {
		oArray := orefs.([]interface{})
		for _, oref := range oArray {
			owner := oref.(map[string]interface{})
			if owner["name"] == name && owner["kind"] == kind {
				return true
			}
		}
	}

	return false
}