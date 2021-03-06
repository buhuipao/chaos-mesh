// Copyright 2019 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KindPodChaos is the kind for pod chaos
const KindPodChaos = "PodChaos"

func init() {
	all.register(KindPodChaos, &ChaosKind{
		Chaos:     &PodChaos{},
		ChaosList: &PodChaosList{},
	})
}

// PodChaosAction represents the chaos action about pods.
type PodChaosAction string

const (
	// PodKillAction represents the chaos action of killing pods.
	PodKillAction PodChaosAction = "pod-kill"
	// PodFailureAction represents the chaos action of injecting errors to pods.
	// This action will cause the pod to not be created for a while.
	PodFailureAction PodChaosAction = "pod-failure"
	// ContainerKillAction represents the chaos action of killing the container
	ContainerKillAction PodChaosAction = "container-kill"
)

// +kubebuilder:object:root=true

// PodChaos is the control script`s spec.
type PodChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a pod chaos experiment
	Spec PodChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the chaos experiment about pods
	Status PodChaosStatus `json:"status"`
}

func (in *PodChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// IsDeleted returns whether this resource has been deleted
func (in *PodChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *PodChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetDuration would return the duration for chaos
func (in *PodChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *PodChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

func (in *PodChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

func (in *PodChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

func (in *PodChaos) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextRecover = nil
		return
	}

	if in.Status.Scheduler.NextRecover == nil {
		in.Status.Scheduler.NextRecover = &metav1.Time{}
	}
	in.Status.Scheduler.NextRecover.Time = t
}

// GetScheduler would return the scheduler for chaos
func (in *PodChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetChaos returns a chaos instance
func (in *PodChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindPodChaos,
		StartTime: in.CreationTimestamp.Time,
		Action:    string(in.Spec.Action),
		Status:    string(in.GetStatus().Experiment.Phase),
	}
	if in.Spec.Duration != nil {
		instance.Duration = *in.Spec.Duration
	}
	if in.DeletionTimestamp != nil {
		instance.EndTime = in.DeletionTimestamp.Time
	}
	return instance
}

// PodChaosSpec defines the attributes that a user creates on a chaos experiment about pods.
type PodChaosSpec struct {
	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`

	// Scheduler defines some schedule rules to
	// control the running time of the chaos experiment about pods.
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

	// Action defines the specific pod chaos action.
	// Supported action: pod-kill / pod-failure / container-kill
	// Default action: pod-kill
	// +kubebuilder:validation:Enum=pod-kill;pod-failure;container-kill
	Action PodChaosAction `json:"action"`

	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	// +kubebuilder:validation:Enum=one;all;fixed;fixed-percent;random-max-percent
	Mode PodMode `json:"mode"`

	// Value is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the percent of pods the server can do chaos action.
	// IF `RandomMaxPercentPodMod`,  provide a number from 0-100 to specify the max percent of pods to do chaos action
	// +optional
	Value string `json:"value"`

	// Duration represents the duration of the chaos action.
	// It is required when the action is `PodFailureAction`.
	// A duration string is a possibly signed sequence of
	// decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// +optional
	Duration *string `json:"duration,omitempty"`

	// ContainerName indicates the name of the container.
	// Needed in container-kill.
	// +optional
	ContainerName string `json:"containerName"`

	// GracePeriod is used in pod-kill action. It represents the duration in seconds before the pod should be deleted.
	// Value must be non-negative integer. The default value is zero that indicates delete immediately.
	// +optional
	// +kubebuilder:validation:Minimum=0
	GracePeriod int64 `json:"gracePeriod"`
}

func (in *PodChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

func (in *PodChaosSpec) GetMode() PodMode {
	return in.Mode
}

func (in *PodChaosSpec) GetValue() string {
	return in.Value
}

// +kubebuilder:object:root=true

// PodChaosList is PodChaos list.
type PodChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PodChaos `json:"items"`
}

// PodChaosStatus represents the current status of the chaos experiment about pods.
type PodChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// PodStatus represents information about the status of a pod in chaos experiment.
type PodStatus struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Action    string `json:"action"`
	HostIP    string `json:"hostIP"`
	PodIP     string `json:"podIP"`

	// A brief CamelCase message indicating details about the chaos action.
	// e.g. "delete this pod" or "pause this pod duration 5m"
	// +optional
	Message string `json:"message"`
}

// ListChaos returns a list of pod chaos
func (in *PodChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
}

func init() {
	SchemeBuilder.Register(&PodChaos{}, &PodChaosList{})
}
