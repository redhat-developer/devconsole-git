# devconsole-git
This service provides functionality for communications with git providers in DevConsole

### Pre-requisites
- [dep][dep_tool] version v0.5.0+.
- [git][git_tool]
- [go][go_tool] version v1.10+.
- [docker][docker_tool] version 17.03+.

### To Download Dependencies
```
make deps
```

### Build 
```
make build
```

### Clean 
```
make clean
```

### Deploy the operator in dev mode

* Make sure minishift is running
* In dev mode, simply run your operator locally:
```
make local
```
> NOTE: To watch all namespaces, `APP_NAMESPACE` is set to empty string. 
If a specific namespace is provided only that project will watched. 
As we reuse `openshift`'s imagestreams for build, we need to access all namespaces.
 
### Deploy the operator with Deployment yaml

* Make sure minishift is running
* To deploy all necessary objects including the operator yaml file run:
```
make deploy-all
```

* Clean previously created resources
```
make clean-resources
```

[dep_tool]:https://golang.github.io/dep/docs/installation.html
[go_tool]:https://golang.org/dl/
[git_tool]:https://git-scm.com/downloads
[docker_tool]:https://docs.docker.com/install/
