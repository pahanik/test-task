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
var _ podValidator = (*cpuRequestValidator)(nil)

// Name returns the name of cpuValidator
func (n cpuRequestValidator) Name() string {
	return "cpu_request_validator"
}

// Validate inspects the cpu requests of a given pod and returns validation.
// The returned validation is only valid if the pod cpu request is less or
// equal predefined value. It works only in namespaces defined by ConfigMap
func (n cpuRequestValidator) Validate(pod *corev1.Pod) (validation, error) {
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
	if _, ok := configmap.Data[pod.ObjectMeta.Namespace]; ok {
		cpuRequest := resource.NewMilliQuantity(2000, resource.DecimalSI)
		for _, container := range pod.Spec.Containers {
			if r, ok := container.Resources.Requests[corev1.ResourceCPU]; ok && r.MilliValue() > cpuRequest.MilliValue() {
				return validation{Valid: false, Reason: fmt.Sprintf("Container %s has cpu request %vm > %vm in %s namespace. Validated namespaces: %v", container.Name, r.MilliValue(), cpuRequest.MilliValue(), pod.ObjectMeta.Namespace, configmap.Data)}, nil
			}
		}
		for _, container := range pod.Spec.InitContainers {
			if r, ok := container.Resources.Requests[corev1.ResourceCPU]; ok && r.MilliValue() > cpuRequest.MilliValue() {
                        	return validation{Valid: false, Reason: fmt.Sprintf("Init container %s has cpu request %vm > %vm in %s namespace. Validated namespaces: %v", container.Name, r.MilliValue(), cpuRequest.MilliValue(), pod.ObjectMeta.Namespace, configmap.Data)}, nil
                	}
		}
	}
	return validation{Valid: true, Reason: "valid cpu request"}, nil
}
