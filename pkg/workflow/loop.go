package workflow

import (
	"fmt"
	"reflect"

	"github.com/ratlabs-io/go-agent-kit/pkg/constants"
)

// Loop represents a loop construct that executes an action repeatedly.
type Loop struct {
	name   string
	action Action
	
	// Loop configuration
	count         int
	whileFunc     Predicate
	untilFunc     Predicate
	items         interface{}
	isWhileLoop   bool
	isUntilLoop   bool
	isCountLoop   bool
	isIterLoop    bool
}

// NewLoop creates a basic counted loop that executes the action a specified number of times.
func NewLoop(name string, count int) *Loop {
	return &Loop{
		name:        name,
		count:       count,
		isCountLoop: true,
	}
}

// NewLoopWhile creates a conditional loop that continues while the condition returns true.
func NewLoopWhile(name string, conditionFunc Predicate) *Loop {
	return &Loop{
		name:        name,
		whileFunc:   conditionFunc,
		isWhileLoop: true,
	}
}

// NewLoopUntil creates a conditional loop that continues until the condition returns true.
func NewLoopUntil(name string, conditionFunc Predicate) *Loop {
	return &Loop{
		name:        name,
		untilFunc:   conditionFunc,
		isUntilLoop: true,
	}
}

// NewLoopOver creates an iterator loop that executes the action for each item in the collection.
// The items parameter can be a slice, array, or map.
func NewLoopOver(name string, items interface{}) *Loop {
	return &Loop{
		name:       name,
		items:      items,
		isIterLoop: true,
	}
}

// WithAction sets the action to be executed in the loop.
func (l *Loop) WithAction(action Action) *Loop {
	l.action = action
	return l
}

// Name returns the name of the loop.
func (l *Loop) Name() string {
	return l.name
}

// Run executes the loop with the given work context.
func (l *Loop) Run(wctx WorkContext) WorkReport {
	if l.action == nil {
		return NewFailedWorkReport(fmt.Errorf("loop %s: no action specified", l.name))
	}

	logger := wctx.Logger()
	logger.Debug("Starting loop", "type", constants.FlowTypeLoop, "name", l.name)

	var report WorkReport
	iteration := 1

	switch {
	case l.isCountLoop:
		report = l.runCountLoop(wctx, iteration)
	case l.isWhileLoop:
		report = l.runWhileLoop(wctx, iteration)
	case l.isUntilLoop:
		report = l.runUntilLoop(wctx, iteration)
	case l.isIterLoop:
		report = l.runIterLoop(wctx, iteration)
	default:
		return NewFailedWorkReport(fmt.Errorf("loop %s: invalid loop configuration", l.name))
	}

	logger.Debug("Completed loop", "type", constants.FlowTypeLoop, "name", l.name, "status", report.Status)
	return report
}

// runCountLoop executes a counted loop.
func (l *Loop) runCountLoop(wctx WorkContext, startIteration int) WorkReport {
	logger := wctx.Logger()
	
	for i := 0; i < l.count; i++ {
		iteration := startIteration + i
		
		// Set loop context
		wctx.Set(constants.KeyLoopIteration, iteration)
		
		logger.Debug("Loop executing iteration", "name", l.name, "iteration", iteration, "total", l.count)
		
		report := l.action.Run(wctx)
		if report.Status == StatusFailure {
			logger.Error("Loop iteration failed", "name", l.name, "iteration", iteration)
			return report
		}
	}
	
	return NewCompletedWorkReport()
}

// runWhileLoop executes a while loop.
func (l *Loop) runWhileLoop(wctx WorkContext, startIteration int) WorkReport {
	logger := wctx.Logger()
	iteration := startIteration
	
	for {
		shouldContinue, err := l.whileFunc(wctx)
		if err != nil {
			logger.Error("Loop while condition error", "name", l.name)
			return NewFailedWorkReport(err)
		}
		if !shouldContinue {
			break
		}
		// Set loop context
		wctx.Set(constants.KeyLoopIteration, iteration)
		
		logger.Debug("Loop executing while iteration", "name", l.name, "iteration", iteration)
		
		report := l.action.Run(wctx)
		if report.Status == StatusFailure {
			logger.Error("Loop while iteration failed", "name", l.name, "iteration", iteration)
			return report
		}
		
		iteration++
		
		// Safety check to prevent infinite loops
		if iteration > 10000 {
			return NewFailedWorkReport(fmt.Errorf("loop %s: exceeded maximum iterations (10000)", l.name))
		}
	}
	
	return NewCompletedWorkReport()
}

