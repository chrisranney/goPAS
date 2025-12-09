package output

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

// Format represents the output format type.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

// Formatter handles output formatting.
type Formatter struct {
	format Format
}

// NewFormatter creates a new Formatter with the specified format.
func NewFormatter(format Format) *Formatter {
	return &Formatter{format: format}
}

// SetFormat changes the output format.
func (f *Formatter) SetFormat(format Format) {
	f.format = format
}

// GetFormat returns the current output format.
func (f *Formatter) GetFormat() Format {
	return f.format
}

// Format outputs the data in the current format.
func (f *Formatter) Format(data interface{}) error {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(data)
	case FormatYAML:
		return f.formatYAML(data)
	default:
		return f.formatTable(data)
	}
}

// FormatSingle outputs a single item.
func (f *Formatter) FormatSingle(data interface{}) error {
	return f.Format(data)
}

func (f *Formatter) formatJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (f *Formatter) formatYAML(data interface{}) error {
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(data)
}

func (f *Formatter) formatTable(data interface{}) error {
	// Handle nil
	if data == nil {
		fmt.Println("No data")
		return nil
	}

	v := reflect.ValueOf(data)

	// Dereference pointer if needed
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			fmt.Println("No data")
			return nil
		}
		v = v.Elem()
	}

	// Handle slice/array
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		return f.formatSliceTable(v)
	}

	// Handle struct
	if v.Kind() == reflect.Struct {
		return f.formatStructTable(v)
	}

	// Handle map
	if v.Kind() == reflect.Map {
		return f.formatMapTable(v)
	}

	// Fallback to simple print
	fmt.Printf("%v\n", data)
	return nil
}

func (f *Formatter) formatSliceTable(v reflect.Value) error {
	if v.Len() == 0 {
		fmt.Println("No items found")
		return nil
	}

	// Get headers from first element
	first := v.Index(0)
	if first.Kind() == reflect.Ptr {
		first = first.Elem()
	}

	if first.Kind() != reflect.Struct {
		// Simple slice, just print values
		for i := 0; i < v.Len(); i++ {
			fmt.Printf("%v\n", v.Index(i).Interface())
		}
		return nil
	}

	headers, fields := getStructHeaders(first.Type())

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetBorder(true)
	table.SetRowLine(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)

	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		row := getStructRow(elem, fields)
		table.Append(row)
	}

	table.Render()
	return nil
}

func (f *Formatter) formatStructTable(v reflect.Value) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Field", "Value"})
	table.SetBorder(true)
	table.SetRowLine(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetColWidth(60)

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		name := getFieldDisplayName(field)
		value := formatFieldValue(v.Field(i))
		if value != "" && value != "0" && value != "<nil>" {
			table.Append([]string{name, value})
		}
	}

	table.Render()
	return nil
}

func (f *Formatter) formatMapTable(v reflect.Value) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Key", "Value"})
	table.SetBorder(true)
	table.SetRowLine(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)

	for _, key := range v.MapKeys() {
		value := v.MapIndex(key)
		table.Append([]string{fmt.Sprintf("%v", key.Interface()), formatFieldValue(value)})
	}

	table.Render()
	return nil
}

func getStructHeaders(t reflect.Type) ([]string, []int) {
	var headers []string
	var fields []int

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		// Skip complex nested types for table display
		kind := field.Type.Kind()
		if kind == reflect.Struct || kind == reflect.Slice || kind == reflect.Map {
			if field.Type.String() != "time.Time" {
				continue
			}
		}
		if kind == reflect.Ptr {
			continue
		}

		headers = append(headers, getFieldDisplayName(field))
		fields = append(fields, i)
	}

	return headers, fields
}

func getFieldDisplayName(field reflect.StructField) string {
	// Check for json tag
	if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
		parts := strings.Split(tag, ",")
		if parts[0] != "" {
			return strings.ToUpper(parts[0])
		}
	}
	return strings.ToUpper(toSnakeCase(field.Name))
}

func getStructRow(v reflect.Value, fields []int) []string {
	var row []string
	for _, i := range fields {
		row = append(row, formatFieldValue(v.Field(i)))
	}
	return row
}

func formatFieldValue(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}

	// Handle pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	// Handle special types
	switch val := v.Interface().(type) {
	case time.Time:
		if val.IsZero() {
			return ""
		}
		return val.Format("2006-01-02 15:04:05")
	case bool:
		if val {
			return "Yes"
		}
		return "No"
	case int64:
		// Could be a unix timestamp
		if val > 1000000000 && val < 2000000000 {
			t := time.Unix(val, 0)
			return t.Format("2006-01-02 15:04:05")
		}
		return fmt.Sprintf("%d", val)
	}

	// Handle slices
	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			return ""
		}
		var items []string
		for i := 0; i < v.Len() && i < 3; i++ {
			items = append(items, fmt.Sprintf("%v", v.Index(i).Interface()))
		}
		if v.Len() > 3 {
			items = append(items, "...")
		}
		return strings.Join(items, ", ")
	}

	// Handle maps
	if v.Kind() == reflect.Map {
		if v.Len() == 0 {
			return ""
		}
		return fmt.Sprintf("(%d items)", v.Len())
	}

	return fmt.Sprintf("%v", v.Interface())
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return string(result)
}

// PrintSuccess prints a success message.
func PrintSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", Success("✓"), msg)
}

// PrintError prints an error message.
func PrintError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", Error("✗"), msg)
}

// PrintWarning prints a warning message.
func PrintWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", Warning("!"), msg)
}

// PrintInfo prints an info message.
func PrintInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", Info("→"), msg)
}

// Table helper for creating custom tables
type Table struct {
	writer *tablewriter.Table
}

// NewTable creates a new table with headers.
func NewTable(headers ...string) *Table {
	t := &Table{
		writer: tablewriter.NewWriter(os.Stdout),
	}
	t.writer.SetHeader(headers)
	t.writer.SetBorder(true)
	t.writer.SetRowLine(false)
	t.writer.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	t.writer.SetAlignment(tablewriter.ALIGN_LEFT)
	t.writer.SetAutoWrapText(false)
	return t
}

// AddRow adds a row to the table.
func (t *Table) AddRow(values ...string) {
	t.writer.Append(values)
}

// Render renders the table to stdout.
func (t *Table) Render() {
	t.writer.Render()
}
