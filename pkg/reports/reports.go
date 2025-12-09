// Package reports provides CyberArk reporting functionality.
// This is equivalent to the report functions in psPAS including
// Get-PASReport, Export-PASReport, Get-PASReportSchedule, New-PASReportSchedule.
package reports

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chrisranney/gopas/internal/session"
)

// Report represents a CyberArk report definition.
type Report struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"`
	Category    string `json:"category,omitempty"`
}

// ReportsResponse represents the response from listing reports.
type ReportsResponse struct {
	Reports []Report `json:"Reports"`
}

// ListReports retrieves available reports.
// This is equivalent to Get-PASReport in psPAS.
func ListReports(ctx context.Context, sess *session.Session) ([]Report, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/Reports", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}

	var result ReportsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse reports response: %w", err)
	}

	return result.Reports, nil
}

// GetReport retrieves a specific report.
func GetReport(ctx context.Context, sess *session.Session, reportID string) (*Report, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if reportID == "" {
		return nil, fmt.Errorf("reportID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Reports/%s", url.PathEscape(reportID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	var result Report
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse report response: %w", err)
	}

	return &result, nil
}

// ExportReportOptions holds options for exporting a report.
type ExportReportOptions struct {
	ReportID   string            `json:"-"` // Used in URL
	Format     string            `json:"format,omitempty"` // CSV, PDF, HTML
	Parameters map[string]string `json:"parameters,omitempty"`
	FromDate   int64             `json:"fromDate,omitempty"`
	ToDate     int64             `json:"toDate,omitempty"`
}

// ReportData represents exported report data.
type ReportData struct {
	ReportID    string `json:"reportId"`
	ReportName  string `json:"reportName"`
	Format      string `json:"format"`
	Data        []byte `json:"data"`
	ContentType string `json:"contentType"`
}

// ExportReport exports a report in the specified format.
// This is equivalent to Export-PASReport in psPAS.
func ExportReport(ctx context.Context, sess *session.Session, opts ExportReportOptions) (*ReportData, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if opts.ReportID == "" {
		return nil, fmt.Errorf("reportID is required")
	}

	// Default format to CSV
	if opts.Format == "" {
		opts.Format = "CSV"
	}

	params := url.Values{}
	params.Set("format", opts.Format)
	if opts.FromDate > 0 {
		params.Set("fromDate", fmt.Sprintf("%d", opts.FromDate))
	}
	if opts.ToDate > 0 {
		params.Set("toDate", fmt.Sprintf("%d", opts.ToDate))
	}
	for k, v := range opts.Parameters {
		params.Set(k, v)
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Reports/%s/Export", url.PathEscape(opts.ReportID)), params)
	if err != nil {
		return nil, fmt.Errorf("failed to export report: %w", err)
	}

	return &ReportData{
		ReportID:    opts.ReportID,
		Format:      opts.Format,
		Data:        resp.Body,
		ContentType: "application/octet-stream", // Will be set by response headers in real implementation
	}, nil
}

// ReportSchedule represents a scheduled report.
type ReportSchedule struct {
	ID              string   `json:"id"`
	ReportID        string   `json:"reportId"`
	ReportName      string   `json:"reportName,omitempty"`
	Frequency       string   `json:"frequency"` // Daily, Weekly, Monthly
	StartTime       string   `json:"startTime,omitempty"`
	DayOfWeek       int      `json:"dayOfWeek,omitempty"` // 0-6 for weekly
	DayOfMonth      int      `json:"dayOfMonth,omitempty"` // 1-31 for monthly
	Format          string   `json:"format,omitempty"` // CSV, PDF, HTML
	Recipients      []string `json:"recipients,omitempty"`
	Enabled         bool     `json:"enabled"`
	LastRunTime     int64    `json:"lastRunTime,omitempty"`
	NextRunTime     int64    `json:"nextRunTime,omitempty"`
}

// ReportSchedulesResponse represents the response from listing report schedules.
type ReportSchedulesResponse struct {
	Schedules []ReportSchedule `json:"Schedules"`
}

// ListReportSchedules retrieves all report schedules.
// This is equivalent to Get-PASReportSchedule in psPAS.
func ListReportSchedules(ctx context.Context, sess *session.Session) ([]ReportSchedule, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/Reports/Schedules", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list report schedules: %w", err)
	}

	var result ReportSchedulesResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse report schedules response: %w", err)
	}

	return result.Schedules, nil
}

