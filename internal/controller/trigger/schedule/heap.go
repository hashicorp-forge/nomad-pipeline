package schedule

import (
	"time"

	"github.com/hashicorp/cronexpr"

	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type scheduledTrigger struct {
	trigger  *state.Trigger
	nextRun  time.Time
	cronExpr *cronexpr.Expression
	index    int
}

type triggerHeap []*scheduledTrigger

func (h triggerHeap) Len() int { return len(h) }

func (h triggerHeap) Less(i, j int) bool { return h[i].nextRun.Before(h[j].nextRun) }

func (h triggerHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *triggerHeap) Push(x any) {
	n := len(*h)
	item := x.(*scheduledTrigger)
	item.index = n
	*h = append(*h, item)
}

func (h *triggerHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}
