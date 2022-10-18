package controllers

import (
	"github.com/MorpheusPH/nginxcontroller/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatusObject interface {
	client.Object
	ConditionObject
	FailureObject
}

type ConditionObject interface {
	SetConditions(conditionType string, status metav1.ConditionStatus, reason string, message string)
	FindCondition(conditionType string) *v1beta1.Condition
}

type FailureObject interface {
	AddFailureCounts()
	ResetFailureCounts()
}

func findCondition(object StatusObject, conditionType string) *v1beta1.Condition {
	return object.FindCondition(conditionType)
}

func releaseProgressing(object StatusObject) {
	setReadyUnknownCondition(object, v1beta1.ProgressingReason, "Reconciliation in progress")
	object.ResetFailureCounts()
}

func releaseNotReady(object StatusObject, message string) {
	setNotReadyCondition(object, v1beta1.FailedReason, message)
}

// Ready - shortcut to set ready condition to true
func setReadyCondition(object StatusObject, reason string) {
	object.SetConditions(v1beta1.Ready, metav1.ConditionTrue, v1beta1.SuccessdedReason, "Reconcile Success")
}

// NotReady - shortcut to set ready condition to false
func setNotReadyCondition(object StatusObject, reason string, message string) {
	object.SetConditions(v1beta1.Ready, metav1.ConditionFalse, reason, message)
}

// Unknown - shortcut to set ready condition to unknown
func setReadyUnknownCondition(object StatusObject, reason string, message string) {
	object.SetConditions(v1beta1.Ready, metav1.ConditionUnknown, reason, message)
}

func addFailureCounts(object StatusObject) {
	object.AddFailureCounts()
}

// func setCondition(object StatusObject, conditionType string, status metav1.ConditionStatus, reason string, message string) {
// 	object.SetConditions(conditionType, status, reason, message)
// 	if status == metav1.ConditionFalse && reason == v1beta1.FailedReason {
// 		object.AddFailureCounts()
// 	}
// }
