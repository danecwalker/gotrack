package main

import (
	"fmt"
	"os"

	"github.com/evanw/esbuild/pkg/api"
)

func main() {
	result := api.Build(api.BuildOptions{
		EntryPoints:       []string{"pkg/tag/tag.ts"},
		Bundle:            true,
		MinifyWhitespace:  false,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		TreeShaking:       api.TreeShakingTrue,
		Target:            api.ES2016,
		Engines: []api.Engine{
			{Name: api.EngineChrome, Version: "58"},
			{Name: api.EngineEdge, Version: "16"},
			{Name: api.EngineFirefox, Version: "57"},
			{Name: api.EngineNode, Version: "12"},
			{Name: api.EngineSafari, Version: "11"},
		},
		Write: true,
		Supported: map[string]bool{
			"arrow": false,
		},
		Outdir: "pkg/tag/dist",
	})

	if len(result.Errors) > 0 {
		fmt.Println("Build failed.")
		for _, err := range result.Errors {
			fmt.Println(err.Text)
		}
		os.Exit(1)
	}
}
