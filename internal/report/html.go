package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cjp2600/stepwise/internal/workflow"
)

// HTMLReportData represents the data structure for HTML report
type HTMLReportData struct {
	Title         string
	GeneratedAt   string
	WorkflowName  string
	TotalTests    int
	PassedTests   int
	FailedTests   int
	SkippedTests  int
	TotalDuration time.Duration
	SuccessRate   float64
	Results       []workflow.TestResult
	WorkflowFile  string
}

// GenerateHTMLReport generates a self-contained HTML report from test results
func GenerateHTMLReport(results []workflow.TestResult, workflowName string, workflowFile string, outputPath string) error {
	// Calculate statistics
	passed := 0
	failed := 0
	skipped := 0
	totalDuration := time.Duration(0)

	for _, result := range results {
		if result.Status == "passed" {
			passed++
		} else if result.Status == "failed" {
			failed++
		} else if result.Status == "skipped" {
			skipped++
		}
		totalDuration += result.Duration
	}

	total := len(results)
	successRate := 0.0
	if total > 0 {
		successRate = float64(passed) / float64(total) * 100
	}

	// Prepare data
	data := HTMLReportData{
		Title:         "Stepwise Test Report",
		GeneratedAt:   time.Now().Format("2006-01-02 15:04:05"),
		WorkflowName:  workflowName,
		TotalTests:    total,
		PassedTests:   passed,
		FailedTests:   failed,
		SkippedTests:  skipped,
		TotalDuration: totalDuration,
		SuccessRate:   successRate,
		Results:       results,
		WorkflowFile:  workflowFile,
	}

	// Generate HTML
	htmlContent, err := generateHTML(data)
	if err != nil {
		return fmt.Errorf("failed to generate HTML: %w", err)
	}

	// Determine output path
	if outputPath == "" {
		timestamp := time.Now().Format("20060102_150405")
		outputPath = fmt.Sprintf("test-report_%s.html", timestamp)
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Write HTML file
	if err := os.WriteFile(outputPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}

// generateHTML generates the HTML content using a template
func generateHTML(data HTMLReportData) (string, error) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
            color: #333;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            overflow: hidden;
        }
        
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }
        
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
            font-weight: 700;
        }
        
        .header .subtitle {
            font-size: 1.1em;
            opacity: 0.9;
            margin-top: 10px;
        }
        
        .summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            padding: 30px;
            background: #f8f9fa;
        }
        
        .summary-card {
            background: white;
            padding: 25px;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
            text-align: center;
            transition: transform 0.2s;
        }
        
        .summary-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        }
        
        .summary-card .label {
            font-size: 0.9em;
            color: #666;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 10px;
        }
        
        .summary-card .value {
            font-size: 2.5em;
            font-weight: 700;
            margin-bottom: 5px;
        }
        
        .summary-card.passed .value {
            color: #28a745;
        }
        
        .summary-card.failed .value {
            color: #dc3545;
        }
        
        .summary-card.total .value {
            color: #007bff;
        }
        
        .summary-card.duration .value {
            color: #6c757d;
            font-size: 1.8em;
        }
        
        .summary-card.success-rate .value {
            color: {{if geFloat .SuccessRate 80.0}}#28a745{{else if geFloat .SuccessRate 50.0}}#ffc107{{else}}#dc3545{{end}};
        }
        
        .results {
            padding: 30px;
        }
        
        .results-header {
            font-size: 1.5em;
            font-weight: 600;
            margin-bottom: 20px;
            color: #333;
        }
        
        .test-result {
            background: white;
            border: 2px solid #e9ecef;
            border-radius: 8px;
            margin-bottom: 20px;
            overflow: hidden;
            transition: all 0.3s;
        }
        
        .test-result:hover {
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        }
        
        .test-result.passed {
            border-left: 4px solid #28a745;
        }
        
        .test-result.failed {
            border-left: 4px solid #dc3545;
        }
        
        .test-result.skipped {
            border-left: 4px solid #ffc107;
        }
        
        .test-header {
            padding: 20px;
            background: #f8f9fa;
            display: flex;
            justify-content: space-between;
            align-items: center;
            cursor: pointer;
        }
        
        .test-header:hover {
            background: #e9ecef;
        }
        
        .test-name {
            font-size: 1.2em;
            font-weight: 600;
            color: #333;
        }
        
        .test-status {
            display: flex;
            align-items: center;
            gap: 10px;
        }
        
        .status-badge {
            padding: 6px 12px;
            border-radius: 20px;
            font-size: 0.85em;
            font-weight: 600;
            text-transform: uppercase;
        }
        
        .status-badge.passed {
            background: #d4edda;
            color: #155724;
        }
        
        .status-badge.failed {
            background: #f8d7da;
            color: #721c24;
        }
        
        .status-badge.skipped {
            background: #fff3cd;
            color: #856404;
        }
        
        .test-duration {
            color: #6c757d;
            font-size: 0.9em;
        }
        
        .test-details {
            padding: 0 20px;
            max-height: 0;
            overflow: hidden;
            transition: max-height 0.3s ease-out;
        }
        
        .test-details.expanded {
            max-height: 2000px;
            padding: 20px;
        }
        
        .detail-section {
            margin-bottom: 20px;
        }
        
        .detail-section h4 {
            font-size: 1em;
            color: #495057;
            margin-bottom: 10px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        
        .error-message {
            background: #f8d7da;
            color: #721c24;
            padding: 15px;
            border-radius: 6px;
            border-left: 4px solid #dc3545;
            font-family: 'Courier New', monospace;
            white-space: pre-wrap;
            word-break: break-word;
        }
        
        .validations {
            display: grid;
            gap: 10px;
        }
        
        .validation {
            padding: 12px;
            border-radius: 6px;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        
        .validation.passed {
            background: #d4edda;
            border-left: 3px solid #28a745;
        }
        
        .validation.failed {
            background: #f8d7da;
            border-left: 3px solid #dc3545;
        }
        
        .validation-icon {
            font-size: 1.2em;
            font-weight: bold;
        }
        
        .validation-details {
            flex: 1;
        }
        
        .validation-type {
            font-weight: 600;
            color: #495057;
        }
        
        .validation-expected {
            color: #6c757d;
            font-size: 0.9em;
            margin-top: 4px;
        }
        
        .captured-data {
            background: #e7f3ff;
            padding: 15px;
            border-radius: 6px;
            border-left: 4px solid #007bff;
        }
        
        .captured-data pre {
            margin: 0;
            font-family: 'Courier New', monospace;
            font-size: 0.9em;
            color: #333;
            white-space: pre-wrap;
            word-break: break-word;
        }
        
        .print-text {
            background: #fff3cd;
            padding: 15px;
            border-radius: 6px;
            border-left: 4px solid #ffc107;
            color: #856404;
            font-family: 'Courier New', monospace;
            white-space: pre-wrap;
        }
        
        .repeat-results {
            margin-top: 15px;
        }
        
        .repeat-iteration {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 6px;
            margin-bottom: 10px;
            border-left: 3px solid #6c757d;
        }
        
        .footer {
            background: #f8f9fa;
            padding: 20px;
            text-align: center;
            color: #6c757d;
            font-size: 0.9em;
        }
        
        .expand-icon {
            transition: transform 0.3s;
        }
        
        .expanded .expand-icon {
            transform: rotate(180deg);
        }
        
        @media (max-width: 768px) {
            .header h1 {
                font-size: 1.8em;
            }
            
            .summary {
                grid-template-columns: 1fr;
            }
            
            .test-header {
                flex-direction: column;
                align-items: flex-start;
                gap: 10px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
            <div class="subtitle">
                <strong>{{.WorkflowName}}</strong>
                {{if .WorkflowFile}}<br>{{.WorkflowFile}}{{end}}
            </div>
            <div class="subtitle" style="margin-top: 15px; font-size: 0.9em;">
                Generated at {{.GeneratedAt}}
            </div>
        </div>
        
        <div class="summary">
            <div class="summary-card total">
                <div class="label">Total Tests</div>
                <div class="value">{{.TotalTests}}</div>
            </div>
            <div class="summary-card passed">
                <div class="label">Passed</div>
                <div class="value">{{.PassedTests}}</div>
            </div>
            <div class="summary-card failed">
                <div class="label">Failed</div>
                <div class="value">{{.FailedTests}}</div>
            </div>
            {{if gt .SkippedTests 0}}
            <div class="summary-card skipped">
                <div class="label">Skipped</div>
                <div class="value">{{.SkippedTests}}</div>
            </div>
            {{end}}
            <div class="summary-card success-rate">
                <div class="label">Success Rate</div>
                <div class="value">{{printf "%.1f" .SuccessRate}}%</div>
            </div>
            <div class="summary-card duration">
                <div class="label">Duration</div>
                <div class="value">{{formatDuration .TotalDuration}}</div>
            </div>
        </div>
        
        <div class="results">
            <div class="results-header">Test Results</div>
            {{range $index, $result := .Results}}
            <div class="test-result {{$result.Status}}" onclick="toggleDetails({{$index}})">
                <div class="test-header">
                    <div class="test-name">{{$result.Name}}</div>
                    <div class="test-status">
                        <span class="status-badge {{$result.Status}}">{{$result.Status}}</span>
                        <span class="test-duration">{{formatDuration $result.Duration}}</span>
                        <span class="expand-icon">▼</span>
                    </div>
                </div>
                <div class="test-details" id="details-{{$index}}">
                    {{if $result.Error}}
                    <div class="detail-section">
                        <h4>Error</h4>
                        <div class="error-message">{{$result.Error}}</div>
                    </div>
                    {{end}}
                    
                    {{if $result.PrintText}}
                    <div class="detail-section">
                        <h4>Print Output</h4>
                        <div class="print-text">{{$result.PrintText}}</div>
                    </div>
                    {{end}}
                    
                    {{if $result.Validations}}
                    <div class="detail-section">
                        <h4>Validations ({{len $result.Validations}})</h4>
                        <div class="validations">
                            {{range $result.Validations}}
                            <div class="validation {{if .Passed}}passed{{else}}failed{{end}}">
                                <span class="validation-icon">{{if .Passed}}✓{{else}}✗{{end}}</span>
                                <div class="validation-details">
                                    <div class="validation-type">{{.Type}}</div>
                                    <div class="validation-expected">
                                        Expected: {{formatValue .Expected}} | Actual: {{formatValue .Actual}}
                                        {{if .Error}}<br><strong>Error:</strong> {{.Error}}{{end}}
                                    </div>
                                </div>
                            </div>
                            {{end}}
                        </div>
                    </div>
                    {{end}}
                    
                    {{if $result.CapturedData}}
                    <div class="detail-section">
                        <h4>Captured Data</h4>
                        <div class="captured-data">
                            <pre>{{formatJSON $result.CapturedData}}</pre>
                        </div>
                    </div>
                    {{end}}
                    
                    {{if $result.RepeatCount}}
                    <div class="detail-section">
                        <h4>Repeat Results ({{$result.RepeatCount}} iterations)</h4>
                        <div class="repeat-results">
                            {{range $i, $repeatResult := $result.RepeatResults}}
                            <div class="repeat-iteration">
                                <strong>Iteration {{add $i 1}}:</strong> 
                                <span class="status-badge {{$repeatResult.Status}}">{{$repeatResult.Status}}</span>
                                <span class="test-duration">{{formatDuration $repeatResult.Duration}}</span>
                                {{if $repeatResult.Error}}
                                <div class="error-message" style="margin-top: 10px;">{{$repeatResult.Error}}</div>
                                {{end}}
                            </div>
                            {{end}}
                        </div>
                    </div>
                    {{end}}
                    
                    {{if $result.Retries}}
                    <div class="detail-section">
                        <h4>Retries</h4>
                        <div>This test was retried {{$result.Retries}} time(s)</div>
                    </div>
                    {{end}}
                </div>
            </div>
            {{end}}
        </div>
        
        <div class="footer">
            Generated by Stepwise Testing Framework
        </div>
    </div>
    
    <script>
        function toggleDetails(index) {
            const details = document.getElementById('details-' + index);
            const testResult = details.closest('.test-result');
            details.classList.toggle('expanded');
            testResult.classList.toggle('expanded');
        }
    </script>
</body>
</html>`

	// Create template with custom functions
	funcMap := template.FuncMap{
		"formatDuration": func(d time.Duration) string {
			ms := d.Milliseconds()
			if ms < 1000 {
				return fmt.Sprintf("%dms", ms)
			}
			seconds := float64(ms) / 1000.0
			if seconds < 60 {
				return fmt.Sprintf("%.2fs", seconds)
			}
			minutes := int(seconds / 60)
			secs := int(seconds) % 60
			return fmt.Sprintf("%dm %ds", minutes, secs)
		},
		"formatValue": func(v interface{}) string {
			if v == nil {
				return "nil"
			}
			switch val := v.(type) {
			case string:
				return val
			case bool:
				return fmt.Sprintf("%t", val)
			case float64:
				return fmt.Sprintf("%.2f", val)
			default:
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					return fmt.Sprintf("%v", v)
				}
				return string(jsonBytes)
			}
		},
		"formatJSON": func(v interface{}) string {
			jsonBytes, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return fmt.Sprintf("%v", v)
			}
			return string(jsonBytes)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"geFloat": func(a, b float64) bool {
			return a >= b
		},
	}

	t, err := template.New("report").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
