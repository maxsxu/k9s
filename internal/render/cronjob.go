// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// CronJob renders a K8s CronJob to screen.
type CronJob struct {
	Base
}

// Header returns a header row.
func (c CronJob) Header(_ string) model1.Header {
	return c.doHeader(c.defaultHeader())
}

func (CronJob) defaultHeader() model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "VS", Attrs: model1.Attrs{VS: true}},
		model1.HeaderColumn{Name: "SCHEDULE"},
		model1.HeaderColumn{Name: "SUSPEND"},
		model1.HeaderColumn{Name: "ACTIVE"},
		model1.HeaderColumn{Name: "LAST_SCHEDULE", Attrs: model1.Attrs{Time: true}},
		model1.HeaderColumn{Name: "SELECTOR", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "CONTAINERS", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "IMAGES", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
	}
}

// Render renders a K8s resource to screen.
func (c CronJob) Render(o interface{}, ns string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected CronJob, but got %T", o)
	}
	if err := c.defaultRow(raw, row); err != nil {
		return err
	}
	if c.specs.isEmpty() {
		return nil
	}

	cols, err := c.specs.realize(raw, c.defaultHeader(), row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

// Render renders a K8s resource to screen.
func (c CronJob) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	var cj batchv1.CronJob
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &cj)
	if err != nil {
		return err
	}

	lastScheduled := "<none>"
	if cj.Status.LastScheduleTime != nil {
		lastScheduled = ToAge(*cj.Status.LastScheduleTime)
	}

	r.ID = client.MetaFQN(cj.ObjectMeta)
	r.Fields = model1.Fields{
		cj.Namespace,
		cj.Name,
		computeVulScore(cj.ObjectMeta, &cj.Spec.JobTemplate.Spec.Template.Spec),
		cj.Spec.Schedule,
		boolPtrToStr(cj.Spec.Suspend),
		strconv.Itoa(len(cj.Status.Active)),
		lastScheduled,
		jobSelector(cj.Spec.JobTemplate.Spec),
		podContainerNames(cj.Spec.JobTemplate.Spec.Template.Spec, true),
		podImageNames(cj.Spec.JobTemplate.Spec.Template.Spec, true),
		mapToStr(cj.Labels),
		"",
		ToAge(cj.GetCreationTimestamp()),
	}

	return nil
}

// Helpers

func jobSelector(spec batchv1.JobSpec) string {
	if spec.Selector == nil {
		return MissingValue
	}
	if len(spec.Selector.MatchLabels) > 0 {
		return mapToStr(spec.Selector.MatchLabels)
	}
	if len(spec.Selector.MatchExpressions) == 0 {
		return ""
	}

	ss := make([]string, 0, len(spec.Selector.MatchExpressions))
	for _, e := range spec.Selector.MatchExpressions {
		ss = append(ss, e.String())
	}

	return strings.Join(ss, " ")
}

func podContainerNames(spec v1.PodSpec, includeInit bool) string {
	cc := make([]string, 0, len(spec.Containers)+len(spec.InitContainers))

	if includeInit {
		for _, c := range spec.InitContainers {
			cc = append(cc, c.Name)
		}
	}
	for _, c := range spec.Containers {
		cc = append(cc, c.Name)
	}

	return strings.Join(cc, ",")
}

func podImageNames(spec v1.PodSpec, includeInit bool) string {
	cc := make([]string, 0, len(spec.Containers)+len(spec.InitContainers))

	if includeInit {
		for _, c := range spec.InitContainers {
			cc = append(cc, c.Image)
		}
	}
	for _, c := range spec.Containers {
		cc = append(cc, c.Image)
	}

	return strings.Join(cc, ",")
}
