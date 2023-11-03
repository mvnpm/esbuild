
## POC to use esbuild with java and plugins

This is a poc to see if we can use esbuild in java, but more integrated.
Maybe we can even use esbuild and extend it with Java based plugins.

### To build

```cmd
go build -o libhello.so -buildmode=c-shared hello.go
chmod -x libhello.so

mvn package
export LD_LIBRARY_PATH=<full path to this folder>

java -jar target/go-helloworld-1.0.0-SNAPSHOT-jar-with-dependencies.jar
```

### TODO
* rename this project to not be called helloworld
* make the go code call java
* create an interface for the java and go communication