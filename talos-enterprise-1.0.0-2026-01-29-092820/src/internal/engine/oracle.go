package engine

import (
	"context"
	"database/sql"
	"math"
	"time"
)

// Oracle is the predictive engine
type Oracle struct {
	db *sql.DB
}

type Forecast struct {
	Timestamp time.Time
	Value     float64
	Low       float64 // Confidence interval lower bound
	High      float64 // Confidence interval upper bound
}

func NewOracle(db *sql.DB) *Oracle {
	return &Oracle{db: db}
}

// ForecastMetric predicts future values for a metric using simple linear regression (for now)
// In production, this would interface with a more complex model or external ML service via Python SDK
func (o *Oracle) ForecastMetric(ctx context.Context, metricName string, horizon time.Duration) (*Forecast, error) {
	// 1. Fetch historical data (last 24 hours)
	query := `SELECT time, value FROM metrics 
			  WHERE name = $1 AND time > NOW() - INTERVAL '24 hours' 
			  ORDER BY time ASC`

	rows, err := o.db.QueryContext(ctx, query, metricName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []struct {
		t time.Time
		v float64
	}

	for rows.Next() {
		var t time.Time
		var v float64
		if err := rows.Scan(&t, &v); err != nil {
			return nil, err
		}
		points = append(points, struct {
			t time.Time
			v float64
		}{t, v})
	}

	if len(points) < 2 {
		return nil, sql.ErrNoRows // Not enough data
	}

	// 2. Simple Linear Regression
	// x = time (unix seconds), y = value
	var sumX, sumY, sumXY, sumXX float64
	n := float64(len(points))

	startTime := points[0].t.Unix()

	for _, p := range points {
		x := float64(p.t.Unix() - startTime) // Normalize time
		y := p.v
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// 3. Project into future
	futureTime := time.Now().Add(horizon)
	futureX := float64(futureTime.Unix() - startTime)
	predictedValue := slope*futureX + intercept

	// 4. Calculate Confidence (Simplified Standard Error)
	// In a real implementation, calculate proper intervals
	stdErr := 0.0
	for _, p := range points {
		x := float64(p.t.Unix() - startTime)
		y := p.v
		pred := slope*x + intercept
		stdErr += math.Pow(y-pred, 2)
	}
	stdErr = math.Sqrt(stdErr / n)

	return &Forecast{
		Timestamp: futureTime,
		Value:     predictedValue,
		Low:       predictedValue - (1.96 * stdErr), // 95% CI
		High:      predictedValue + (1.96 * stdErr),
	}, nil
}

// PredictResourceStress uses forecasts to predict if a resource will be stressed
func (o *Oracle) PredictResourceStress(ctx context.Context, resourceID string) (bool, error) {
	// Predict CPU usage for next 1 hour
	cpuForecast, err := o.ForecastMetric(ctx, "cpu_usage:"+resourceID, 1*time.Hour)
	if err != nil {
		return false, err // Treat as no stress if can't predict
	}

	// Threshold check (e.g., > 80%)
	if cpuForecast.Value > 80.0 {
		return true, nil
	}

	// Check upper bound for risk aversion
	if cpuForecast.High > 90.0 {
		return true, nil // Potential spike
	}

	return false, nil
}
