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

type NodeModulesImportResolver struct{
	build api.PluginBuild
}

func (resolver NodeModulesImportResolver) CanonicalizeURL(url string) (string, error) {
	var filePath = url
	if strings.HasSuffix(url, "scss") {
		result := resolver.build.Resolve(url, api.ResolveOptions{
							Kind:      api.ResolveCSSImportRule,
							ResolveDir: ".",
						})
		filePath = result.Path
	} else {
		packagePath, fileName := filepath.Split(url)
		filePath = filepath.Join(packagePath, "_"+fileName+".scss")
	}

	if strings.HasPrefix(filePath, "file:") {
		return filePath, nil
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", nil
	}

	path, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("Error converting relative path to absolute path", err)
	}

	return "file://" + path, nil
}

func (resolver NodeModulesImportResolver) Load(canonicalizedURL string) (godartsass.Import, error) {
	path := canonicalizedURL[7:]

	content, err := os.ReadFile(path)
	if err != nil {
		return godartsass.Import{}, err
	}

	// Return the parsed import data
	return godartsass.Import{
		Content:      string(content),
		SourceSyntax: findSourceSyntax(path),
	}, nil
}

func compileSass(inputPath, outputPath string, build api.PluginBuild) (string, error) {
	// Read the input Sass/SCSS file
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return "", err
	}

	// add sass to the path
	current, err := os.Executable()
	if err != nil {
		return "", err
	}
	bin := filepath.Dir(current)
	pack := filepath.Dir(bin)
	dartSass := filepath.Join(filepath.Dir(pack), "dart-sass")
	os.Setenv("PATH", os.Getenv("PATH")+":"+dartSass)

	sourceSyntax := findSourceSyntax(inputPath)

	// Create a Dart Sass compiler
	compiler, err := godartsass.Start(godartsass.Options{})
	if err != nil {
		return "", err
	}
	defer compiler.Close()

	// Compile the Sass/SCSS to CSS
	output, err := compiler.Execute(godartsass.Args{
		Source:          string(input),
		OutputStyle:     godartsass.OutputStyleCompressed,
		SourceSyntax:    sourceSyntax,
		IncludePaths:    []string{filepath.Dir(inputPath)},
		EnableSourceMap: true,
		ImportResolver:  NodeModulesImportResolver{
			build,
		},
	})
	if err != nil {
		return "", err
	}

	return output.CSS, nil
}

func findSourceSyntax(inputPath string) godartsass.SourceSyntax {
	extension := filepath.Ext(inputPath)
	var sourceSyntax godartsass.SourceSyntax
	if extension == ".scss" {
		sourceSyntax = godartsass.SourceSyntaxSCSS
	} else {
		sourceSyntax = godartsass.SourceSyntaxSASS
	}
	return sourceSyntax
}

var scssPlugin = api.Plugin{
	Name: "sass-loader",
	Setup: func(build api.PluginBuild) {
		build.OnLoad(api.OnLoadOptions{Filter: `^.*(scss|sass)$`},
			func(args api.OnLoadArgs) (api.OnLoadResult, error) {
				// Compile the Sass/SCSS file to CSS
				extension := filepath.Ext(args.Path)
				filenameWithoutExtension := strings.TrimSuffix(args.Path, extension)
				outputPath := filenameWithoutExtension + ".css"
				result, err := compileSass(args.Path, outputPath, build)
				if err != nil {
					return api.OnLoadResult{}, err
				}

				// Modify the import path to the generated CSS file
				args.Path = outputPath

				return api.OnLoadResult{Contents: &result, Loader: api.LoaderCSS}, nil
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
