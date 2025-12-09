// Package reports provides tests for reporting functionality.
package reports

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chrisranney/gopas/internal/client"
	"github.com/chrisranney/gopas/internal/session"
)

// createTestSession creates a test session with a mock server
func createTestSession(t *testing.T, handler http.Handler) (*session.Session, *httptest.Server) {
	server := httptest.NewServer(handler)

	sess, err := session.NewSession(server.URL)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	sess.Client = createTestClient(t, server.URL)
	sess.SetAuthenticated("testuser", "test-token", "CyberArk")

	return sess, server
}

// createTestClient creates a test client with mock server URL
func createTestClient(t *testing.T, serverURL string) *client.Client {
	c, err := client.NewClient(client.Config{BaseURL: serverURL})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	c.SetAuthToken("test-token")
	return c
}

func TestListReports(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse []Report
		serverStatus   int
		wantErr        bool
		wantCount      int
	}{
		{
			name: "successful list",
			serverResponse: []Report{
				{ID: "report-1", Name: "Account Activity", Type: "Activity"},
				{ID: "report-2", Name: "Safe Audit", Type: "Audit"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantCount:    2,
		},
		{
			name:         "server error",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					response := ReportsResponse{Reports: tt.serverResponse}
					json.NewEncoder(w).Encode(response)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ListReports(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("ListReports() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListReports() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("ListReports() returned %d reports, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestGetReport(t *testing.T) {
	tests := []struct {
		name           string
		reportID       string
		serverResponse *Report
		serverStatus   int
		wantErr        bool
	}{
		{
			name:     "successful get",
			reportID: "report-1",
			serverResponse: &Report{
				ID:          "report-1",
				Name:        "Account Activity",
				Description: "Shows account activity",
				Type:        "Activity",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:     "empty report ID",
			reportID: "",
			wantErr:  true,
		},
		{
			name:         "not found",
			reportID:     "nonexistent",
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := GetReport(context.Background(), sess, tt.reportID)
			if tt.wantErr {
				if err == nil {
					t.Error("GetReport() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetReport() unexpected error: %v", err)
				return
			}

			if result.ID != tt.serverResponse.ID {
				t.Errorf("GetReport().ID = %v, want %v", result.ID, tt.serverResponse.ID)
			}
		})
	}
}

func TestExportReport(t *testing.T) {
	tests := []struct {
		name         string
		opts         ExportReportOptions
		serverData   []byte
		serverStatus int
		wantErr      bool
	}{
		{
			name: "successful export CSV",
			opts: ExportReportOptions{
				ReportID: "report-1",
				Format:   "CSV",
			},
			serverData:   []byte("header1,header2\nvalue1,value2"),
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "successful export with dates",
			opts: ExportReportOptions{
				ReportID: "report-1",
				Format:   "PDF",
				FromDate: 1705315800,
				ToDate:   1705402200,
			},
			serverData:   []byte("%PDF-1.4..."),
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "empty report ID",
			opts: ExportReportOptions{
				Format: "CSV",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/octet-stream")
				w.WriteHeader(tt.serverStatus)
				if tt.serverData != nil {
					w.Write(tt.serverData)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ExportReport(context.Background(), sess, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("ExportReport() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ExportReport() unexpected error: %v", err)
				return
			}

			if len(result.Data) == 0 {
				t.Error("ExportReport().Data is empty")
			}
		})
	}
}

func TestListReportSchedules(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse []ReportSchedule
		serverStatus   int
		wantErr        bool
		wantCount      int
	}{
		{
			name: "successful list",
			serverResponse: []ReportSchedule{
				{ID: "sched-1", ReportID: "report-1", Frequency: "Daily", Enabled: true},
				{ID: "sched-2", ReportID: "report-2", Frequency: "Weekly", Enabled: false},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantCount:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					response := ReportSchedulesResponse{Schedules: tt.serverResponse}
					json.NewEncoder(w).Encode(response)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := ListReportSchedules(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("ListReportSchedules() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ListReportSchedules() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("ListReportSchedules() returned %d schedules, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestCreateReportSchedule(t *testing.T) {
	tests := []struct {
		name           string
		opts           CreateReportScheduleOptions
		serverResponse *ReportSchedule
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful create",
			opts: CreateReportScheduleOptions{
				ReportID:  "report-1",
				Frequency: "Daily",
				StartTime: "08:00",
				Format:    "CSV",
				Enabled:   true,
			},
			serverResponse: &ReportSchedule{
				ID:        "sched-new",
				ReportID:  "report-1",
				Frequency: "Daily",
				Enabled:   true,
			},
			serverStatus: http.StatusCreated,
			wantErr:      false,
		},
		{
			name: "missing report ID",
			opts: CreateReportScheduleOptions{
				Frequency: "Daily",
			},
			wantErr: true,
		},
		{
			name: "missing frequency",
			opts: CreateReportScheduleOptions{
				ReportID: "report-1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := CreateReportSchedule(context.Background(), sess, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("CreateReportSchedule() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateReportSchedule() unexpected error: %v", err)
				return
			}

			if result.ID != tt.serverResponse.ID {
				t.Errorf("CreateReportSchedule().ID = %v, want %v", result.ID, tt.serverResponse.ID)
			}
		})
	}
}

func TestDeleteReportSchedule(t *testing.T) {
	tests := []struct {
		name         string
		scheduleID   string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful delete",
			scheduleID:   "sched-1",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:       "empty schedule ID",
			scheduleID: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}
				w.WriteHeader(tt.serverStatus)
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			err := DeleteReportSchedule(context.Background(), sess, tt.scheduleID)
			if tt.wantErr {
				if err == nil {
					t.Error("DeleteReportSchedule() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteReportSchedule() unexpected error: %v", err)
			}
		})
	}
}

func TestGetUserLicenseReport(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse *UserLicenseReport
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "successful get",
			serverResponse: &UserLicenseReport{
				TotalUsers:        100,
				LicensedUsers:     80,
				UnlicensedUsers:   20,
				UsersWithAccess:   75,
				UsersWithoutAccess: 25,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "server error",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			})

			sess, server := createTestSession(t, handler)
			defer server.Close()

			result, err := GetUserLicenseReport(context.Background(), sess)
			if tt.wantErr {
				if err == nil {
					t.Error("GetUserLicenseReport() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetUserLicenseReport() unexpected error: %v", err)
				return
			}

			if result.TotalUsers != tt.serverResponse.TotalUsers {
				t.Errorf("GetUserLicenseReport().TotalUsers = %v, want %v", result.TotalUsers, tt.serverResponse.TotalUsers)
			}
		})
	}
}

func TestListReports_InvalidSession(t *testing.T) {
	_, err := ListReports(context.Background(), nil)
	if err == nil {
		t.Error("ListReports() with nil session expected error, got nil")
	}
}
