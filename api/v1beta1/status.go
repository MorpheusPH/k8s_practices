package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConditionStatus struct {
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
}

func (cs *ConditionStatus) SetConditions(conditionType string, status metav1.ConditionStatus, reason string, message string) {
	var c *Condition
	for i := range cs.Conditions {
		if cs.Conditions[i].Type == conditionType {
			c = &cs.Conditions[i]
		}
	}
	if c == nil {
		addCondition(cs, conditionType, status, reason, message)
	} else {
		if c.Status == status && c.Reason == reason && c.Message == message {
			return
		}
		now := metav1.Now()
		c.LastUpdateTime = now
		if c.Status != status {
			c.LastTransitionTime = now
		}
		c.Status = status
		c.Reason = reason
		c.Message = message
	}
}

func addCondition(cs *ConditionStatus, conditionType string, status metav1.ConditionStatus, reason string, message string) {
	now := metav1.Now()
	c := Condition{
		Type:               conditionType,
		LastUpdateTime:     now,
		LastTransitionTime: now,
		Status:             status,
		Reason:             reason,
		Message:            message,
	}
	cs.Conditions = append(cs.Conditions, c)
}

func (cs *ConditionStatus) FindCondition(conditionType string) *Condition {
	for i := range cs.Conditions {
		if cs.Conditions[i].Type == conditionType {
			return &cs.Conditions[i]
		}
	}

	return nil
}

type FailureStatus struct {
	Failures int64 `json:"failures,omitempty"`
}

func (fs *FailureStatus) AddFailureCounts() {
	if fs.Failures < 5 {
		fs.Failures++
	}
}

func (fs *FailureStatus) ResetFailureCounts() {
	fs.Failures = 0
}
