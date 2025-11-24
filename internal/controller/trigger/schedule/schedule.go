package schedule

import (
	"container/heap"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/cronexpr"
	"go.uber.org/zap"

	serverstate "github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/logger"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

// TriggerCoordinator manages scheduled trigger execution using a heap-based priority queue
type Trigger struct {
	logger *zap.Logger
	state  serverstate.State

	runFn func(flowID, namespace, trigger string, vars map[string]any) error

	// heap is the priority queue of scheduled triggers
	heap *triggerHeap
	lock sync.RWMutex

	// triggers maps trigger ID to scheduled trigger info
	triggers map[string]*scheduledTrigger

	// control channels
	stopCh   chan struct{}
	updateCh chan struct{}
	wg       sync.WaitGroup
}

// TriggerCoordinatorConfig contains configuration for the trigger coordinator
type TriggerConfig struct {
	Logger *zap.Logger
	State  serverstate.State
	RunFn  func(flowID, namespace, trigger string, vars map[string]any) error
}

// NewTriggerCoordinator creates a new trigger coordinator
func NewTrigger(cfg *TriggerConfig) *Trigger {
	h := &triggerHeap{}
	heap.Init(h)

	return &Trigger{
		logger:   cfg.Logger.Named(logger.ComponentNameTriggerSchedule),
		state:    cfg.State,
		runFn:    cfg.RunFn,
		heap:     h,
		triggers: make(map[string]*scheduledTrigger),
		stopCh:   make(chan struct{}),
		updateCh: make(chan struct{}, 1),
	}
}

func (tc *Trigger) Start() error {
	tc.logger.Info("starting trigger coordinator")

	if err := tc.loadTriggersFromState(); err != nil {
		return fmt.Errorf("failed to load triggers from state: %w", err)
	}

	tc.wg.Add(1)
	go tc.run()

	return nil
}

// Stop gracefully stops the trigger coordinator
func (tc *Trigger) Stop() {
	tc.logger.Info("stopping trigger coordinator")
	close(tc.stopCh)
	tc.wg.Wait()
	tc.logger.Info("trigger coordinator stopped")
}

// AddTrigger adds a new trigger to the scheduler
func (tc *Trigger) CreateTrigger(trigger *state.Trigger) error {
	tc.lock.Lock()
	defer tc.lock.Unlock()

	cfg, err := decodeTriggerConfig(trigger)
	if err != nil {
		return err
	}

	for _, cron := range cfg.Crons {

		cronExpr, err := cronexpr.Parse(cron)
		if err != nil {
			return fmt.Errorf("failed to parse cron expression %q: %w", cron, err)
		}

		nextRun := cronExpr.Next(time.Now())

		st := &scheduledTrigger{
			trigger:  trigger,
			nextRun:  nextRun,
			cronExpr: cronExpr,
		}

		tc.triggers[tc.triggerKey(trigger)] = st
		heap.Push(tc.heap, st)

		tc.logger.Info("added trigger to scheduler",
			zap.String("trigger_id", trigger.ID),
			zap.String("namespace", trigger.Namespace),
			zap.String("flow_id", trigger.Flow),
			zap.Time("next_run", nextRun),
		)
	}

	// Signal the scheduler to update
	tc.signalUpdate()

	return nil
}

// RemoveTrigger removes a trigger from the scheduler
func (tc *Trigger) DeleteTrigger(trigger *state.Trigger) error {
	tc.lock.Lock()
	defer tc.lock.Unlock()

	key := tc.makeKey(trigger.ID, trigger.Namespace)
	st, exists := tc.triggers[key]
	if !exists {
		return fmt.Errorf("trigger %s not found", trigger.ID)
	}

	// Remove from heap
	if st.index >= 0 && st.index < tc.heap.Len() {
		heap.Remove(tc.heap, st.index)
	}

	delete(tc.triggers, key)

	tc.logger.Info("removed trigger from scheduler",
		zap.String("trigger_id", trigger.ID),
		zap.String("namespace", trigger.Namespace),
	)

	// Signal the scheduler to update
	tc.signalUpdate()

	return nil
}

// UpdateTrigger updates an existing trigger in the scheduler
func (tc *Trigger) UpdateTrigger(trigger *state.Trigger) error {
	// Remove and re-add to update the schedule
	key := tc.triggerKey(trigger)

	tc.lock.Lock()
	st, exists := tc.triggers[key]
	if exists && st.index >= 0 && st.index < tc.heap.Len() {
		heap.Remove(tc.heap, st.index)
		delete(tc.triggers, key)
	}
	tc.lock.Unlock()

	return tc.CreateTrigger(trigger)
}

// run is the main scheduling loop
func (tc *Trigger) run() {
	defer tc.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-tc.stopCh:
			return
		case <-ticker.C:
			tc.checkAndExecuteTriggers()
		case <-tc.updateCh:
			tc.checkAndExecuteTriggers()
		}
	}
}

