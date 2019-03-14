package apiserver

import (
	"fmt"
	"github.com/MatousJobanek/build-environment-detector/detector"
	"github.com/MatousJobanek/build-environment-detector/detector/git"
	"strings"

	"github.com/emicklei/go-restful"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/version"
	genericapiserver "k8s.io/apiserver/pkg/server"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

const GroupName = "kubepoc.bryzak.com"
const GroupVersion = "v1"

var (
	Scheme             = runtime.NewScheme()
	Codecs             = serializer.NewCodecFactory(Scheme)
	SchemeBuilder      = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme        = SchemeBuilder.AddToScheme
	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: GroupVersion}
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion)
	return nil
}

func init() {
	utilruntime.Must(AddToScheme(Scheme))

	// Setting VersionPriority is critical in the InstallAPIGroup call (done in New())
	utilruntime.Must(Scheme.SetVersionPriority(SchemeGroupVersion))
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: GroupVersion})

	unversioned := schema.GroupVersion{Group: "", Version: GroupVersion}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}

type ExtraConfig struct {
	// Place you custom config here.
}

type Config struct {
	GenericConfig *genericapiserver.RecommendedConfig
	ExtraConfig   ExtraConfig
}

// PocServer contains state for a Kubernetes cluster master/api server.
type PocServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *ExtraConfig
}

type CompletedConfig struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfig
}

type GitSource struct {
	Source string
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (cfg *Config) Complete() CompletedConfig {
	c := completedConfig{
		cfg.GenericConfig.Complete(),
		&cfg.ExtraConfig,
	}

	c.GenericConfig.Version = &version.Info{
		Major: "1",
		Minor: "0",
	}

	return CompletedConfig{&c}
}

// New returns a new instance of ProvenanceServer from the given config.
func (c completedConfig) New() (*PocServer, error) {
	genericServer, err := c.GenericConfig.New("kube proof of concept server", genericapiserver.NewEmptyDelegate())
	if err != nil {
		return nil, err
	}

	s := &PocServer{
		GenericAPIServer: genericServer,
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(GroupName, Scheme, metav1.ParameterCodec, Codecs)

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	installCompositionPocWebService(s)

	return s, nil
}

func installCompositionPocWebService(pocServer *PocServer) {
	path := fmt.Sprintf("/apis/%s/%s/namespaces/{namespace}", GroupName, GroupVersion)
	fmt.Println("WS PATH:" + path)

	ws := getWebService()
	ws.Path(path).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	testPath := "/{resource-id}/test"
	fmt.Println("Test Path:" + testPath)
	ws.Route(ws.GET(testPath).To(testResponse))

	detectPath := "/detect"
	fmt.Println("Detect Path:" + detectPath)
	ws.Route(ws.POST(detectPath).To(detectResponse))

	pocServer.GenericAPIServer.Handler.GoRestfulContainer.Add(ws)
	fmt.Println("Done registering.")
}

func getWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/apis")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON, restful.MIME_XML)
	ws.ApiVersion(GroupName)
	return ws
}

func testResponse(request *restful.Request, response *restful.Response) {
	fmt.Println("Handling request...")
	resourceName := request.PathParameter("resource-id")
	requestPath := request.Request.URL.Path
	resourcePathSlice := strings.Split(requestPath, "/")
	resourceKind := resourcePathSlice[6] // Kind is 7th element in the slice
	responseString := "Resource Name:" + resourceName + " Resource Kind: " + resourceKind + "\n"
	response.Write([]byte(responseString))
}

func detectResponse(request *restful.Request, response *restful.Response) {
	fmt.Println("Handling request...")

	gitSource := &GitSource{}

	err := request.ReadEntity(gitSource)
	if err != nil {
		response.Write([]byte((err.Error())))
	}

	fmt.Println("Detecting git source: " + gitSource.Source)

	src := &git.Source{
		URL:    gitSource.Source,
		Secret: git.NewUsernamePassword("anonymous", ""),
	}

	buildEnvStats, err := detector.DetectBuildEnvironments(src)
	if err != nil {
		response.Write([]byte((err.Error())))
	}

	if buildEnvStats != nil {
		for _, lang := range buildEnvStats.SortedLanguages {
			response.Write([]byte("Language found: "))
			response.Write([]byte(lang))
			response.Write([]byte("\r\n"))
		}

		for _, tool := range buildEnvStats.DetectedBuildTools {
			response.Write([]byte("Build tool found: "))
			response.Write([]byte(fmt.Sprint(*tool)))
			response.Write([]byte("\r\n"))
		}
	} else {
		response.Write([]byte("No languages detected"))
	}
}
