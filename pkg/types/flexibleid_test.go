package types

import (
	"encoding/json"
	"testing"
)

func TestFlexibleID_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    FlexibleID
		wantErr bool
	}{
		{
			name: "string UUID",
			json: `"550e8400-e29b-41d4-a716-446655440000"`,
			want: FlexibleID("550e8400-e29b-41d4-a716-446655440000"),
		},
		{
			name: "string number",
			json: `"12345"`,
			want: FlexibleID("12345"),
		},
		{
			name: "integer",
			json: `12345`,
			want: FlexibleID("12345"),
		},
		{
			name: "large integer",
			json: `9223372036854775807`,
			want: FlexibleID("9223372036854775807"),
		},
		{
			name: "zero",
			json: `0`,
			want: FlexibleID("0"),
		},
		{
			name: "empty string",
			json: `""`,
			want: FlexibleID(""),
		},
		{
			name: "null",
			json: `null`,
			want: FlexibleID(""),
		},
		{
			name:    "boolean",
			json:    `true`,
			wantErr: true,
		},
		{
			name:    "object",
			json:    `{"id": "test"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got FlexibleID
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("FlexibleID.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("FlexibleID.UnmarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlexibleID_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		f    FlexibleID
		want string
	}{
		{
			name: "UUID string",
			f:    FlexibleID("550e8400-e29b-41d4-a716-446655440000"),
			want: `"550e8400-e29b-41d4-a716-446655440000"`,
		},
		{
			name: "number string",
			f:    FlexibleID("12345"),
			want: `"12345"`,
		},
		{
			name: "empty string",
			f:    FlexibleID(""),
			want: `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.f)
			if err != nil {
				t.Errorf("FlexibleID.MarshalJSON() error = %v", err)
				return
			}
			if string(got) != tt.want {
				t.Errorf("FlexibleID.MarshalJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestFlexibleID_String(t *testing.T) {
	tests := []struct {
		name string
		f    FlexibleID
		want string
	}{
		{
			name: "UUID",
			f:    FlexibleID("550e8400-e29b-41d4-a716-446655440000"),
			want: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name: "number",
			f:    FlexibleID("12345"),
			want: "12345",
		},
		{
			name: "empty",
			f:    FlexibleID(""),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.String(); got != tt.want {
				t.Errorf("FlexibleID.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlexibleID_InStruct(t *testing.T) {
	type TestStruct struct {
		ID   FlexibleID `json:"id"`
		Name string     `json:"name"`
	}

	tests := []struct {
		name    string
		json    string
		wantID  FlexibleID
		wantErr bool
	}{
		{
			name:   "string ID",
			json:   `{"id": "uuid-123", "name": "test"}`,
			wantID: FlexibleID("uuid-123"),
		},
		{
			name:   "integer ID",
			json:   `{"id": 12345, "name": "test"}`,
			wantID: FlexibleID("12345"),
		},
		{
			name:   "large integer ID",
			json:   `{"id": 9223372036854775807, "name": "test"}`,
			wantID: FlexibleID("9223372036854775807"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TestStruct
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal struct error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.wantID {
				t.Errorf("ID = %v, want %v", got.ID, tt.wantID)
			}
		})
	}
}

func TestFlexibleBool_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    FlexibleBool
		wantErr bool
	}{
		{
			name: "boolean true",
			json: `true`,
			want: FlexibleBool(true),
		},
		{
			name: "boolean false",
			json: `false`,
			want: FlexibleBool(false),
		},
		{
			name: "string true lowercase",
			json: `"true"`,
			want: FlexibleBool(true),
		},
		{
			name: "string false lowercase",
			json: `"false"`,
			want: FlexibleBool(false),
		},
		{
			name: "string True titlecase",
			json: `"True"`,
			want: FlexibleBool(true),
		},
		{
			name: "string False titlecase",
			json: `"False"`,
			want: FlexibleBool(false),
		},
		{
			name: "string TRUE uppercase",
			json: `"TRUE"`,
			want: FlexibleBool(true),
		},
		{
			name: "string FALSE uppercase",
			json: `"FALSE"`,
			want: FlexibleBool(false),
		},
		{
			name: "string 1",
			json: `"1"`,
			want: FlexibleBool(true),
		},
		{
			name: "string 0",
			json: `"0"`,
			want: FlexibleBool(false),
		},
		{
			name: "empty string",
			json: `""`,
			want: FlexibleBool(false),
		},
		{
			name:    "invalid string",
			json:    `"yes"`,
			wantErr: true,
		},
		{
			name:    "number",
			json:    `1`,
			wantErr: true,
		},
		{
			name:    "object",
			json:    `{"value": true}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got FlexibleBool
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("FlexibleBool.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("FlexibleBool.UnmarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlexibleBool_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		f    FlexibleBool
		want string
	}{
		{
			name: "true",
			f:    FlexibleBool(true),
			want: `true`,
		},
		{
			name: "false",
			f:    FlexibleBool(false),
			want: `false`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.f)
			if err != nil {
				t.Errorf("FlexibleBool.MarshalJSON() error = %v", err)
				return
			}
			if string(got) != tt.want {
				t.Errorf("FlexibleBool.MarshalJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestFlexibleBool_Bool(t *testing.T) {
	tests := []struct {
		name string
		f    FlexibleBool
		want bool
	}{
		{
			name: "true",
			f:    FlexibleBool(true),
			want: true,
		},
		{
			name: "false",
			f:    FlexibleBool(false),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Bool(); got != tt.want {
				t.Errorf("FlexibleBool.Bool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlexibleBool_InStruct(t *testing.T) {
	type TestStruct struct {
		Active FlexibleBool `json:"active"`
		Name   string       `json:"name"`
	}

	tests := []struct {
		name       string
		json       string
		wantActive FlexibleBool
		wantErr    bool
	}{
		{
			name:       "boolean true",
			json:       `{"active": true, "name": "test"}`,
			wantActive: FlexibleBool(true),
		},
		{
			name:       "boolean false",
			json:       `{"active": false, "name": "test"}`,
			wantActive: FlexibleBool(false),
		},
		{
			name:       "string true",
			json:       `{"active": "true", "name": "test"}`,
			wantActive: FlexibleBool(true),
		},
		{
			name:       "string false",
			json:       `{"active": "false", "name": "test"}`,
			wantActive: FlexibleBool(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TestStruct
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal struct error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Active != tt.wantActive {
				t.Errorf("Active = %v, want %v", got.Active, tt.wantActive)
			}
		})
	}
}
