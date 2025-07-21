package performance

import (
	"fmt"
	"sync"
	"time"

	"github.com/cjp2600/stepwise/internal/http"
	"github.com/cjp2600/stepwise/internal/logger"
)

// LoadTest represents a load test configuration
type LoadTest struct {
	Concurrency int           `yaml:"concurrency" json:"concurrency"`
	Duration    time.Duration `yaml:"duration" json:"duration"`
	Rate        int           `yaml:"rate" json:"rate"` // requests per second
	Request     *http.Request `yaml:"request" json:"request"`
}

// LoadTestResult represents the result of a load test
type LoadTestResult struct {
	TotalRequests       int             `json:"total_requests"`
	SuccessfulRequests  int             `json:"successful_requests"`
	FailedRequests      int             `json:"failed_requests"`
	TotalDuration       time.Duration   `json:"total_duration"`
	AverageResponseTime time.Duration   `json:"average_response_time"`
	MinResponseTime     time.Duration   `json:"min_response_time"`
	MaxResponseTime     time.Duration   `json:"max_response_time"`
	RequestsPerSecond   float64         `json:"requests_per_second"`
	ErrorRate           float64         `json:"error_rate"`
	ResponseTimes       []time.Duration `json:"response_times"`
	Errors              []string        `json:"errors"`
}

// LoadTester represents a load testing engine
type LoadTester struct {
	client *http.Client
	logger *logger.Logger
}

// NewLoadTester creates a new load tester
func NewLoadTester(client *http.Client, log *logger.Logger) *LoadTester {
	return &LoadTester{
		client: client,
		logger: log,
	}
}

// RunLoadTest executes a load test
func (lt *LoadTester) RunLoadTest(test *LoadTest) (*LoadTestResult, error) {
	lt.logger.Info("Starting load test",
		"concurrency", test.Concurrency,
		"duration", test.Duration,
		"rate", test.Rate)

	startTime := time.Now()
	var wg sync.WaitGroup
	results := make(chan *requestResult, test.Concurrency*100) // Buffer for results
	errors := make(chan error, test.Concurrency*100)

	// Start worker goroutines
	for i := 0; i < test.Concurrency; i++ {
		wg.Add(1)
		go lt.worker(i, test, &wg, results, errors)
	}

	// Wait for all workers to complete
	wg.Wait()
	close(results)
	close(errors)

	// Collect results
	var responseTimes []time.Duration
	var errorMessages []string
	totalRequests := 0
	successfulRequests := 0
	failedRequests := 0

	for result := range results {
		totalRequests++
		if result.error != nil {
			failedRequests++
			errorMessages = append(errorMessages, result.error.Error())
		} else {
			successfulRequests++
			responseTimes = append(responseTimes, result.duration)
		}
	}

	// Calculate metrics
	duration := time.Since(startTime)
	avgResponseTime := lt.calculateAverageResponseTime(responseTimes)
	minResponseTime := lt.calculateMinResponseTime(responseTimes)
	maxResponseTime := lt.calculateMaxResponseTime(responseTimes)
	requestsPerSecond := float64(totalRequests) / duration.Seconds()
	errorRate := float64(failedRequests) / float64(totalRequests) * 100

	lt.logger.Info("Load test completed",
		"total_requests", totalRequests,
		"successful_requests", successfulRequests,
		"failed_requests", failedRequests,
		"duration", duration,
		"requests_per_second", requestsPerSecond,
		"error_rate", errorRate)

	return &LoadTestResult{
		TotalRequests:       totalRequests,
		SuccessfulRequests:  successfulRequests,
		FailedRequests:      failedRequests,
		TotalDuration:       duration,
		AverageResponseTime: avgResponseTime,
		MinResponseTime:     minResponseTime,
		MaxResponseTime:     maxResponseTime,
		RequestsPerSecond:   requestsPerSecond,
		ErrorRate:           errorRate,
		ResponseTimes:       responseTimes,
		Errors:              errorMessages,
	}, nil
}

// requestResult represents the result of a single request
type requestResult struct {
	duration time.Duration
	error    error
}

