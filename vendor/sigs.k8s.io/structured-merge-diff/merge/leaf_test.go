/*
Copyright 2018 The Kubernetes Authors.

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

package merge_test

import (
	"testing"

	"sigs.k8s.io/structured-merge-diff/fieldpath"
	. "sigs.k8s.io/structured-merge-diff/internal/fixture"
	"sigs.k8s.io/structured-merge-diff/merge"
	"sigs.k8s.io/structured-merge-diff/typed"
)

var leafFieldsParser = func() typed.ParseableType {
	parser, err := typed.NewParser(`types:
- name: leafFields
  map:
    fields:
    - name: numeric
      type:
        scalar: numeric
    - name: string
      type:
        scalar: string
    - name: bool
      type:
        scalar: boolean`)
	if err != nil {
		panic(err)
	}
	return parser.Type("leafFields")
}()

func TestUpdateLeaf(t *testing.T) {
	tests := map[string]TestCase{
		"apply_twice": {
			Ops: []Operation{
				Apply{
					Manager: "default",
					Object: `
						numeric: 1
						string: "string"
					`,
					APIVersion: "v1",
				},
				Apply{
					Manager: "default",
					Object: `
						numeric: 2
						string: "string"
						bool: false
					`,
					APIVersion: "v1",
				},
			},
			Object: `
				numeric: 2
				string: "string"
				bool: false
			`,
			Managed: fieldpath.ManagedFields{
				"default": fieldpath.NewVersionedSet(
					_NS(
						_P("numeric"), _P("string"), _P("bool"),
					),
					"v1",
					false,
				),
			},
		},
		"apply_twice_different_versions": {
			Ops: []Operation{
				Apply{
					Manager: "default",
					Object: `
						numeric: 1
						string: "string"
					`,
					APIVersion: "v1",
				},
				Apply{
					Manager: "default",
					Object: `
						numeric: 2
						string: "string"
						bool: false
					`,
					APIVersion: "v2",
				},
			},
			Object: `
				numeric: 2
				string: "string"
				bool: false
			`,
			Managed: fieldpath.ManagedFields{
				"default": fieldpath.NewVersionedSet(
					_NS(
						_P("numeric"), _P("string"), _P("bool"),
					),
					"v2",
					false,
				),
			},
		},
		"apply_update_apply_no_conflict": {
			Ops: []Operation{
				Apply{
					Manager:    "default",
					APIVersion: "v1",
					Object: `
						numeric: 1
						string: "string"
					`,
				},
				Update{
					Manager:    "controller",
					APIVersion: "v1",
					Object: `
						numeric: 1
						string: "string"
						bool: true
					`,
				},
				Apply{
					Manager:    "default",
					APIVersion: "v1",
					Object: `
						numeric: 2
						string: "string"
					`,
				},
			},
			Object: `
				numeric: 2
				string: "string"
				bool: true
			`,
			Managed: fieldpath.ManagedFields{
				"default": fieldpath.NewVersionedSet(
					_NS(
						_P("numeric"), _P("string"),
					),
					"v1",
					false,
				),
				"controller": fieldpath.NewVersionedSet(
					_NS(
						_P("bool"),
					),
					"v1",
					false,
				),
			},
		},
		"apply_update_apply_no_conflict_different_version": {
			Ops: []Operation{
				Apply{
					Manager:    "default",
					APIVersion: "v1",
					Object: `
						numeric: 1
						string: "string"
					`,
				},
				Update{
					Manager:    "controller",
					APIVersion: "v2",
					Object: `
						numeric: 1
						string: "string"
						bool: true
					`,
				},
				Apply{
					Manager:    "default",
					APIVersion: "v1",
					Object: `
						numeric: 2
						string: "string"
					`,
				},
			},
			Object: `
				numeric: 2
				string: "string"
				bool: true
			`,
			Managed: fieldpath.ManagedFields{
				"default": fieldpath.NewVersionedSet(
					_NS(
						_P("numeric"), _P("string"),
					),
					"v1",
					false,
				),
				"controller": fieldpath.NewVersionedSet(
					_NS(
						_P("bool"),
					),
					"v2",
					false,
				),
			},
		},
		"apply_update_apply_with_conflict": {
			Ops: []Operation{
				Apply{
					Manager:    "default",
					APIVersion: "v1",
					Object: `
						numeric: 1
						string: "string"
					`,
				},
				Update{
					Manager:    "controller",
					APIVersion: "v1",
					Object: `
						numeric: 1
						string: "controller string"
						bool: true
					`,
				},
				Apply{
					Manager:    "default",
					APIVersion: "v1",
					Object: `
						numeric: 2
						string: "user string"
					`,
					Conflicts: merge.Conflicts{
						merge.Conflict{Manager: "controller", Path: _P("string")},
					},
				},
				ForceApply{
					Manager:    "default",
					APIVersion: "v1",
					Object: `
						numeric: 2
						string: "user string"
					`,
				},
			},
			Object: `
				numeric: 2
				string: "user string"
				bool: true
			`,
			Managed: fieldpath.ManagedFields{
				"default": fieldpath.NewVersionedSet(
					_NS(
						_P("numeric"), _P("string"),
					),
					"v1",
					false,
				),
				"controller": fieldpath.NewVersionedSet(
					_NS(
						_P("bool"),
					),
					"v1",
					false,
				),
			},
		},
		"apply_update_apply_with_conflict_across_version": {
			Ops: []Operation{
				Apply{
					Manager:    "default",
					APIVersion: "v1",
					Object: `
						numeric: 1
						string: "string"
					`,
				},
				Update{
					Manager:    "controller",
					APIVersion: "v2",
					Object: `
						numeric: 1
						string: "controller string"
						bool: true
					`,
				},
				Apply{
					Manager:    "default",
					APIVersion: "v1",
					Object: `
						numeric: 2
						string: "user string"
					`,
					Conflicts: merge.Conflicts{
						merge.Conflict{Manager: "controller", Path: _P("string")},
					},
				},
				ForceApply{
					Manager:    "default",
					APIVersion: "v1",
					Object: `
						numeric: 2
						string: "user string"
					`,
				},
			},
			Object: `
				numeric: 2
				string: "user string"
				bool: true
			`,
			Managed: fieldpath.ManagedFields{
				"default": fieldpath.NewVersionedSet(
					_NS(
						_P("numeric"), _P("string"),
					),
					"v1",
					false,
				),
				"controller": fieldpath.NewVersionedSet(
					_NS(
						_P("bool"),
					),
					"v2",
					false,
				),
			},
		},
		"apply_twice_dangling": {
			Ops: []Operation{
				Apply{
					Manager:    "default",
					APIVersion: "v1",
					Object: `
						numeric: 1
						string: "string"
						bool: false
					`,
				},
				Apply{
					Manager:    "default",
					APIVersion: "v1",
					Object: `
						string: "new string"
					`,
				},
			},
			Object: `
				numeric: 1
				string: "new string"
				bool: false
			`,
			Managed: fieldpath.ManagedFields{
				"default": fieldpath.NewVersionedSet(
					_NS(
						_P("string"),
					),
					"v1",
					false,
				),
			},
		},
		"apply_twice_dangling_different_version": {
			Ops: []Operation{
				Apply{
					Manager:    "default",
					APIVersion: "v1",
					Object: `
						numeric: 1
						string: "string"
						bool: false
					`,
				},
				Apply{
					Manager:    "default",
					APIVersion: "v2",
					Object: `
						string: "new string"
					`,
				},
			},
			Object: `
				numeric: 1
				string: "new string"
				bool: false
			`,
			Managed: fieldpath.ManagedFields{
				"default": fieldpath.NewVersionedSet(
					_NS(
						_P("string"),
					),
					"v2",
					false,
				),
			},
		},
		"update_remove_empty_set": {
			Ops: []Operation{
				Apply{
					Manager:    "default",
					APIVersion: "v1",
					Object: `
						string: "string"
					`,
				},
				Update{
					Manager:    "controller",
					APIVersion: "v1",
					Object: `
						string: "new string"
					`,
				},
			},
			Object: `
				string: "new string"
			`,
			Managed: fieldpath.ManagedFields{
				"controller": fieldpath.NewVersionedSet(
					_NS(
						_P("string"),
					),
					"v1",
					false,
				),
			},
		},
		"apply_remove_empty_set": {
			Ops: []Operation{
				Apply{
					Manager:    "default",
					APIVersion: "v1",
					Object: `
						string: "string"
					`,
				},
				Apply{
					Manager:    "default",
					APIVersion: "v1",
					Object:     "",
				},
			},
			Object: `
				string: "string"
			`,
			Managed: fieldpath.ManagedFields{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := test.Test(leafFieldsParser); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func BenchmarkLeafConflictAcrossVersion(b *testing.B) {
	test := TestCase{
		Ops: []Operation{
			Apply{
				Manager:    "default",
				APIVersion: "v1",
				Object: `
					numeric: 1
					string: "string"
				`,
			},
			Update{
				Manager:    "controller",
				APIVersion: "v2",
				Object: `
					numeric: 1
					string: "controller string"
					bool: true
				`,
			},
			Apply{
				Manager:    "default",
				APIVersion: "v1",
				Object: `
					numeric: 2
					string: "user string"
				`,
				Conflicts: merge.Conflicts{
					merge.Conflict{Manager: "controller", Path: _P("string")},
				},
			},
			ForceApply{
				Manager:    "default",
				APIVersion: "v1",
				Object: `
					numeric: 2
					string: "user string"
				`,
			},
		},
		Object: `
			numeric: 2
			string: "user string"
			bool: true
		`,
		Managed: fieldpath.ManagedFields{
			"default": fieldpath.NewVersionedSet(
				_NS(
					_P("numeric"), _P("string"),
				),
				"v1",
				false,
			),
			"controller": fieldpath.NewVersionedSet(
				_NS(
					_P("bool"),
				),
				"v2",
				false,
			),
		},
	}

	// Make sure this passes...
	if err := test.Test(leafFieldsParser); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := test.Bench(leafFieldsParser); err != nil {
			b.Fatal(err)
		}
	}
}
