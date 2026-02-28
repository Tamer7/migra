package engine

import (
	"time"

	"github.com/migra/migra/pkg/migra"
)

// SummarizeResults creates a summary of execution results
func SummarizeResults(results []migra.ServiceResult, totalDuration time.Duration) *Result {
	summary := &Result{
		Services: results,
		Duration: totalDuration,
	}

	for _, r := range results {
		if r.Success {
			summary.TotalSuccess++
		} else {
			summary.TotalFailure++
		}
	}

	return summary
}
