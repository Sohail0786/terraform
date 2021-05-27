package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform/internal/addrs"
	"github.com/hashicorp/terraform/internal/command/arguments"
	"github.com/hashicorp/terraform/internal/configs/configschema"
	"github.com/hashicorp/terraform/internal/tfdiags"
	"github.com/zclconf/go-cty/cty"
)

// Add is the view interface for the "terraform add" command.
type Add interface {
	Resource(addrs.AbsResourceInstance, *configschema.Block, string, cty.Value) error
	Diagnostics(tfdiags.Diagnostics)
}

// NewAdd returns an initialized Validate implementation for the given ViewType.
func NewAdd(vt arguments.ViewType, view *View, args *arguments.Add) Add {
	return &addHuman{
		view:     view,
		optional: args.Optional,
		outPath:  args.OutPath,
	}
}

type addHuman struct {
	view     *View
	optional bool
	outPath  string
}

func (v *addHuman) Resource(addr addrs.AbsResourceInstance, schema *configschema.Block, provider string, stateVal cty.Value) error {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("resource %q %q {\n", addr.Resource.Resource.Type, addr.Resource.Resource.Name))
	if provider != "" {
		buf.WriteString(strings.Repeat(" ", 2))
		buf.WriteString(fmt.Sprintf("provider = %s\n", provider))
	}

	if stateVal.RawEquals(cty.NilVal) {
		err := v.writeConfigAttributes(&buf, schema.Attributes, 2)
		if err != nil {
			return err
		}
		err = v.writeConfigBlocks(&buf, schema.BlockTypes, 2)
		if err != nil {
			return err
		}
	} else {
		err := v.writeConfigAttributesFromExisting(&buf, stateVal, schema.Attributes, 2)
		if err != nil {
			return err
		}
		err = v.writeConfigBlocksFromExisting(&buf, stateVal, schema.BlockTypes, 2)
		if err != nil {
			return err
		}
	}

	buf.WriteString("}")

	// The output better be valid HCL which can be parsed and formatted.
	formatted := hclwrite.Format([]byte(buf.String()))
	_, err := v.view.streams.Println(string(formatted))

	return err
}

func (v *addHuman) Diagnostics(diags tfdiags.Diagnostics) {
	v.view.Diagnostics(diags)
}

func (v *addHuman) writeConfigAttributes(buf *strings.Builder, attrs map[string]*configschema.Attribute, indent int) error {
	if len(attrs) == 0 {
		return nil
	}

	// Get a list of sorted attribute names so the output will be consistent between runs.
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i := range keys {
		name := keys[i]
		attrS := attrs[name]
		if attrS.NestedType != nil {
			err := v.writeConfigNestedTypeAttribute(buf, name, attrS, indent)
			if err != nil {
				return err
			}
			continue
		}
		if attrS.Required {
			buf.WriteString(strings.Repeat(" ", indent))
			buf.WriteString(fmt.Sprintf("%s = ", name))

			if attrS.Sensitive {
				buf.WriteString("(sensitive)")
			} else {
				tok := hclwrite.TokensForValue(attrS.EmptyValue())
				_, err := tok.WriteTo(buf)
				if err != nil {
					return err
				}
			}
			writeAttrTypeConstraint(buf, attrS)

		} else if attrS.Optional && v.optional {
			buf.WriteString(strings.Repeat(" ", indent))
			buf.WriteString(fmt.Sprintf("%s = ", name))

			if attrS.Sensitive {
				buf.WriteString("(sensitive)")
			} else {
				tok := hclwrite.TokensForValue(attrS.EmptyValue())
				_, err := tok.WriteTo(buf)
				if err != nil {
					return err
				}
			}
			writeAttrTypeConstraint(buf, attrS)
		}
	}
	return nil
}

