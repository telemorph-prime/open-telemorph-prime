package promql

import (
	"math"
	"sort"
	"time"
)

// FunctionRegistry holds all available PromQL functions
type FunctionRegistry struct {
	functions map[string]Function
}

// Function represents a PromQL function
type Function struct {
	Name        string
	Description string
	Args        []ArgType
	ReturnType  ReturnType
	Handler     func(args []interface{}) (interface{}, error)
}

// ArgType represents function argument types
type ArgType int

const (
	ArgTypeInstantVector ArgType = iota
	ArgTypeRangeVector
	ArgTypeScalar
	ArgTypeString
)

// ReturnType represents function return types
type ReturnType int

const (
	ReturnTypeInstantVector ReturnType = iota
	ReturnTypeRangeVector
	ReturnTypeScalar
	ReturnTypeString
)

// NewFunctionRegistry creates a new function registry
func NewFunctionRegistry() *FunctionRegistry {
	registry := &FunctionRegistry{
		functions: make(map[string]Function),
	}
	registry.registerBuiltinFunctions()
	return registry
}

// registerBuiltinFunctions registers all built-in PromQL functions
func (fr *FunctionRegistry) registerBuiltinFunctions() {
	// Rate functions
	fr.Register(Function{
		Name:        "rate",
		Description: "Calculates the per-second average rate of increase",
		Args:        []ArgType{ArgTypeRangeVector},
		ReturnType:  ReturnTypeInstantVector,
		Handler:     fr.handleRate,
	})

	fr.Register(Function{
		Name:        "increase",
		Description: "Calculates the increase in the time series",
		Args:        []ArgType{ArgTypeRangeVector},
		ReturnType:  ReturnTypeInstantVector,
		Handler:     fr.handleIncrease,
	})

	// Aggregation functions
	fr.Register(Function{
		Name:        "sum",
		Description: "Sum over dimensions",
		Args:        []ArgType{ArgTypeInstantVector},
		ReturnType:  ReturnTypeInstantVector,
		Handler:     fr.handleSum,
	})

	fr.Register(Function{
		Name:        "avg",
		Description: "Average over dimensions",
		Args:        []ArgType{ArgTypeInstantVector},
		ReturnType:  ReturnTypeInstantVector,
		Handler:     fr.handleAvg,
	})

	fr.Register(Function{
		Name:        "count",
		Description: "Count of elements",
		Args:        []ArgType{ArgTypeInstantVector},
		ReturnType:  ReturnTypeInstantVector,
		Handler:     fr.handleCount,
	})

	fr.Register(Function{
		Name:        "min",
		Description: "Minimum value",
		Args:        []ArgType{ArgTypeInstantVector},
		ReturnType:  ReturnTypeInstantVector,
		Handler:     fr.handleMin,
	})

	fr.Register(Function{
		Name:        "max",
		Description: "Maximum value",
		Args:        []ArgType{ArgTypeInstantVector},
		ReturnType:  ReturnTypeInstantVector,
		Handler:     fr.handleMax,
	})

	// Math functions
	fr.Register(Function{
		Name:        "abs",
		Description: "Absolute value",
		Args:        []ArgType{ArgTypeInstantVector},
		ReturnType:  ReturnTypeInstantVector,
		Handler:     fr.handleAbs,
	})

	fr.Register(Function{
		Name:        "ceil",
		Description: "Round up to nearest integer",
		Args:        []ArgType{ArgTypeInstantVector},
		ReturnType:  ReturnTypeInstantVector,
		Handler:     fr.handleCeil,
	})

	fr.Register(Function{
		Name:        "floor",
		Description: "Round down to nearest integer",
		Args:        []ArgType{ArgTypeInstantVector},
		ReturnType:  ReturnTypeInstantVector,
		Handler:     fr.handleFloor,
	})

	fr.Register(Function{
		Name:        "round",
		Description: "Round to nearest integer",
		Args:        []ArgType{ArgTypeInstantVector},
		ReturnType:  ReturnTypeInstantVector,
		Handler:     fr.handleRound,
	})

	// Time functions
	fr.Register(Function{
		Name:        "time",
		Description: "Current Unix timestamp",
		Args:        []ArgType{},
		ReturnType:  ReturnTypeScalar,
		Handler:     fr.handleTime,
	})

	fr.Register(Function{
		Name:        "timestamp",
		Description: "Unix timestamp of the sample",
		Args:        []ArgType{ArgTypeInstantVector},
		ReturnType:  ReturnTypeInstantVector,
		Handler:     fr.handleTimestamp,
	})
}

// Register registers a new function
func (fr *FunctionRegistry) Register(fn Function) {
	fr.functions[fn.Name] = fn
}

// Get retrieves a function by name
func (fr *FunctionRegistry) Get(name string) (Function, bool) {
	fn, exists := fr.functions[name]
	return fn, exists
}

