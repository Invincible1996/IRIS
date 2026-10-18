package sys

import "github.com/versenilvis/iris/spec"

func init() {
	spec.Register(&spec.Spec{
		Name:        "zip",
		Description: "package and compress files",
		Generator:   spec.FileGenerator(),
		Options: []spec.Option{
			{Name: "-r", Description: "recurse into directories"},
			{Name: "-q", Description: "quiet mode"},
			{Name: "-v", Description: "verbose mode"},
			{Name: "-u", Description: "update existing entries"},
			{Name: "-f", Description: "freshen existing entries"},
			{Name: "-d", Description: "delete entries from archive"},
			{Name: "-T", Description: "test archive integrity"},
			{Name: "-j", Description: "junk directory names"},
			{Name: "-x", Description: "exclude entries"},
			{Name: "-i", Description: "include only entries"},
		},
	})
}
