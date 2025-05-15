package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/bep/godartsass/v2"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/evanw/esbuild/pkg/cli"

	_ "embed"
)

//go:embed version.txt
var version string

type NodeModulesImportResolver struct {
	build        api.PluginBuild
	inputPath    string
	includeFiles []string
}

type SassCompileResult struct {
	output       string
	includeFiles []string
	err          error
}

func NodeResolve(filePath string, build api.PluginBuild) (string, []api.Message) {
	result := build.Resolve(filePath, api.ResolveOptions{
		Kind:       api.ResolveCSSImportRule,
		ResolveDir: ".",
	})
	return result.Path, result.Errors
}

func LocalFile(dir string, filePath string) (string, error) {
	localFilePath := filepath.Join(dir, filePath)
	if _, err := os.Stat(localFilePath); os.IsNotExist(err) {
		return "", err
	}
	return localFilePath, nil
}

func LocalOrNodeResolve(filePath string, dir string, build api.PluginBuild) (string, error) {
	nodeResult, errs := NodeResolve(filePath, build)

	if errs == nil {
		return nodeResult, nil
	}

	localFilePath, err := LocalFile(dir, filePath)

	if err == nil {
		return localFilePath, nil
	}

	return "", fmt.Errorf("not found")
}

func (resolver *NodeModulesImportResolver) CanonicalizeURL(filePath string) (string, error) {
	dir, _ := filepath.Split(resolver.inputPath)
	if !strings.HasSuffix(filePath, "scss") {
		filePath = filePath + ".scss"
	}

	u, err := url.Parse(filePath)
	if err == nil && u.Scheme == "file" {
		filePath = u.Path
		dir = ""
	}

	file, err := LocalOrNodeResolve(filePath, dir, resolver.build)
	if err == nil {
		resolver.includeFiles = append(resolver.includeFiles, file)
		return "file://" + file, nil
	}

	packagePath, fileName := filepath.Split(filePath)
	fileWithPrefix := filepath.Join(packagePath, "_"+fileName)

	filePrefix, err := LocalOrNodeResolve(fileWithPrefix, dir, resolver.build)
	if err == nil {
		resolver.includeFiles = append(resolver.includeFiles, filePrefix)
		return "file://" + filePrefix, nil
	}

	return "", fmt.Errorf("failed to canonicalize URL '%s': %w", filePath, err)
}

func (resolver NodeModulesImportResolver) Load(canonicalizedURL string) (godartsass.Import, error) {
	u, err := url.Parse(canonicalizedURL)
	if err == nil && u.Scheme == "file" {
		canonicalizedURL = u.Path
	}

	content, err := os.ReadFile(canonicalizedURL)
	if err != nil {
		return godartsass.Import{}, fmt.Errorf("failed to read file '%s': %w", canonicalizedURL, err)
	}

	// Return the parsed import data
	return godartsass.Import{
		Content:      string(content),
		SourceSyntax: findSourceSyntax(canonicalizedURL),
	}, nil
}

func compileSass(inputPath string, build api.PluginBuild) SassCompileResult {
	// Read the input Sass/SCSS file
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return SassCompileResult{err: fmt.Errorf("failed to read input file '%s': %w", inputPath, err)}
	}

	// add sass to the path
	current, err := os.Executable()
	if err != nil {
		return SassCompileResult{err: fmt.Errorf("failed to get executable path: %w", err)}
	}
	bin := filepath.Dir(current)
	pack := filepath.Dir(bin)
	dartSass := filepath.Join(filepath.Dir(pack), "dart-sass", "sass")

	sourceSyntax := findSourceSyntax(inputPath)

	// Create a Dart Sass compiler
	compiler, err := godartsass.Start(godartsass.Options{
		DartSassEmbeddedFilename: dartSass,
	})
	if err != nil {
		return SassCompileResult{err: fmt.Errorf("failed to start Dart Sass compiler: %w", err)}
	}
	defer compiler.Close()

	resolver := NodeModulesImportResolver{
		build,
		inputPath,
		[]string{},
	}
	// Compile the Sass/SCSS to CSS
	output, err := compiler.Execute(godartsass.Args{
		Source:          string(input),
		OutputStyle:     godartsass.OutputStyleCompressed,
		SourceSyntax:    sourceSyntax,
		IncludePaths:    []string{filepath.Dir(inputPath)},
		EnableSourceMap: true,
		ImportResolver:  &resolver,
	})
	if err != nil {
		return SassCompileResult{err: fmt.Errorf("failed to execute Dart Sass compilation: %w", err)}
	}

	return SassCompileResult{output: output.CSS, includeFiles: resolver.includeFiles, err: nil}
}

func findSourceSyntax(inputPath string) godartsass.SourceSyntax {
	extension := filepath.Ext(inputPath)
	var sourceSyntax = godartsass.SourceSyntaxSCSS
	if extension == ".sass" {
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
				result := compileSass(args.Path, build)
				if result.err != nil {
					return api.OnLoadResult{}, fmt.Errorf("failed to compile Sass file '%s': %w", args.Path, result.err)
				}

				// Modify the import path to the generated CSS file
				args.Path = outputPath
				return api.OnLoadResult{Contents: &result.output, Loader: api.LoaderCSS, WatchFiles: result.includeFiles}, nil
			})
	},
}

var tailwindPlugin = api.Plugin{
	Name: "tailwind-loader",
	Setup: func(build api.PluginBuild) {
		build.OnLoad(api.OnLoadOptions{Filter: `\.css$`},
			func(args api.OnLoadArgs) (api.OnLoadResult, error) {
				content, err := os.ReadFile(args.Path)
				if err != nil {
					return api.OnLoadResult{}, fmt.Errorf("to load css file from '%s': %w", args.Path, err)
				}
				cssContent := string(content)
				if strings.Contains(cssContent, "tailwind") {

					// Create a temporary input file with the CSS content
					tempDir := os.TempDir()
					outputPath := filepath.Join(tempDir, "tailwind.output.css")

					cmd := exec.Command("tailwindcss",
						"-i", args.Path,
						"-o", outputPath,
						"--minify",
					)

					// Run the Tailwind CLI
					output, err := cmd.CombinedOutput()
					if err != nil {
						return api.OnLoadResult{}, fmt.Errorf("failed to run tailwindcss on '%s': %v\n%s", args.Path, err, string(output))
					}

					cssContent, err := os.ReadFile(outputPath)
					if err != nil {
						return api.OnLoadResult{}, fmt.Errorf("failed to read generated tailwind css from '%s': %w", outputPath, err)
					}

					// Clean up the temporary output file
					os.Remove(outputPath)

					cssContentStr := string(cssContent)
					return api.OnLoadResult{Contents: &cssContentStr, Loader: api.LoaderCSS, WatchFiles: []string{"tailwind.config.js"}}, nil
				}
				return api.OnLoadResult{Contents: &cssContent, Loader: api.LoaderCSS}, nil
			})
	},
}

func main() {
	osArgs := os.Args[1:]
	argsEnd := 0
	for _, arg := range osArgs {
		switch {
		case arg == "--version":
			fmt.Printf("%s", version)
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
	os.Exit(cli.RunWithPlugins(osArgs, []api.Plugin{scssPlugin, tailwindPlugin}))
}