// GetReportSchedule retrieves a specific report schedule.
func GetReportSchedule(ctx context.Context, sess *session.Session, scheduleID string) (*ReportSchedule, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if scheduleID == "" {
		return nil, fmt.Errorf("scheduleID is required")
	}

	resp, err := sess.Client.Get(ctx, fmt.Sprintf("/Reports/Schedules/%s", url.PathEscape(scheduleID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get report schedule: %w", err)
	}

	var result ReportSchedule
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse report schedule response: %w", err)
	}

	return &result, nil
}

// CreateReportScheduleOptions holds options for creating a report schedule.
type CreateReportScheduleOptions struct {
	ReportID    string   `json:"reportId"`
	Frequency   string   `json:"frequency"` // Daily, Weekly, Monthly
	StartTime   string   `json:"startTime,omitempty"`
	DayOfWeek   int      `json:"dayOfWeek,omitempty"` // 0-6 for weekly
	DayOfMonth  int      `json:"dayOfMonth,omitempty"` // 1-31 for monthly
	Format      string   `json:"format,omitempty"` // CSV, PDF, HTML
	Recipients  []string `json:"recipients,omitempty"`
	Enabled     bool     `json:"enabled"`
}

// CreateReportSchedule creates a new report schedule.
// This is equivalent to New-PASReportSchedule in psPAS.
func CreateReportSchedule(ctx context.Context, sess *session.Session, opts CreateReportScheduleOptions) (*ReportSchedule, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if opts.ReportID == "" {
		return nil, fmt.Errorf("reportID is required")
	}

	if opts.Frequency == "" {
		return nil, fmt.Errorf("frequency is required")
	}

	resp, err := sess.Client.Post(ctx, "/Reports/Schedules", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create report schedule: %w", err)
	}

	var result ReportSchedule
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse report schedule response: %w", err)
	}

	return &result, nil
}

// UpdateReportSchedule updates an existing report schedule.
func UpdateReportSchedule(ctx context.Context, sess *session.Session, scheduleID string, opts CreateReportScheduleOptions) (*ReportSchedule, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	if scheduleID == "" {
		return nil, fmt.Errorf("scheduleID is required")
	}

	resp, err := sess.Client.Put(ctx, fmt.Sprintf("/Reports/Schedules/%s", url.PathEscape(scheduleID)), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to update report schedule: %w", err)
	}

	var result ReportSchedule
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse report schedule response: %w", err)
	}

	return &result, nil
}

// DeleteReportSchedule deletes a report schedule.
func DeleteReportSchedule(ctx context.Context, sess *session.Session, scheduleID string) error {
	if sess == nil || !sess.IsValid() {
		return fmt.Errorf("valid session is required")
	}

	if scheduleID == "" {
		return fmt.Errorf("scheduleID is required")
	}

	_, err := sess.Client.Delete(ctx, fmt.Sprintf("/Reports/Schedules/%s", url.PathEscape(scheduleID)))
	if err != nil {
		return fmt.Errorf("failed to delete report schedule: %w", err)
	}

	return nil
}

// UserLicenseReport represents user license information.
type UserLicenseReport struct {
	TotalUsers        int `json:"totalUsers"`
	LicensedUsers     int `json:"licensedUsers"`
	UnlicensedUsers   int `json:"unlicensedUsers"`
	UsersWithAccess   int `json:"usersWithAccess"`
	UsersWithoutAccess int `json:"usersWithoutAccess"`
}

// GetUserLicenseReport retrieves user license information.
// This is equivalent to Get-PASUserLicenseReport in psPAS.
func GetUserLicenseReport(ctx context.Context, sess *session.Session) (*UserLicenseReport, error) {
	if sess == nil || !sess.IsValid() {
		return nil, fmt.Errorf("valid session is required")
	}

	resp, err := sess.Client.Get(ctx, "/Reports/UserLicense", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user license report: %w", err)
	}

	var result UserLicenseReport
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse user license report: %w", err)
	}

	return &result, nil
}
