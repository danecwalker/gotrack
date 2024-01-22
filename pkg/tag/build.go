package tag

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/evanw/esbuild/pkg/api"
)

func buildJS() ([]byte, error) {
	templateFile := "pkg/tag/tag.ts"
	t := template.Must(template.New("tag.ts").ParseFiles(templateFile))
	var buffer strings.Builder
	err := t.Execute(&buffer, map[string]interface{}{
		"error":         "error 1",
		"another_error": "another",
	})
	if err != nil {
		return nil, err
	}

	result := api.Transform(buffer.String(), api.TransformOptions{
		Loader:            api.LoaderTS,
		MinifyWhitespace:  true,
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
		Supported: map[string]bool{
			"arrow":         false,
			"destructuring": true,
		},
		Platform: api.PlatformBrowser,
		Banner:   "/* gotrack v0.0.1 */",
		TsconfigRaw: `{
			"compilerOptions": {
				"target": "es5",
				"lib": ["dom", "es2015"],
				"module": "none",
				"strict": true,
				"esModuleInterop": true,
				"skipLibCheck": true
			},
		}`,
	})

	if len(result.Errors) > 0 {
		var errs []string
		for _, err := range result.Errors {
			errs = append(errs, fmt.Sprintf("%s:%d:%d: %s", err.Location.File, err.Location.Line, err.Location.Column, err.Text))
		}
		return nil, fmt.Errorf(strings.Join(errs, "\n"))
	}

	return result.Code, nil
}
