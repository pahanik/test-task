package validation

import (
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// cpuValidator is a container for validating the cpu requests of pods
type cpuRequestValidator struct {
	Logger logrus.FieldLogger
}

// cpuValidator implements the podValidator interface
var _ podSpecValidator = (*cpuRequestValidator)(nil)

// Name returns the name of cpuValidator
func (n cpuRequestValidator) Name() string {
	return "cpu_request_validator"
}

func (n cpuRequestValidator) ApplyValidationRules(container corev1.Container, user string, cont_type string, namespace string, namespaces map[string]string) (validation, error) {
        cpuRequest := resource.NewMilliQuantity(2000, resource.DecimalSI)
        adminCpuRequest := resource.NewMilliQuantity(3000, resource.DecimalSI)
        r, ok := container.Resources.Requests[corev1.ResourceCPU]
        if ok {
                if ! ( user == "kubernetes-admin" && r.MilliValue() > adminCpuRequest.MilliValue() ) {
                        if r.MilliValue() > cpuRequest.MilliValue() {
                                return validation{Valid: false, Reason: fmt.Sprintf("%s %s has cpu request %vm > %vm in %s namespace. Validated namespaces: %v", cont_type, container.Name, r.MilliValue(), cpuRequest.MilliValue(), namespace, namespaces)}, nil
                        }
                }
        }
	return validation{Valid: true, Reason: "valid cpu request"}, nil
}

// Validate inspects the cpu requests of a given pod and returns validation.
// The returned validation is only valid if the pod cpu request is less or
// equal predefined value. It works only in namespaces defined by ConfigMap
func (n cpuRequestValidator) Validate(spec corev1.PodSpec, namespace string, user string) (validation, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	configmap, err := clientset.CoreV1().ConfigMaps("default").Get(context.TODO(), "validator-config", metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	if _, ok := configmap.Data[namespace]; ok {
		for _, container := range spec.Containers {
			if v, e := n.ApplyValidationRules(container, user, "Container", namespace, configmap.Data); nil == e && !v.Valid {
				return v, nil
			}
		}
		for _, container := range spec.InitContainers {
			if v, e := n.ApplyValidationRules(container, user, "Init container", namespace, configmap.Data); nil == e && !v.Valid {
				return v, nil
			}
		}
	}
	return validation{Valid: true, Reason: "valid cpu request"}, nil
}
