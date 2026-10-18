package sys

import "github.com/versenilvis/iris/spec"

func init() {
	spec.Register(&spec.Spec{
		Name:        "xattr",
		Description: "display and manipulate extended attributes",
		Generator:   spec.FileGenerator(),
		Options: []spec.Option{
			{Name: "-c", Description: "clear all extended attributes"},
			{Name: "-d", Description: "delete an extended attribute"},
			{Name: "-l", Description: "list extended attributes"},
			{Name: "-p", Description: "print an extended attribute"},
			{Name: "-r", Description: "operate recursively"},
			{Name: "-w", Description: "write an extended attribute"},
			{Name: "-x", Description: "display attribute values in hexadecimal"},
		},
	})
}
