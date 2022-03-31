// Package admission handles kubernetes admissions,
// it takes admission requests and returns admission reviews;
// for example, to mutate or validate pods
package admission

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/slackhq/simple-kubernetes-webhook/pkg/mutation"
	"github.com/slackhq/simple-kubernetes-webhook/pkg/validation"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	appsv1 "k8s.io/api/apps/v1"
)

// Admitter is a container for admission business
type Admitter struct {
	Logger  *logrus.Entry
	Request *admissionv1.AdmissionRequest
}

// MutatePodReview takes an admission request and mutates the pod within,
// it returns an admission review with mutations as a json patch (if any)
func (a Admitter) MutatePodReview() (*admissionv1.AdmissionReview, error) {
	pod, err := a.Pod()
	if err != nil {
		e := fmt.Sprintf("could not parse pod in admission review request: %v", err)
		return reviewResponse(a.Request.UID, false, http.StatusBadRequest, e), err
	}

	m := mutation.NewMutator(a.Logger)
	patch, err := m.MutatePodPatch(pod)
	if err != nil {
		e := fmt.Sprintf("could not mutate pod: %v", err)
		return reviewResponse(a.Request.UID, false, http.StatusBadRequest, e), err
	}

	return patchReviewResponse(a.Request.UID, patch)
}

// MutatePodReview takes an admission request and validates the pod within
// it returns an admission review
func (a Admitter) ValidateReview() (*admissionv1.AdmissionReview, error) {
	var err error
        var spec corev1.PodSpec
	var namespace string
	switch a.Request.Kind.Kind{
		case "Pod":
			var pod *corev1.Pod
			pod, err = a.Pod()
			spec = pod.Spec
			namespace = pod.ObjectMeta.Namespace
		case "ReplicaSet":
			var rs *appsv1.ReplicaSet
			rs, err = a.ReplicaSet()
			spec = rs.Spec.Template.Spec
			namespace = rs.ObjectMeta.Namespace
	}
	if err != nil {
		e := fmt.Sprintf("could not parse pod in admission review request: %v", err)
		return reviewResponse(a.Request.UID, false, http.StatusBadRequest, e), err
	}

	v := validation.NewValidator(a.Logger)
	val, err := v.ValidatePodSpec(spec, namespace, a.Request.UserInfo.Username)
	if err != nil {
		e := fmt.Sprintf("could not validate pod: %v", err)
		return reviewResponse(a.Request.UID, false, http.StatusBadRequest, e), err
	}

	if !val.Valid {
		return reviewResponse(a.Request.UID, false, http.StatusForbidden, val.Reason), nil
	}

	return reviewResponse(a.Request.UID, true, http.StatusAccepted, "valid pod"), nil
}

// Pod extracts a pod from an admission request
func (a Admitter) Pod() (*corev1.Pod, error) {
	p := corev1.Pod{}
	if err := json.Unmarshal(a.Request.Object.Raw, &p); err != nil {
		return nil, err
	}

	return &p, nil
}

func (a Admitter) ReplicaSet() (*appsv1.ReplicaSet, error) {
        p := appsv1.ReplicaSet{}
        if err := json.Unmarshal(a.Request.Object.Raw, &p); err != nil {
                return nil, err
        }

        return &p, nil
}

// reviewResponse TODO: godoc
func reviewResponse(uid types.UID, allowed bool, httpCode int32,
	reason string) *admissionv1.AdmissionReview {
	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     uid,
			Allowed: allowed,
			Result: &metav1.Status{
				Code:    httpCode,
				Message: reason,
			},
		},
	}
}

// patchReviewResponse builds an admission review with given json patch
func patchReviewResponse(uid types.UID, patch []byte) (*admissionv1.AdmissionReview, error) {
	patchType := admissionv1.PatchTypeJSONPatch

	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:       uid,
			Allowed:   true,
			PatchType: &patchType,
			Patch:     patch,
		},
	}, nil
}
