package test

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const SecretName = "my-secret"

func PrepareClient(gvkObjects ...GvkObject) (client.Client, *runtime.Scheme) {
	var allObject []runtime.Object

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	for _, gvkObject := range gvkObjects {
		groupVersion, objects := gvkObject()
		s.AddKnownTypes(groupVersion, objects...)
		allObject = append(allObject, objects...)
	}

	// Create a fake client to mock API calls.
	return fake.NewFakeClient(allObject...), s
}

type GvkObject func() (schema.GroupVersion, []runtime.Object)

func RegisterGvkObject(gv schema.GroupVersion, types ...runtime.Object) GvkObject {
	return func() (schema.GroupVersion, []runtime.Object) {
		return gv, types
	}
}

func NewReconcileRequest(name string) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: Namespace,
		},
	}
}

func NewSecret(secretType corev1.SecretType, data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SecretName,
			Namespace: Namespace,
		},
		Type: secretType,
		Data: data,
	}
}
