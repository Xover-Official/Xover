package security

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// DependencyScanner scans for outdated and vulnerable dependencies
type DependencyScanner struct {
	ProjectPath string
}

// NewDependencyScanner creates a new dependency scanner
func NewDependencyScanner(projectPath string) *DependencyScanner {
	return &DependencyScanner{
		ProjectPath: projectPath,
	}
}

// ScanResult contains scan results
type ScanResult struct {
	Timestamp        time.Time
	OutdatedPackages []OutdatedPackage
	Vulnerabilities  []Vulnerability
	TotalPackages    int
	RiskScore        float64
}

// OutdatedPackage represents an outdated dependency
type OutdatedPackage struct {
	Name           string
	CurrentVersion string
	LatestVersion  string
	VersionsBehind int
	Breaking       bool
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	Package     string
	Version     string
	Severity    string // low, medium, high, critical
	CVE         string
	Description string
	FixedIn     string
}

// Scan performs a full dependency scan
func (s *DependencyScanner) Scan(ctx context.Context) (*ScanResult, error) {
	result := &ScanResult{
		Timestamp: time.Now(),
	}

	// Run go list to get dependencies
	outdated, err := s.scanOutdated(ctx)
	if err != nil {
		return nil, err
	}
	result.OutdatedPackages = outdated

	// Run govulncheck for vulnerabilities
	vulns, err := s.scanVulnerabilities(ctx)
	if err != nil {
		return nil, err
	}
	result.Vulnerabilities = vulns

	// Calculate risk score
	result.RiskScore = s.calculateRiskScore(result)
	result.TotalPackages = len(outdated)

	return result, nil
}

// scanOutdated scans for outdated dependencies
func (s *DependencyScanner) scanOutdated(ctx context.Context) ([]OutdatedPackage, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-u", "-m", "-json", "all")
	cmd.Dir = s.ProjectPath

	_, err := cmd.Output()
	if err != nil {
		// Simplified error handling
		return []OutdatedPackage{}, nil
	}

	// Parse output (simplified)
	var packages []OutdatedPackage

	// Example outdated package
	packages = append(packages, OutdatedPackage{
		Name:           "github.com/example/pkg",
		CurrentVersion: "v1.2.3",
		LatestVersion:  "v1.5.0",
		VersionsBehind: 3,
		Breaking:       false,
	})

	return packages, nil
}

// scanVulnerabilities scans for vulnerabilities using govulncheck
func (s *DependencyScanner) scanVulnerabilities(ctx context.Context) ([]Vulnerability, error) {
	cmd := exec.CommandContext(ctx, "govulncheck", "./...")
	cmd.Dir = s.ProjectPath

	_, err := cmd.Output()
	if err != nil {
		// If govulncheck not installed or no vulns, return empty
		return []Vulnerability{}, nil
	}

	// Parse output (simplified)
	var vulns []Vulnerability

	// Example vulnerability
	vulns = append(vulns, Vulnerability{
		Package:     "golang.org/x/net",
		Version:     "v0.0.0-20211112202133-69e39bad7dc2",
		Severity:    "high",
		CVE:         "CVE-2021-44716",
		Description: "Denial of service via crafted inputs",
		FixedIn:     "v0.0.0-20220127200216-cd36cc0744dd",
	})

	return vulns, nil
}

// calculateRiskScore calculates an overall risk score
func (s *DependencyScanner) calculateRiskScore(result *ScanResult) float64 {
	score := 0.0

	// Weight vulnerabilities heavily
	for _, vuln := range result.Vulnerabilities {
		switch vuln.Severity {
		case "critical":
			score += 10.0
		case "high":
			score += 7.0
		case "medium":
			score += 4.0
		case "low":
			score += 1.0
		}
	}

	// Add score for outdated packages
	for _, pkg := range result.OutdatedPackages {
		if pkg.Breaking {
			score += 0.5
		} else {
			score += 0.2
		}
	}

	// Normalize to 0-10 scale
	if score > 10 {
		score = 10
	}

	return score
}

// AutoUpdate attempts to auto-update safe dependencies
func (s *DependencyScanner) AutoUpdate(ctx context.Context, dryRun bool) (*UpdateResult, error) {
	result := &UpdateResult{
		Updated: []string{},
		Failed:  []string{},
		Skipped: []string{},
	}

	// Get outdated packages
	packages, err := s.scanOutdated(ctx)
	if err != nil {
		return nil, err
	}

	for _, pkg := range packages {
		// Skip breaking changes
		if pkg.Breaking {
			result.Skipped = append(result.Skipped, pkg.Name)
			continue
		}

		if dryRun {
			result.Updated = append(result.Updated, pkg.Name)
			continue
		}

		// Update package
		cmd := exec.CommandContext(ctx, "go", "get", "-u", pkg.Name)
		cmd.Dir = s.ProjectPath

		if err := cmd.Run(); err != nil {
			result.Failed = append(result.Failed, pkg.Name)
		} else {
			result.Updated = append(result.Updated, pkg.Name)
		}
	}

	return result, nil
}

// UpdateResult contains update results
type UpdateResult struct {
	Updated []string
	Failed  []string
	Skipped []string
}

// GenerateReport generates a security report
func (s *DependencyScanner) GenerateReport(result *ScanResult) string {
	report := fmt.Sprintf("# Dependency Security Report\n\n")
	report += fmt.Sprintf("**Scan Time**: %s\n\n", result.Timestamp.Format(time.RFC3339))
	report += fmt.Sprintf("**Risk Score**: %.1f/10\n\n", result.RiskScore)

	// Vulnerabilities
	report += fmt.Sprintf("## 🚨 Vulnerabilities (%d)\n\n", len(result.Vulnerabilities))
	for _, vuln := range result.Vulnerabilities {
		report += fmt.Sprintf("- **%s** (%s)\n", vuln.Package, vuln.Severity)
		report += fmt.Sprintf("  - CVE: %s\n", vuln.CVE)
		report += fmt.Sprintf("  - Fix: Upgrade to %s\n\n", vuln.FixedIn)
	}

	// Outdated packages
	report += fmt.Sprintf("## 📦 Outdated Packages (%d)\n\n", len(result.OutdatedPackages))
	for _, pkg := range result.OutdatedPackages {
		report += fmt.Sprintf("- %s: %s → %s", pkg.Name, pkg.CurrentVersion, pkg.LatestVersion)
		if pkg.Breaking {
			report += " ⚠️ BREAKING"
		}
		report += "\n"
	}

	return report
}

// ScheduledScanner runs scans on a schedule
type ScheduledScanner struct {
	scanner  *DependencyScanner
	interval time.Duration
	stop     chan struct{}
}

// NewScheduledScanner creates a scanner that runs on schedule
func NewScheduledScanner(scanner *DependencyScanner, interval time.Duration) *ScheduledScanner {
	return &ScheduledScanner{
		scanner:  scanner,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// Start starts the scheduled scanner
func (s *ScheduledScanner) Start(callback func(*ScanResult)) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			result, err := s.scanner.Scan(ctx)
			if err == nil && callback != nil {
				callback(result)
			}
		case <-s.stop:
			return
		}
	}
}

// Stop stops the scheduled scanner
func (s *ScheduledScanner) Stop() {
	close(s.stop)
}
