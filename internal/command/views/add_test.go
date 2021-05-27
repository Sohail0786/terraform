package views

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/internal/configs/configschema"
	"github.com/zclconf/go-cty/cty"
)

func TestAdd_WriteConfigBlocksFromExisting(t *testing.T) {

	t.Run("NestingSingle", func(t *testing.T) {
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
			t.Errorf("wrong output:\n%s", cmp.Diff(expected, buf.String()))
		}
	})

	t.Run("NestingList", func(t *testing.T) {
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
			t.Fatalf("wrong output:\n%s", cmp.Diff(expected, buf.String()))
		}
	})

	t.Run("NestingMap", func(t *testing.T) {
		v := addHuman{optional: true}
		val := cty.ObjectVal(map[string]cty.Value{
			"root_block_device": cty.MapVal(map[string]cty.Value{
				"1": cty.ObjectVal(map[string]cty.Value{
					"volume_type": cty.StringVal("foo"),
				}),
				"2": cty.ObjectVal(map[string]cty.Value{
					"volume_type": cty.StringVal("bar"),
				}),
			}),
		})
		schema := addTestSchema(configschema.NestingMap)
		var buf strings.Builder
		v.writeConfigBlocksFromExisting(&buf, val, schema.BlockTypes, 0)

		expected := `root_block_device "1" {
  volume_type = "foo"
}
root_block_device "2" {
  volume_type = "bar"
}
`

		if !cmp.Equal(buf.String(), expected) {
			t.Fatalf("wrong output:\n%s", cmp.Diff(expected, buf.String()))
		}
	})
}

func TestAdd_WriteConfigNestedTypeAttributeFromExisting(t *testing.T) {
	t.Run("NestingSingle", func(t *testing.T) {
		v := addHuman{optional: true}
		val := cty.ObjectVal(map[string]cty.Value{
			"disks": cty.ObjectVal(map[string]cty.Value{
				"mount_point": cty.StringVal("/mnt/foo"),
				"size":        cty.StringVal("50GB"),
			}),
		})
		schema := addTestSchema(configschema.NestingSingle)
		var buf strings.Builder
		v.writeConfigNestedTypeAttributeFromExisting(&buf, "disks", schema.Attributes["disks"], val, 0)

		expected := `disks = {
  mount_point = "/mnt/foo"
  size = "50GB"
}
`

		if !cmp.Equal(buf.String(), expected) {
			t.Fatalf("wrong output:\n%s", cmp.Diff(expected, buf.String()))
		}
	})

	t.Run("NestingSingle_sensitive", func(t *testing.T) {
		v := addHuman{optional: true}
		val := cty.ObjectVal(map[string]cty.Value{
			"disks": cty.ObjectVal(map[string]cty.Value{
				"mount_point": cty.StringVal("/mnt/foo"),
				"size":        cty.StringVal("50GB"),
			}),
		})
		schema := addTestSchemaSensitive(configschema.NestingSingle)
		var buf strings.Builder
		v.writeConfigNestedTypeAttributeFromExisting(&buf, "disks", schema.Attributes["disks"], val, 0)

		expected := `disks = { (sensitive) }
`

		if !cmp.Equal(buf.String(), expected) {
			t.Fatalf("wrong output:\n%s", cmp.Diff(expected, buf.String()))
		}
	})

	t.Run("NestingList", func(t *testing.T) {
		v := addHuman{optional: true}
		val := cty.ObjectVal(map[string]cty.Value{
			"disks": cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"mount_point": cty.StringVal("/mnt/foo"),
					"size":        cty.StringVal("50GB"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"mount_point": cty.StringVal("/mnt/bar"),
					"size":        cty.StringVal("250GB"),
				}),
			}),
		})

		schema := addTestSchema(configschema.NestingList)
		var buf strings.Builder
		v.writeConfigNestedTypeAttributeFromExisting(&buf, "disks", schema.Attributes["disks"], val, 0)

		expected := `disks = [
  {
    mount_point = "/mnt/foo"
    size = "50GB"
  },
  {
    mount_point = "/mnt/bar"
    size = "250GB"
  },
]
`

		if !cmp.Equal(buf.String(), expected) {
			fmt.Println(buf.String())
			t.Fatalf("wrong output:\n%s", cmp.Diff(expected, buf.String()))
		}
	})

	t.Run("NestingMap", func(t *testing.T) {
		v := addHuman{optional: true}
		val := cty.ObjectVal(map[string]cty.Value{
			"disks": cty.MapVal(map[string]cty.Value{
				"foo": cty.ObjectVal(map[string]cty.Value{
					"mount_point": cty.StringVal("/mnt/foo"),
					"size":        cty.StringVal("50GB"),
				}),
				"bar": cty.ObjectVal(map[string]cty.Value{
					"mount_point": cty.StringVal("/mnt/bar"),
					"size":        cty.StringVal("250GB"),
				}),
			}),
		})
		schema := addTestSchema(configschema.NestingMap)
		var buf strings.Builder
		v.writeConfigNestedTypeAttributeFromExisting(&buf, "disks", schema.Attributes["disks"], val, 0)

		expected := `disks = {
  bar = {
    mount_point = "/mnt/bar"
    size = "250GB"
  },
  foo = {
    mount_point = "/mnt/foo"
    size = "50GB"
  },
}
`

		if !cmp.Equal(buf.String(), expected) {
			fmt.Println(buf.String())
			t.Fatalf("wrong output:\n%s", cmp.Diff(expected, buf.String()))
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

// addTestSchemaSensitive returns a schema with a sensitive NestedType and a
// NestedBlock with sensitive attributes.
func addTestSchemaSensitive(nesting configschema.NestingMode) *configschema.Block {
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
				Sensitive: true,
			},
		},
		BlockTypes: map[string]*configschema.NestedBlock{
			"root_block_device": {
				Block: configschema.Block{
					Attributes: map[string]*configschema.Attribute{
						"volume_type": {
							Type:      cty.String,
							Optional:  true,
							Computed:  true,
							Sensitive: true,
						},
					},
				},
				Nesting: nesting,
			},
		},
	}
}