// checkAndExecuteTriggers checks if any triggers are ready to execute
func (tc *Trigger) checkAndExecuteTriggers() {
	now := time.Now()

	tc.lock.Lock()
	defer tc.lock.Unlock()

	// Process all triggers that are due
	for tc.heap.Len() > 0 {
		st := (*tc.heap)[0]

		// If the next trigger is in the future, we're done
		if st.nextRun.After(now) {
			break
		}

		// Pop the trigger from the heap
		heap.Pop(tc.heap)

		// Execute the trigger in a goroutine
		tc.executeTrigger(st)

		// Calculate next run time and re-add to heap
		st.nextRun = st.cronExpr.Next(now)
		heap.Push(tc.heap, st)

		tc.logger.Debug("rescheduled trigger",
			zap.String("trigger_id", st.trigger.ID),
			zap.String("namespace", st.trigger.Namespace),
			zap.Time("next_run", st.nextRun),
		)
	}
}

// executeTrigger executes a trigger by running its associated flow
func (tc *Trigger) executeTrigger(st *scheduledTrigger) {
	tc.logger.Info("executing trigger",
		zap.String("trigger_id", st.trigger.ID),
		zap.String("namespace", st.trigger.Namespace),
		zap.String("flow_id", st.trigger.Flow),
		zap.Time("scheduled_time", st.nextRun),
	)

	// Execute flow in a separate goroutine to avoid blocking
	go func(trigger *state.Trigger) {
		if err := tc.runFn(trigger.Flow, trigger.Namespace, trigger.ID, nil); err != nil {
			tc.logger.Error("failed to execute triggered flow",
				zap.String("trigger_id", trigger.ID),
				zap.String("namespace", trigger.Namespace),
				zap.String("flow_id", trigger.Flow),
				zap.Error(err),
			)
		} else {
			tc.logger.Info("successfully triggered flow",
				zap.String("trigger_id", trigger.ID),
				zap.String("namespace", trigger.Namespace),
				zap.String("flow_id", trigger.Flow),
			)
		}
	}(st.trigger)
}

// signalUpdate signals the scheduler to check for updates
func (tc *Trigger) signalUpdate() {
	select {
	case tc.updateCh <- struct{}{}:
	default:
		// Channel already has a pending update signal
	}
}

// triggerKey creates a composite key for a trigger
func (tc *Trigger) triggerKey(trigger *state.Trigger) string {
	return tc.makeKey(trigger.ID, trigger.Namespace)
}

// makeKey creates a composite key from ID and namespace
func (tc *Trigger) makeKey(id, namespace string) string {
	return fmt.Sprintf("%s/%s", namespace, id)
}

// GetNextRun returns the next scheduled run time for a trigger
func (tc *Trigger) GetNextRun(triggerID, namespace string) (time.Time, error) {
	tc.lock.RLock()
	defer tc.lock.RUnlock()

	key := tc.makeKey(triggerID, namespace)
	st, exists := tc.triggers[key]
	if !exists {
		return time.Time{}, fmt.Errorf("trigger %s not found", triggerID)
	}

	return st.nextRun, nil
}

// loadTriggersFromState loads all cron triggers from the state backend
func (tc *Trigger) loadTriggersFromState() error {
	tc.logger.Info("loading triggers from state backend")

	// List all triggers from state (using "*" to get all namespaces)
	listResp, errResp := tc.state.Triggers().List(&serverstate.TriggersListReq{
		Namespace: "*",
	})
	if errResp != nil {
		return fmt.Errorf("failed to list triggers: %w", errResp.Err())
	}

	// Track how many triggers we successfully loaded
	loadedCount := 0
	errorCount := 0

	for _, triggerStub := range listResp.Triggers {
		getResp, errResp := tc.state.Triggers().Get(&serverstate.TriggersGetReq{
			ID:        triggerStub.ID,
			Namespace: triggerStub.Namespace,
		})
		if errResp != nil {
			tc.logger.Warn("failed to get trigger details",
				zap.String("trigger_id", triggerStub.ID),
				zap.String("namespace", triggerStub.Namespace),
				zap.Error(errResp.Err()))
			errorCount++
			continue
		}

		// Only load cron triggers
		if getResp.Trigger.Source.Provider != "cron" {
			continue
		}

		// Add the trigger to the scheduler
		if err := tc.CreateTrigger(getResp.Trigger); err != nil {
			tc.logger.Warn("failed to add trigger to scheduler",
				zap.String("trigger_id", getResp.Trigger.ID),
				zap.String("namespace", getResp.Trigger.Namespace),
				zap.Error(err))
			errorCount++
			continue
		}

		loadedCount++
	}

	tc.logger.Info("finished loading triggers from state",
		zap.Int("loaded", loadedCount),
		zap.Int("errors", errorCount),
		zap.Int("total", len(listResp.Triggers)))

	return nil
}
