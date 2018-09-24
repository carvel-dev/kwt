package resource

import (
	"fmt"
	"strings"

	cmdcore "github.com/cppforlife/kwt/pkg/kwt/cmd/core"
	ctlres "github.com/cppforlife/kwt/pkg/kwt/resources"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ResourceFlags struct {
	Ref ctlres.ResourceRef
}

func (s *ResourceFlags) Set(cmd *cobra.Command, flagsFactory cmdcore.FlagsFactory) {
	resRef := NewResourceRefFlag(&s.Ref)
	cmd.Flags().VarP(resRef, "resource-type", "t", "Specified resource type (format: 'group/version/resource')")
}

type ResourceRefFlag struct {
	value *ctlres.ResourceRef
}

var _ pflag.Value = &ResourceRefFlag{}

func NewResourceRefFlag(value *ctlres.ResourceRef) *ResourceRefFlag {
	return &ResourceRefFlag{value}
}

func (s *ResourceRefFlag) Set(val string) error {
	pieces := strings.Split(val, "/")
	if len(pieces) != 3 {
		return fmt.Errorf("Expected resource ref '%s' to be in format 'group/version/resource'", val)
	}

	gvr := schema.GroupVersionResource{
		Group:    pieces[0],
		Version:  pieces[1],
		Resource: pieces[2],
	}

	*s.value = ctlres.ResourceRef{gvr}

	return nil
}

func (s *ResourceRefFlag) Type() string   { return "string" }
func (s *ResourceRefFlag) String() string { return "" } // default for usage
