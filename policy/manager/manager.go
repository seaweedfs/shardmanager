package manager

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/seaweedfs/shardmanager/policy"
	"github.com/seaweedfs/shardmanager/policy/engine"
)

// PolicyManager manages policy evaluation and execution
type PolicyManager struct {
	engine       *engine.Engine
	store        policy.PolicyStore
	evalInterval time.Duration
	eventChan    chan struct{}
	stopChan     chan struct{}
	wg           sync.WaitGroup
	mu           sync.RWMutex
	isRunning    bool
}

// NewPolicyManager creates a new policy manager
func NewPolicyManager(
	metricProvider engine.MetricProvider,
	actionExecutor engine.ActionExecutor,
	store policy.PolicyStore,
	evalInterval time.Duration,
) *PolicyManager {
	return &PolicyManager{
		engine:       engine.NewEngine(metricProvider, actionExecutor),
		store:        store,
		evalInterval: evalInterval,
		eventChan:    make(chan struct{}, 1),
		stopChan:     make(chan struct{}),
	}
}

// Start begins periodic policy evaluation
func (pm *PolicyManager) Start(ctx context.Context) error {
	pm.mu.Lock()
	if pm.isRunning {
		pm.mu.Unlock()
		return nil
	}
	pm.isRunning = true
	pm.mu.Unlock()

	pm.wg.Add(1)
	go pm.run(ctx)
	return nil
}

// Stop halts policy evaluation
func (pm *PolicyManager) Stop() {
	pm.mu.Lock()
	if !pm.isRunning {
		pm.mu.Unlock()
		return
	}
	pm.isRunning = false
	pm.mu.Unlock()

	close(pm.stopChan)
	pm.wg.Wait()
}

// TriggerEvaluation triggers an immediate policy evaluation
func (pm *PolicyManager) TriggerEvaluation() {
	select {
	case pm.eventChan <- struct{}{}:
	default:
		// Channel is full, evaluation already triggered
	}
}

// run handles periodic and event-driven policy evaluation
func (pm *PolicyManager) run(ctx context.Context) {
	defer pm.wg.Done()

	ticker := time.NewTicker(pm.evalInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopChan:
			return
		case <-ticker.C:
			pm.evaluatePolicies(ctx)
		case <-pm.eventChan:
			pm.evaluatePolicies(ctx)
		}
	}
}

// evaluatePolicies retrieves and evaluates all policies
func (pm *PolicyManager) evaluatePolicies(ctx context.Context) {
	policies, err := pm.store.List(ctx)
	if err != nil {
		log.Printf("Failed to list policies: %v", err)
		return
	}

	if err := pm.engine.EvaluatePolicies(ctx, policies); err != nil {
		log.Printf("Failed to evaluate policies: %v", err)
	}
}