// worker is a worker goroutine that executes requests
func (lt *LoadTester) worker(id int, test *LoadTest, wg *sync.WaitGroup, results chan<- *requestResult, errors chan<- error) {
	defer wg.Done()

	// Calculate delay between requests based on rate
	var delay time.Duration
	if test.Rate > 0 {
		delay = time.Second / time.Duration(test.Rate/test.Concurrency)
	}

	startTime := time.Now()
	requestCount := 0

	for time.Since(startTime) < test.Duration {
		requestStart := time.Now()

		// Execute request
		_, err := lt.client.Execute(test.Request)
		duration := time.Since(requestStart)

		// Send result
		results <- &requestResult{
			duration: duration,
			error:    err,
		}

		requestCount++

		// Log progress
		if requestCount%100 == 0 {
			lt.logger.Debug("Worker progress",
				"worker_id", id,
				"requests", requestCount,
				"duration", time.Since(startTime))
		}

		// Rate limiting
		if delay > 0 {
			time.Sleep(delay)
		}
	}

	lt.logger.Debug("Worker completed",
		"worker_id", id,
		"total_requests", requestCount,
		"duration", time.Since(startTime))
}

// calculateAverageResponseTime calculates the average response time
func (lt *LoadTester) calculateAverageResponseTime(times []time.Duration) time.Duration {
	if len(times) == 0 {
		return 0
	}

	var total time.Duration
	for _, t := range times {
		total += t
	}
	return total / time.Duration(len(times))
}

// calculateMinResponseTime calculates the minimum response time
func (lt *LoadTester) calculateMinResponseTime(times []time.Duration) time.Duration {
	if len(times) == 0 {
		return 0
	}

	min := times[0]
	for _, t := range times {
		if t < min {
			min = t
		}
	}
	return min
}

// calculateMaxResponseTime calculates the maximum response time
func (lt *LoadTester) calculateMaxResponseTime(times []time.Duration) time.Duration {
	if len(times) == 0 {
		return 0
	}

	max := times[0]
	for _, t := range times {
		if t > max {
			max = t
		}
	}
	return max
}

// StressTest represents a stress test configuration
type StressTest struct {
	InitialConcurrency int           `yaml:"initial_concurrency" json:"initial_concurrency"`
	MaxConcurrency     int           `yaml:"max_concurrency" json:"max_concurrency"`
	StepDuration       time.Duration `yaml:"step_duration" json:"step_duration"`
	StepIncrease       int           `yaml:"step_increase" json:"step_increase"`
	Request            *http.Request `yaml:"request" json:"request"`
}

// StressTestResult represents the result of a stress test
type StressTestResult struct {
	Steps   []*LoadTestResult  `json:"steps"`
	Summary *StressTestSummary `json:"summary"`
}

// StressTestSummary represents a summary of stress test results
type StressTestSummary struct {
	TotalSteps     int             `json:"total_steps"`
	MaxRPS         float64         `json:"max_rps"`
	MaxConcurrency int             `json:"max_concurrency"`
	BreakingPoint  *LoadTestResult `json:"breaking_point"`
}

// RunStressTest executes a stress test
func (lt *LoadTester) RunStressTest(test *StressTest) (*StressTestResult, error) {
	lt.logger.Info("Starting stress test",
		"initial_concurrency", test.InitialConcurrency,
		"max_concurrency", test.MaxConcurrency,
		"step_duration", test.StepDuration)

	var steps []*LoadTestResult
	currentConcurrency := test.InitialConcurrency
	var breakingPoint *LoadTestResult

	for currentConcurrency <= test.MaxConcurrency {
		loadTest := &LoadTest{
			Concurrency: currentConcurrency,
			Duration:    test.StepDuration,
			Request:     test.Request,
		}

		result, err := lt.RunLoadTest(loadTest)
		if err != nil {
			return nil, fmt.Errorf("load test failed at concurrency %d: %w", currentConcurrency, err)
		}

		steps = append(steps, result)

		// Check if we've hit a breaking point (high error rate)
		if result.ErrorRate > 10.0 && breakingPoint == nil {
			breakingPoint = result
		}

		lt.logger.Info("Stress test step completed",
			"concurrency", currentConcurrency,
			"requests_per_second", result.RequestsPerSecond,
			"error_rate", result.ErrorRate)

		currentConcurrency += test.StepIncrease
	}

	// Calculate summary
	summary := &StressTestSummary{
		TotalSteps:    len(steps),
		BreakingPoint: breakingPoint,
	}

	for _, step := range steps {
		if step.RequestsPerSecond > summary.MaxRPS {
			summary.MaxRPS = step.RequestsPerSecond
		}
		if step.SuccessfulRequests > 0 {
			summary.MaxConcurrency = step.TotalRequests / int(test.StepDuration.Seconds())
		}
	}

	return &StressTestResult{
		Steps:   steps,
		Summary: summary,
	}, nil
}

