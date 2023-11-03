package main

import "encoding/json"
import "os"
import "strings"
import "github.com/evanw/esbuild/pkg/api"

import "C"

var envPlugin = api.Plugin{
  Name: "env",
  Setup: func(build api.PluginBuild) {
    // Intercept import paths called "env" so esbuild doesn't attempt
    // to map them to a file system location. Tag them with the "env-ns"
    // namespace to reserve them for this plugin.
    build.OnResolve(api.OnResolveOptions{Filter: `^env$`},
      func(args api.OnResolveArgs) (api.OnResolveResult, error) {
        return api.OnResolveResult{
          Path:      args.Path,
          Namespace: "env-ns",
        }, nil
      })

    // Load paths tagged with the "env-ns" namespace and behave as if
    // they point to a JSON file containing the environment variables.
    build.OnLoad(api.OnLoadOptions{Filter: `.*`, Namespace: "env-ns"},
      func(args api.OnLoadArgs) (api.OnLoadResult, error) {
        mappings := make(map[string]string)
        for _, item := range os.Environ() {
          if equals := strings.IndexByte(item, '='); equals != -1 {
            mappings[item[:equals]] = item[equals+1:]
          }
        }
        bytes, err := json.Marshal(mappings)
        if err != nil {
          return api.OnLoadResult{}, err
        }
        contents := string(bytes)
        return api.OnLoadResult{
          Contents: &contents,
          Loader:   api.LoaderJSON,
        }, nil
      })
  },
}

//export main
func main() {
  result := api.Build(api.BuildOptions{
    EntryPoints: []string{"app.js"},
    Bundle:      true,
    Outfile:     "out.js",
    Plugins:     []api.Plugin{envPlugin},
    Write:       true,
  })

  if len(result.Errors) > 0 {
    os.Exit(1)
  }
}