func (v *addHuman) writeConfigAttributesFromExisting(buf *strings.Builder, stateVal cty.Value, attrs map[string]*configschema.Attribute, indent int) error {
	if len(attrs) == 0 {
		return nil
	}

	// Get a list of sorted attribute names so the output will be consistent between runs.
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i := range keys {
		name := keys[i]
		attrS := attrs[name]
		if attrS.NestedType != nil {
			err := v.writeConfigNestedTypeAttributeFromExisting(buf, name, attrS, stateVal, indent)
			if err != nil {
				return err
			}
			continue
		}
		if attrS.Required {
			buf.WriteString(strings.Repeat(" ", indent))
			buf.WriteString(fmt.Sprintf("%s = ", name))

			if attrS.Sensitive {
				buf.WriteString("(sensitive)")
			} else {
				var val cty.Value
				if stateVal.Type().HasAttribute(name) {
					val = stateVal.GetAttr(name)
				} else {
					val = attrS.EmptyValue()
				}
				tok := hclwrite.TokensForValue(val)
				_, err := tok.WriteTo(buf)
				if err != nil {
					return err
				}
			}

		} else if attrS.Optional && v.optional {
			buf.WriteString(strings.Repeat(" ", indent))
			buf.WriteString(fmt.Sprintf("%s = ", name))

			if attrS.Sensitive {
				buf.WriteString("(sensitive)")
			} else {
				var val cty.Value
				if !stateVal.RawEquals(cty.NilVal) && stateVal.Type().HasAttribute(name) {
					val = stateVal.GetAttr(name)
				} else {
					val = attrS.EmptyValue()
				}
				tok := hclwrite.TokensForValue(val)
				_, err := tok.WriteTo(buf)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (v *addHuman) writeConfigBlocks(buf *strings.Builder, blocks map[string]*configschema.NestedBlock, indent int) error {
	if len(blocks) == 0 {
		return nil
	}

	// Get a list of sorted block names so the output will be consistent between runs.
	names := make([]string, 0, len(blocks))
	for k := range blocks {
		names = append(names, k)
	}
	sort.Strings(names)

	for i := range names {
		name := names[i]
		blockS := blocks[name]

		if blockS.MinItems > 0 {
			buf.WriteString(strings.Repeat(" ", indent))
			buf.WriteString(fmt.Sprintf("%s {\n", name))
			if len(blockS.Attributes) > 0 {
				v.writeConfigAttributes(buf, blockS.Attributes, indent+2)
			}
			if len(blockS.BlockTypes) > 0 {
				v.writeConfigBlocks(buf, blockS.BlockTypes, indent+2)
			}
			buf.WriteString(strings.Repeat(" ", indent))
			buf.WriteString("}\n")
		}
	}
	return nil
}

func (v *addHuman) writeConfigNestedTypeAttribute(buf *strings.Builder, name string, schema *configschema.Attribute, indent int) error {
	if schema.Required == false && v.optional == false {
		return nil
	}

	buf.WriteString(strings.Repeat(" ", indent))
	buf.WriteString(fmt.Sprintf("%s = ", name))

	switch schema.NestedType.Nesting {
	case configschema.NestingSingle:
		buf.WriteString("{")
		writeAttrTypeConstraint(buf, schema)
		v.writeConfigAttributes(buf, schema.NestedType.Attributes, indent+2)
		buf.WriteString(strings.Repeat(" ", indent))
		buf.WriteString("}\n")
	case configschema.NestingList, configschema.NestingSet:
		buf.WriteString("[{")
		writeAttrTypeConstraint(buf, schema)
		v.writeConfigAttributes(buf, schema.NestedType.Attributes, indent+2)
		buf.WriteString(strings.Repeat(" ", indent))
		buf.WriteString("}]\n")
	case configschema.NestingMap:
		buf.WriteString("{")
		writeAttrTypeConstraint(buf, schema)
		buf.WriteString(strings.Repeat(" ", indent+2))
		buf.WriteString("key = {\n")
		v.writeConfigAttributes(buf, schema.NestedType.Attributes, indent+4)
		buf.WriteString(strings.Repeat(" ", indent+2))
		buf.WriteString("}\n")
		buf.WriteString(strings.Repeat(" ", indent))
		buf.WriteString("}\n")
	}

	return nil
}

func (v *addHuman) writeConfigBlocksFromExisting(buf *strings.Builder, stateVal cty.Value, blocks map[string]*configschema.NestedBlock, indent int) error {
	if len(blocks) == 0 {
		return nil
	}

	// Get a list of sorted block names so the output will be consistent between runs.
	names := make([]string, 0, len(blocks))
	for k := range blocks {
		names = append(names, k)
	}
	sort.Strings(names)

	for _, name := range names {
		blockS := blocks[name]
		blockVal := stateVal.GetAttr(name)
		v.writeConfigNestedBlockFromExisting(buf, name, blockS, blockVal, indent)
	}

	return nil
}

func (v *addHuman) writeConfigNestedTypeAttributeFromExisting(buf *strings.Builder, name string, schema *configschema.Attribute, stateVal cty.Value, indent int) error {
	return nil
}

func (v *addHuman) writeConfigNestedBlockFromExisting(buf *strings.Builder, name string, schema *configschema.NestedBlock, stateVal cty.Value, indent int) error {
	switch schema.Nesting {
	case configschema.NestingSingle, configschema.NestingGroup:
		buf.WriteString(fmt.Sprintf("%s {\n", name))
		v.writeConfigAttributesFromExisting(buf, stateVal, schema.Attributes, indent+2)
		v.writeConfigBlocksFromExisting(buf, stateVal, schema.BlockTypes, indent+2)
		buf.WriteString("\n}\n")
		return nil
	case configschema.NestingList:
		listVals := stateVal.AsValueSlice()
		for i := range listVals {
			buf.WriteString(fmt.Sprintf("%s {\n", name))
			v.writeConfigAttributesFromExisting(buf, listVals[i], schema.Attributes, indent+2)
			v.writeConfigBlocksFromExisting(buf, listVals[i], schema.BlockTypes, indent+2)
			buf.WriteString("\n}\n")
		}
		return nil
	}

	return nil

}

func writeAttrTypeConstraint(buf *strings.Builder, schema *configschema.Attribute) {
	if schema.Required {
		buf.WriteString(" # REQUIRED ")
	} else {
		buf.WriteString(" # OPTIONAL ")
	}

	if schema.NestedType != nil {
		buf.WriteString(fmt.Sprintf("%s\n", schema.NestedType.ImpliedType().FriendlyName()))
	} else {
		buf.WriteString(fmt.Sprintf("%s\n", schema.Type.FriendlyName()))
	}
	return
}