// PerformanceTest represents a performance test configuration
type PerformanceTest struct {
	Name        string      `yaml:"name" json:"name"`
	Description string      `yaml:"description" json:"description"`
	LoadTest    *LoadTest   `yaml:"load_test" json:"load_test"`
	StressTest  *StressTest `yaml:"stress_test" json:"stress_test"`
	Thresholds  *Thresholds `yaml:"thresholds" json:"thresholds"`
}

// Thresholds represents performance thresholds
type Thresholds struct {
	MaxResponseTime      time.Duration `yaml:"max_response_time" json:"max_response_time"`
	MinRequestsPerSecond float64       `yaml:"min_requests_per_second" json:"min_requests_per_second"`
	MaxErrorRate         float64       `yaml:"max_error_rate" json:"max_error_rate"`
}

// PerformanceTestResult represents the result of a performance test
type PerformanceTestResult struct {
	Name       string            `json:"name"`
	Passed     bool              `json:"passed"`
	LoadTest   *LoadTestResult   `json:"load_test,omitempty"`
	StressTest *StressTestResult `json:"stress_test,omitempty"`
	Violations []string          `json:"violations,omitempty"`
}

// RunPerformanceTest executes a performance test
func (lt *LoadTester) RunPerformanceTest(test *PerformanceTest) (*PerformanceTestResult, error) {
	lt.logger.Info("Starting performance test", "name", test.Name)

	result := &PerformanceTestResult{
		Name:   test.Name,
		Passed: true,
	}

	var violations []string

	// Run load test if specified
	if test.LoadTest != nil {
		loadResult, err := lt.RunLoadTest(test.LoadTest)
		if err != nil {
			return nil, fmt.Errorf("load test failed: %w", err)
		}
		result.LoadTest = loadResult

		// Check thresholds
		if test.Thresholds != nil {
			if test.Thresholds.MaxResponseTime > 0 && loadResult.AverageResponseTime > test.Thresholds.MaxResponseTime {
				violations = append(violations, fmt.Sprintf("Average response time %v exceeds threshold %v", loadResult.AverageResponseTime, test.Thresholds.MaxResponseTime))
				result.Passed = false
			}

			if test.Thresholds.MinRequestsPerSecond > 0 && loadResult.RequestsPerSecond < test.Thresholds.MinRequestsPerSecond {
				violations = append(violations, fmt.Sprintf("Requests per second %.2f below threshold %.2f", loadResult.RequestsPerSecond, test.Thresholds.MinRequestsPerSecond))
				result.Passed = false
			}

			if test.Thresholds.MaxErrorRate > 0 && loadResult.ErrorRate > test.Thresholds.MaxErrorRate {
				violations = append(violations, fmt.Sprintf("Error rate %.2f%% exceeds threshold %.2f%%", loadResult.ErrorRate, test.Thresholds.MaxErrorRate))
				result.Passed = false
			}
		}
	}

	// Run stress test if specified
	if test.StressTest != nil {
		stressResult, err := lt.RunStressTest(test.StressTest)
		if err != nil {
			return nil, fmt.Errorf("stress test failed: %w", err)
		}
		result.StressTest = stressResult
	}

	result.Violations = violations

	lt.logger.Info("Performance test completed",
		"name", test.Name,
		"passed", result.Passed,
		"violations", len(violations))

	return result, nil
}
