package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	// add sass to the path
	current, err := os.Executable()
	if err != nil {
		return err
	}
	bin := filepath.Dir(current)
	pack := filepath.Dir(bin)
	dartSass := filepath.Join(filepath.Dir(pack), "dart-sass")
	os.Setenv("PATH", os.Getenv("PATH")+":"+dartSass)

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
		IncludePaths:    []string{filepath.Dir(inputPath)},
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
	argsEnd := 0
	for _, arg := range osArgs {
		switch {
		case arg == "--version":
			fmt.Printf("%s-scss\n", "0.19.9")
			os.Exit(0)
		case arg == "--watch" || arg == "--watch=forever":
			go func() {
				// This just discards information from stdin because we don't use
				// it and we can avoid unnecessarily allocating space for it
				buffer := make([]byte, 512)
				for {
					_, err := os.Stdin.Read(buffer)
					if err != nil {

						// Only exit cleanly if stdin was closed cleanly
						if err == io.EOF {
							os.Exit(0)
						} else {
							os.Exit(1)
						}
					}

					// Some people attempt to keep esbuild's watch mode open by piping
					// an infinite stream of data to stdin such as with "< /dev/zero".
					// This will make esbuild spin at 100% CPU. To avoid this, put a
					// small delay after we read some data from stdin.
					time.Sleep(4 * time.Millisecond)
				}
			}()
			if arg != "--watch" {
				osArgs = append(osArgs[:argsEnd], osArgs[argsEnd+1:]...)
				osArgs = append(osArgs, "--watch")
			}
		}
		argsEnd++
	}
	os.Exit(cli.RunWithPlugins(osArgs, []api.Plugin{scssPlugin}))
}
