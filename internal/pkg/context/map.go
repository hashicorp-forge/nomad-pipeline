package context

import (
	"time"
)

func (c *Context) AsMap() map[string]any {
	m := map[string]any{
		"nomad_pipeline": c.nomadPipelineAsMap(),
	}

	if c.Specifications != nil {
		m["specifications"] = c.specificationsAsMap()
	}

	if c.Inline != nil {
		m["inline"] = c.inlineAsMap()
	}

	if c.Variables != nil {
		m["var"] = c.Variables["var"].(map[string]any)
	}

	return m
}

func (c *Context) nomadPipelineAsMap() map[string]any {
	if c.NomadPipeline == nil {
		return map[string]any{}
	}

	return map[string]any{
		"flow_id":     c.NomadPipeline.FlowID,
		"flow_type":   c.NomadPipeline.FlowType,
		"namespace":   c.NomadPipeline.Namespace,
		"run_id":      c.NomadPipeline.RunID,
		"status":      c.NomadPipeline.Status,
		"trigger":     c.NomadPipeline.Trigger,
		"create_time": formatTime(c.NomadPipeline.CreateTime),
		"start_time":  formatTime(c.NomadPipeline.StartTime),
		"end_time":    formatTime(c.NomadPipeline.EndTime),
	}
}

func (c *Context) specificationsAsMap() map[string]any {
	m := make(map[string]any, len(c.Specifications))

	for _, spec := range c.Specifications {
		m[spec.ID] = spec.asMap()
	}

	return m
}

func (s *SpecificationContext) asMap() map[string]any {
	m := map[string]any{
		"id":         s.ID,
		"status":     s.Status,
		"start_time": formatTime(s.StartTime),
		"end_time":   formatTime(s.EndTime),
	}

	if s.NomadJobID != "" {
		m["nomad_job_id"] = s.NomadJobID
	}

	if s.NomadJobNamespace != "" {
		m["nomad_job_namespace"] = s.NomadJobNamespace
	}

	return m
}

func (c *Context) inlineAsMap() map[string]any {
	if c.Inline == nil {
		return map[string]any{}
	}

	m := map[string]any{
		"id":    c.Inline.ID,
		"steps": c.stepsAsMap(),
	}

	return m
}

func (c *Context) stepsAsMap() map[string]any {
	if c.Inline == nil || c.Inline.Steps == nil {
		return map[string]any{}
	}

	m := make(map[string]any, len(c.Inline.Steps))

	for _, step := range c.Inline.Steps {
		m[step.ID] = step.asMap()
	}

	return m
}

func (s *StepContext) asMap() map[string]any {
	return map[string]any{
		"id":         s.ID,
		"status":     s.Status,
		"exit_code":  s.ExitCode,
		"start_time": formatTime(s.StartTime),
		"end_time":   formatTime(s.EndTime),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
