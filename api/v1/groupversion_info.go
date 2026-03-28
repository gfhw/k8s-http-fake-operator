package v1

import (
	"os"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// getAPIGroup returns the API group from environment variable or uses default
func getAPIGroup() string {
	if group := os.Getenv("HTTP_TEST_STUB_API_GROUP"); group != "" {
		return group
	}
	return "httpteststub.example.com"
}

var (
	SchemeGroupVersion = schema.GroupVersion{
		Group:   getAPIGroup(),
		Version: "v1",
	}

	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&WebApp{},
		&WebAppList{},
		&HTTPTestStub{},
		&HTTPTestStubList{},
	)
	return nil
}
