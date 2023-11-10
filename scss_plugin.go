package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bep/godartsass/v2"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/evanw/esbuild/pkg/cli"
)

func compileSass(inputPath, outputPath string) error {
	// Read the input Sass/SCSS file
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	extension := filepath.Ext(inputPath)
	var sourceSyntax godartsass.SourceSyntax
	if extension == ".scss" {
		sourceSyntax = godartsass.SourceSyntaxSCSS
	} else {
		sourceSyntax = godartsass.SourceSyntaxSASS
	}

	// Create a Dart Sass compiler
	compiler, err := godartsass.Start(godartsass.Options{})
	if err != nil {
		return err
	}
	defer compiler.Close()

	// Compile the Sass/SCSS to CSS
	output, err := compiler.Execute(godartsass.Args{
		Source:          string(input),
		OutputStyle:     godartsass.OutputStyleCompressed,
		SourceSyntax:    sourceSyntax,
		EnableSourceMap: true,
	})
	if err != nil {
		return err
	}

	// Write the compiled CSS to the output file
	err = os.WriteFile(outputPath, []byte(output.CSS), 0644)
	if err != nil {
		return err
	}

	return nil
}

var scssPlugin = api.Plugin{
	Name: "sass-loader",
	Setup: func(build api.PluginBuild) {
		build.OnResolve(api.OnResolveOptions{Filter: `^.*(scss|sass)$`},
			func(args api.OnResolveArgs) (api.OnResolveResult, error) {
				return api.OnResolveResult{
					Path:      args.Path,
					Namespace: "scss",
				}, nil
			})

		build.OnLoad(api.OnLoadOptions{Filter: `.*`, Namespace: "scss"},
			func(args api.OnLoadArgs) (api.OnLoadResult, error) {
				// Compile the Sass/SCSS file to CSS
				extension := filepath.Ext(args.Path)
				filenameWithoutExtension := strings.TrimSuffix(args.Path, extension)
				outputPath := filenameWithoutExtension + ".css"
				err := compileSass(args.Path, outputPath)
				if err != nil {
					return api.OnLoadResult{}, err
				}

				// Modify the import path to the generated CSS file
				args.Path = outputPath

				// Record the generated CSS as a dependency
				build.OnResolve(api.OnResolveOptions{
					Namespace: "file",
					Filter:    outputPath,
				}, func(args api.OnResolveArgs) (api.OnResolveResult, error) {
					return api.OnResolveResult{Path: outputPath, External: true}, nil
				})

				result := ""
				return api.OnLoadResult{Contents: &result}, nil
			})
	},
}

func main() {
	osArgs := os.Args[1:]
	os.Exit(cli.RunWithPlugins(osArgs, []api.Plugin{scssPlugin}))
}