// runUntilLoop executes an until loop. 
func (l *Loop) runUntilLoop(wctx WorkContext, startIteration int) WorkReport {
	logger := wctx.Logger()
	iteration := startIteration
	
	for {
		shouldStop, err := l.untilFunc(wctx)
		if err != nil {
			logger.Error("Loop until condition error", "name", l.name)
			return NewFailedWorkReport(err)
		}
		if shouldStop {
			break
		}
		// Set loop context
		wctx.Set(constants.KeyLoopIteration, iteration)
		
		logger.Debug("Loop executing until iteration", "name", l.name, "iteration", iteration)
		
		report := l.action.Run(wctx)
		if report.Status == StatusFailure {
			logger.Error("Loop until iteration failed", "name", l.name, "iteration", iteration)
			return report
		}
		
		iteration++
		
		// Safety check to prevent infinite loops
		if iteration > 10000 {
			return NewFailedWorkReport(fmt.Errorf("loop %s: exceeded maximum iterations (10000)", l.name))
		}
	}
	
	return NewCompletedWorkReport()
}

// runIterLoop executes an iterator loop over a collection.
func (l *Loop) runIterLoop(wctx WorkContext, startIteration int) WorkReport {
	// Use reflection to handle different collection types
	itemsValue := reflect.ValueOf(l.items)
	
	switch itemsValue.Kind() {
	case reflect.Slice, reflect.Array:
		return l.runSliceLoop(wctx, itemsValue, startIteration)
	case reflect.Map:
		return l.runMapLoop(wctx, itemsValue, startIteration)
	default:
		return NewFailedWorkReport(fmt.Errorf("loop %s: unsupported collection type: %T", l.name, l.items))
	}
}

// runSliceLoop executes a loop over a slice or array.
func (l *Loop) runSliceLoop(wctx WorkContext, itemsValue reflect.Value, startIteration int) WorkReport {
	logger := wctx.Logger()
	
	for i := 0; i < itemsValue.Len(); i++ {
		iteration := startIteration + i
		item := itemsValue.Index(i).Interface()
		
		// Set loop context
		wctx.Set(constants.KeyLoopIteration, iteration)
		wctx.Set(constants.KeyCurrentIndex, i)
		wctx.Set(constants.KeyCurrentItem, item)
		
		logger.Debug("Loop executing slice iteration", "name", l.name, "iteration", iteration, "index", i)
		
		report := l.action.Run(wctx)
		if report.Status == StatusFailure {
			logger.Error("Loop slice iteration failed", "name", l.name, "iteration", iteration, "index", i)
			return report
		}
	}
	
	return NewCompletedWorkReport()
}

// runMapLoop executes a loop over a map.
func (l *Loop) runMapLoop(wctx WorkContext, itemsValue reflect.Value, startIteration int) WorkReport {
	logger := wctx.Logger()
	
	keys := itemsValue.MapKeys()
	for i, key := range keys {
		iteration := startIteration + i
		value := itemsValue.MapIndex(key).Interface()
		
		// For maps, we set the key as the index and value as the item
		wctx.Set(constants.KeyLoopIteration, iteration)
		wctx.Set(constants.KeyCurrentIndex, key.Interface())
		wctx.Set(constants.KeyCurrentItem, value)
		
		logger.Debug("Loop executing map iteration", "name", l.name, "iteration", iteration, "key", key.Interface())
		
		report := l.action.Run(wctx)
		if report.Status == StatusFailure {
			logger.Error("Loop map iteration failed", "name", l.name, "iteration", iteration, "key", key.Interface())
			return report
		}
	}
	
	return NewCompletedWorkReport()
}