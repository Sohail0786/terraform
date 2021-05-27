package views

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/internal/configs/configschema"
	"github.com/zclconf/go-cty/cty"
)

func TestAdd_WriteConfigBlocksFromExisting(t *testing.T) {

	t.Run("NestingMode single", func(t *testing.T) {
		v := addHuman{optional: true}
		val := cty.ObjectVal(map[string]cty.Value{
			"root_block_device": cty.ObjectVal(map[string]cty.Value{
				"volume_type": cty.StringVal("foo"),
			}),
		})
		schema := addTestSchema(configschema.NestingSingle)
		var buf strings.Builder
		v.writeConfigBlocksFromExisting(&buf, val, schema.BlockTypes, 0)

		expected := `root_block_device {
  volume_type = "foo"
}
`

		if !cmp.Equal(buf.String(), expected) {
			t.Fatalf("wrong output:\n%s", cmp.Diff(buf.String(), expected))
		}
	})

	t.Run("NestingMode list", func(t *testing.T) {
		v := addHuman{optional: true}
		val := cty.ObjectVal(map[string]cty.Value{
			"root_block_device": cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"volume_type": cty.StringVal("foo"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"volume_type": cty.StringVal("bar"),
				}),
			}),
		})
		schema := addTestSchema(configschema.NestingList)
		var buf strings.Builder
		v.writeConfigBlocksFromExisting(&buf, val, schema.BlockTypes, 0)

		expected := `root_block_device {
  volume_type = "foo"
}
root_block_device {
  volume_type = "bar"
}
`

		if !cmp.Equal(buf.String(), expected) {
			t.Fatalf("wrong output:\n%s", cmp.Diff(buf.String(), expected))
		}

	})
}

func addTestSchema(nesting configschema.NestingMode) *configschema.Block {
	return &configschema.Block{
		Attributes: map[string]*configschema.Attribute{
			"id":  {Type: cty.String, Optional: true, Computed: true},
			"ami": {Type: cty.String, Optional: true},
			"disks": {
				NestedType: &configschema.Object{
					Attributes: map[string]*configschema.Attribute{
						"mount_point": {Type: cty.String, Optional: true},
						"size":        {Type: cty.String, Optional: true},
					},
					Nesting: nesting,
				},
			},
		},
		BlockTypes: map[string]*configschema.NestedBlock{
			"root_block_device": {
				Block: configschema.Block{
					Attributes: map[string]*configschema.Attribute{
						"volume_type": {
							Type:     cty.String,
							Optional: true,
							Computed: true,
						},
					},
				},
				Nesting: nesting,
			},
		},
	}
}
