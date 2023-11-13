package querygen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestK8sAuditFieldSet_QueryGen(t *testing.T) {
	kfs := K8sAuditFieldSet{
		K8sApiId{}: {
			"plain_field":     {},
			"override_field":  {},
			"override_field2": {},
		},
		K8sApiId{"api1.com", "SomeObj"}: {
			"override_field": {srcFields: []string{"other_field"}},
		},
	}
	fields := []string{"plain_field", "override_field", "override_field2"}
	includeFieldsCmd := `fields override_field,override_field2,plain_field`
	type args struct {
		index      string
		api        K8sApiId
		searchExpr string
		extra      []FieldSet
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Simple query",
			args: args{
				index:      "some_idx",
				api:        K8sApiId{"other.api.com", "SomeOtherObj"},
				searchExpr: "foo bar baz",
			},
			want: `search index="some_idx" log_type=audit ` +
				`"objectRef.apiGroup"="other.api.com" ` +
				`"objectRef.resource"="SomeOtherObj" ` +
				`foo bar baz` +
				`|` + includeFieldsCmd +
				`|` + excludeFieldsCmd,
		},
		{
			name: "Query on customized object",
			args: args{
				index:      "some_idx",
				api:        K8sApiId{"api1.com", "SomeObj"},
				searchExpr: "foo bar baz",
			},
			want: `search index="some_idx" log_type=audit ` +
				`"objectRef.apiGroup"="api1.com" ` +
				`"objectRef.resource"="SomeObj" ` +
				`foo bar baz` +
				`|eval override_field='other_field'` +
				`|` + includeFieldsCmd +
				`|` + excludeFieldsCmd,
		},
		{
			name: "Query with extra fields",
			args: args{
				index:      "some_idx",
				api:        K8sApiId{"api1.com", "SomeObj"},
				searchExpr: "foo bar baz",
				extra: []FieldSet{
					{
						"override_field2": {srcExpr: "foo()"},
						"added_field":     {},
					},
				},
			},
			want: `search index="some_idx" log_type=audit ` +
				`"objectRef.apiGroup"="api1.com" ` +
				`"objectRef.resource"="SomeObj" ` +
				`foo bar baz` +
				`|eval override_field='other_field',override_field2=foo()` +
				`|fields added_field,override_field,override_field2,plain_field` +
				`|` + excludeFieldsCmd,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := kfs.QueryGen(tt.args.index, tt.args.api, tt.args.searchExpr, fields, tt.args.extra...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
