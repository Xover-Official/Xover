package engine

import (
	"fmt"
	"time"

	"github.com/project-atlas/atlas/internal/cloud"
	"github.com/project-atlas/atlas/internal/logger"
)

type Scheduler struct{}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// IsOffPeak returns true if current time is night (10PM-6AM) or weekend
func (s *Scheduler) IsOffPeak() bool {
	now := time.Now()
	hour := now.Hour()
	weekday := now.Weekday()

	if weekday == time.Saturday || weekday == time.Sunday {
		return true
	}

	if hour >= 22 || hour < 6 {
		return true
	}

	return false
}

func (s *Scheduler) GenerateSchedulePlan(res *cloud.ResourceV2) (*ActionPlan, error) {
	// 1. Check for 'Indie-Force' (Hyper-Aggressive)
	if s.IsIndieForceWindow() {
		if mode, ok := res.Tags["atlas:mode"]; ok && mode == "indie" {
			if crit, ok := res.Tags["atlas:critical"]; !ok || crit != "true" {
				logger.LogAction(logger.Architect, "IndieForce", "ENGAGED",
					fmt.Sprintf("Indie-Force shutdown active for '%s' (Non-critical window).", res.ID))

				return &ActionPlan{
					Type:        "FORCE_STOP",
					Description: "Indie-Force: Nightly kill-switch for non-critical dev resources.",
					RiskScore:   2.0,
					ImpactScore: 9.0,
					EstSavings:  res.CostPerMonth * 0.70, // Max savings
				}, nil
			}
		}
	}

	if s.IsOffPeak() {
		// Identify resources tagged for scheduling
		if schedule, ok := res.Tags["atlas:schedule"]; ok && schedule == "nightly" {
			logger.LogAction(logger.Architect, "ScheduleCheck", "OFF-PEAK",
				fmt.Sprintf("Resource %s is idle during off-peak hours.", res.ID))

			return &ActionPlan{
				Type:        "STOP_RESOURCE",
				Description: fmt.Sprintf("Stop %s during off-peak hours to save cost.", res.ID),
				RiskScore:   1.0,
				ImpactScore: 5.0,
				EstSavings:  res.CostPerMonth * 0.4,
			}, nil
		}
	}
	return nil, nil
}

// IsIndieForceWindow returns true during 12 AM - 6 AM (Deep Night)
func (s *Scheduler) IsIndieForceWindow() bool {
	hour := time.Now().Hour()
	return hour >= 0 && hour < 6
}