// List returns all registered function names
func (fr *FunctionRegistry) List() []string {
	var names []string
	for name := range fr.functions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Function handlers

func (fr *FunctionRegistry) handleRate(args []interface{}) (interface{}, error) {
	// This would be implemented by the evaluator
	return nil, nil
}

func (fr *FunctionRegistry) handleIncrease(args []interface{}) (interface{}, error) {
	// This would be implemented by the evaluator
	return nil, nil
}

func (fr *FunctionRegistry) handleSum(args []interface{}) (interface{}, error) {
	// This would be implemented by the evaluator
	return nil, nil
}

func (fr *FunctionRegistry) handleAvg(args []interface{}) (interface{}, error) {
	// This would be implemented by the evaluator
	return nil, nil
}

func (fr *FunctionRegistry) handleCount(args []interface{}) (interface{}, error) {
	// This would be implemented by the evaluator
	return nil, nil
}

func (fr *FunctionRegistry) handleMin(args []interface{}) (interface{}, error) {
	// This would be implemented by the evaluator
	return nil, nil
}

func (fr *FunctionRegistry) handleMax(args []interface{}) (interface{}, error) {
	// This would be implemented by the evaluator
	return nil, nil
}

func (fr *FunctionRegistry) handleAbs(args []interface{}) (interface{}, error) {
	// This would be implemented by the evaluator
	return nil, nil
}

func (fr *FunctionRegistry) handleCeil(args []interface{}) (interface{}, error) {
	// This would be implemented by the evaluator
	return nil, nil
}

func (fr *FunctionRegistry) handleFloor(args []interface{}) (interface{}, error) {
	// This would be implemented by the evaluator
	return nil, nil
}

func (fr *FunctionRegistry) handleRound(args []interface{}) (interface{}, error) {
	// This would be implemented by the evaluator
	return nil, nil
}

func (fr *FunctionRegistry) handleTime(args []interface{}) (interface{}, error) {
	return float64(time.Now().Unix()), nil
}

func (fr *FunctionRegistry) handleTimestamp(args []interface{}) (interface{}, error) {
	// This would be implemented by the evaluator
	return nil, nil
}

// Utility functions for mathematical operations

// ApplyAbs applies absolute value to all points in a series
func ApplyAbs(series []MetricSeries) []MetricSeries {
	result := make([]MetricSeries, len(series))
	for i, s := range series {
		result[i] = MetricSeries{
			MetricName: s.MetricName,
			Labels:     s.Labels,
			Points:     make([]MetricPoint, len(s.Points)),
		}
		for j, point := range s.Points {
			result[i].Points[j] = MetricPoint{
				Timestamp: point.Timestamp,
				Value:     math.Abs(point.Value),
				Labels:    point.Labels,
			}
		}
	}
	return result
}

// ApplyCeil applies ceiling function to all points in a series
func ApplyCeil(series []MetricSeries) []MetricSeries {
	result := make([]MetricSeries, len(series))
	for i, s := range series {
		result[i] = MetricSeries{
			MetricName: s.MetricName,
			Labels:     s.Labels,
			Points:     make([]MetricPoint, len(s.Points)),
		}
		for j, point := range s.Points {
			result[i].Points[j] = MetricPoint{
				Timestamp: point.Timestamp,
				Value:     math.Ceil(point.Value),
				Labels:    point.Labels,
			}
		}
	}
	return result
}

// ApplyFloor applies floor function to all points in a series
func ApplyFloor(series []MetricSeries) []MetricSeries {
	result := make([]MetricSeries, len(series))
	for i, s := range series {
		result[i] = MetricSeries{
			MetricName: s.MetricName,
			Labels:     s.Labels,
			Points:     make([]MetricPoint, len(s.Points)),
		}
		for j, point := range s.Points {
			result[i].Points[j] = MetricPoint{
				Timestamp: point.Timestamp,
				Value:     math.Floor(point.Value),
				Labels:    point.Labels,
			}
		}
	}
	return result
}

// ApplyRound applies round function to all points in a series
func ApplyRound(series []MetricSeries) []MetricSeries {
	result := make([]MetricSeries, len(series))
	for i, s := range series {
		result[i] = MetricSeries{
			MetricName: s.MetricName,
			Labels:     s.Labels,
			Points:     make([]MetricPoint, len(s.Points)),
		}
		for j, point := range s.Points {
			result[i].Points[j] = MetricPoint{
				Timestamp: point.Timestamp,
				Value:     math.Round(point.Value),
				Labels:    point.Labels,
			}
		}
	}
	return result
}

// ApplyTimestamp extracts timestamps from all points in a series
func ApplyTimestamp(series []MetricSeries) []MetricSeries {
	result := make([]MetricSeries, len(series))
	for i, s := range series {
		result[i] = MetricSeries{
			MetricName: s.MetricName,
			Labels:     s.Labels,
			Points:     make([]MetricPoint, len(s.Points)),
		}
		for j, point := range s.Points {
			result[i].Points[j] = MetricPoint{
				Timestamp: point.Timestamp,
				Value:     float64(point.Timestamp.Unix()),
				Labels:    point.Labels,
			}
		}
	}
	return result
}

