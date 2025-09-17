package main

import (
	"fmt"
	"net/rpc"
	"strings"
	"time"
	"sync"
	"math/rand"
)

type TestResult struct {
	VMNumber  int
	Pattern   string
	LineCount int
	Output    string
	Error     error
	Latency   time.Duration
}

type TestSuite struct {
	clients []*rpc.Client
}

func NewTestSuite() *TestSuite {
	// VMV IP ADDRESSES
	addresses := []string{
		"172.22.159.124:4425",
		"172.22.155.198:4425",
		"172.22.155.125:4425",
		"172.22.159.125:4425",
		"172.22.155.199:4425",
		"172.22.155.126:4425",
		"172.22.159.126:4425",
		"172.22.155.200:4425",
		"172.22.155.127:4425",
		"172.22.159.127:4425",
	}

	clients := make([]*rpc.Client, len(addresses))
	connectedCount := 0
	
	fmt.Println("=== CONNECTING TO VMs ===")
	for i, addr := range addresses {
		client, err := rpc.DialHTTP("tcp", addr)
		if err != nil {
			fmt.Printf("VM %02d: Failed to connect (%s) - %v\n", i + 1, addr, err)
			clients[i] = nil
		} else {
			fmt.Printf("VM %02d: Connected (%s)\n", i + 1, addr)
			clients[i] = client
			connectedCount++
		}
	}
	
	fmt.Printf("Successfully connected to %d/%d VMs\n\n", connectedCount, len(addresses))
	
	return &TestSuite{clients: clients}
}

// Demo Test 1: Frequent Pattern
func (ts *TestSuite) TestFrequentPattern() {
	fmt.Println("\n=== TEST 1: FREQUENT PATTERN ===")
	pattern := "MSIE"
	results := ts.executeDistributedGrep(pattern, false)
	ts.analyzeResults("Frequent Pattern", results)
}

// Demo Test 2: Infrequent Pattern
func (ts *TestSuite) TestInfrequentPattern() {
	fmt.Println("\n=== TEST 2: INFREQUENT PATTERN ===")
	pattern := "harper"
	results := ts.executeDistributedGrep(pattern, false)
	ts.analyzeResults("Infrequent Pattern", results)
}

// Demo Test 3: Regular Expression
func (ts *TestSuite) TestRegularExpression() {
	fmt.Println("\n=== TEST 3: REGULAR EXPRESSION ===")
	pattern := "harper\\|guzman"
	results := ts.executeDistributedGrep(pattern, true)
	ts.analyzeResults("Regular Expression", results)
}

// Demo Test 4: Fault Tolerance
func (ts *TestSuite) TestFaultTolerance() {
	fmt.Println("\n=== TEST 4: FAULT TOLERANCE ===")
	
	// Simulate VM failure by setting client to nil
	crashedVM := 3
	originalClient := ts.clients[crashedVM]
	ts.clients[crashedVM] = nil
	fmt.Printf("Simulated VM %02d failure\n", crashedVM)
	
	// Test with frequent pattern
	pattern := "MSIE"
	results := ts.executeDistributedGrep(pattern, false)
	ts.analyzeResults("Fault Tolerance Test", results)
	
	// Restore client
	ts.clients[crashedVM] = originalClient
	fmt.Printf("VM %02d restored for subsequent tests\n", crashedVM)
}

// Execute grep across all available VMs
func (ts *TestSuite) executeDistributedGrep(pattern string, isRegex bool) []TestResult {
	var cmd string
	if isRegex {
		cmd = fmt.Sprintf("grep -E \"%s\"", pattern)
	} else {
		cmd = fmt.Sprintf("grep \"%s\"", pattern)
	}
	
	fmt.Printf("Executing command: %s\n", cmd)
	
	var results []TestResult
	var wg sync.WaitGroup
	resultsChan := make(chan TestResult, len(ts.clients))
	
	for i, client := range ts.clients {
		// only run on 4 machines
		if i == 4 {
			break
		}
		wg.Add(1)
		go func(vmNum int, c *rpc.Client) {
			defer wg.Done()
			
			result := TestResult{
				VMNumber: vmNum,
				Pattern:  pattern,
			}
			
			if c == nil {
				result.Error = fmt.Errorf("VM not available")
				resultsChan <- result
				return
			}
			
			start := time.Now()
			var reply string
			err := c.Call("VM.Grep", cmd, &reply)
			result.Latency = time.Since(start)
			
			if err != nil {
				if strings.Contains(err.Error(), "no match found") {
					result.LineCount = 0
					result.Output = ""
				} else {
					result.Error = err
				}
			} else {
				result.Output = reply
				if reply != "" {
					result.LineCount = len(strings.Split(strings.TrimSpace(reply), "\n"))
				} else {
					result.LineCount = 0
				}
			}
			
			resultsChan <- result
		}(i, client)
	}
	
	wg.Wait()
	close(resultsChan)
	
	for result := range resultsChan {
		results = append(results, result)
	}
	
	return results
}

// Analyze and display test results
func (ts *TestSuite) analyzeResults(testName string, results []TestResult) {
	fmt.Printf("\n--- %s Results ---\n", testName)
	
	totalLines := 0
	successfulVMs := 0
	var totalLatency time.Duration
	unavailableVMs := 0
	
	for _, result := range results {
		if result.Error != nil {
			if strings.Contains(result.Error.Error(), "not available") {
				fmt.Printf("VM %02d: UNAVAILABLE\n", result.VMNumber + 1)
				unavailableVMs++
			} else {
				fmt.Printf("VM %02d: ERROR - %v\n", result.VMNumber + 1, result.Error)
			}
		} else {
			fmt.Printf("VM %02d: %d lines matched (latency: %v)\n", 
				result.VMNumber + 1, result.LineCount, result.Latency)
			totalLines += result.LineCount
			successfulVMs++
			totalLatency += result.Latency
		}
	}
	
	fmt.Printf("\n** SUMMARY **\n")
	fmt.Printf("- Total matching lines across all VMs: %d\n", totalLines)
	fmt.Printf("- Successful VMs: %d\n", successfulVMs)
	if unavailableVMs > 0 {
		fmt.Printf("- Unavailable VMs: %d\n", unavailableVMs)
	}
	if successfulVMs > 0 {
		avgLatency := totalLatency / time.Duration(successfulVMs)
		fmt.Printf("- Average latency: %v\n", avgLatency)
	}
	
	// Display sample outputs from first few VMs with results
	fmt.Printf("\n** SAMPLE OUTPUTS **\n")
	sampleCount := 0
	maxSamples := 3
	
	for _, result := range results {
		if sampleCount >= maxSamples {
			break
		}
		
		if result.Error == nil && result.Output != "" {
			fmt.Printf("\nVM %02d sample output:\n", result.VMNumber + 1)
			lines := strings.Split(result.Output, "\n")
			
			displayLines := 3
			if len(lines) < displayLines {
				displayLines = len(lines)
			}
			
			for i := 0; i < displayLines; i++ {
				line := strings.TrimSpace(lines[i])
				if line != "" {
					fmt.Printf("  %s\n", line)
				}
			}
			
			if len(lines) > displayLines {
				fmt.Printf("  ... (%d more lines)\n", len(lines)-displayLines)
			}
			
			sampleCount++
		}
	}
	
	if sampleCount == 0 {
		fmt.Println("No sample outputs available")
	}
	
	fmt.Println(strings.Repeat("=", 60))
}

// Performance test with multiple trials
func (ts *TestSuite) TestPerformance() {
	fmt.Println("\n=== PERFORMANCE TEST ===")
	
	patterns := []string{"MSIE", "harper", "smith"}
	trials := 5
	
	for _, pattern := range patterns {
		fmt.Printf("\nTesting pattern '%s' with %d trials:\n", pattern, trials)
		
		var latencies []time.Duration
		var lineCounts []int
		
		for trial := 1; trial <= trials; trial++ {
			fmt.Printf("Trial %d/%d... ", trial, trials)
			
			results := ts.executeDistributedGrep(pattern, false)
			
			var totalLatency time.Duration
			totalLines := 0
			successfulVMs := 0
			
			for _, result := range results {
				if result.Error == nil {
					totalLatency += result.Latency
					totalLines += result.LineCount
					successfulVMs++
				}
			}
			
			if successfulVMs > 0 {
				avgLatency := totalLatency / time.Duration(successfulVMs)
				latencies = append(latencies, avgLatency)
				lineCounts = append(lineCounts, totalLines)
				fmt.Printf("avg latency: %v, total lines: %d\n", avgLatency, totalLines)
			} else {
				fmt.Println("no successful queries")
			}
		}
		
		// Calculate performance statistics
		if len(latencies) > 0 {
			ts.calculatePerformanceStats(pattern, latencies, lineCounts)
		}
	}
}

func (ts *TestSuite) calculatePerformanceStats(pattern string, latencies []time.Duration, lineCounts []int) {
	if len(latencies) == 0 {
		return
	}
	
	// Calculate average latency
	var totalLatency time.Duration
	for _, latency := range latencies {
		totalLatency += latency
	}
	avgLatency := totalLatency / time.Duration(len(latencies))
	
	// Find min and max latency
	minLatency := latencies[0]
	maxLatency := latencies[0]
	for _, latency := range latencies {
		if latency < minLatency {
			minLatency = latency
		}
		if latency > maxLatency {
			maxLatency = latency
		}
	}
	
	// Calculate average line count
	totalLineCount := 0
	for _, count := range lineCounts {
		totalLineCount += count
	}
	avgLineCount := float64(totalLineCount) / float64(len(lineCounts))
	
	fmt.Printf("\nPerformance Statistics for '%s':\n", pattern)
	fmt.Printf("- Successful trials: %d\n", len(latencies))
	fmt.Printf("- Average latency: %v\n", avgLatency)
	fmt.Printf("- Min latency: %v\n", minLatency)
	fmt.Printf("- Max latency: %v\n", maxLatency)
	fmt.Printf("- Average line count: %.1f\n", avgLineCount)
}

// Run all demo tests
func (ts *TestSuite) RunDemoTests() {
	fmt.Println("Starting Demo Tests for MP1...")
	
	// Check which VMs are connected
	connectedCount := 0
	for i, client := range ts.clients {
		if client != nil {
			connectedCount++
		} else {
			fmt.Printf("VM %02d: Not connected\n", i)
		}
	}
	
	if connectedCount == 0 {
		fmt.Println("ERROR: No VMs connected. Please start servers first.")
		return
	}
	
	fmt.Printf("Running tests on %d connected VMs...\n", connectedCount)
	
	// Run demo tests
	ts.TestFrequentPattern()
	ts.TestInfrequentPattern()
	ts.TestRegularExpression() 
	ts.TestFaultTolerance()
	
	// Additional performance test
	ts.TestPerformance()
	
	fmt.Println("\n=== ALL DEMO TESTS COMPLETED ===")
	fmt.Println("Ready for TA demo!")
}

// Close all connections
func (ts *TestSuite) Close() {
	fmt.Println("\nClosing connections...")
	for i, client := range ts.clients {
		if client != nil {
			client.Close()
			fmt.Printf("Closed VM %02d\n", i + 1)
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	
	fmt.Println("CS425 MP1 - Distributed Log Querier Unit Tests")
	fmt.Println("=" + strings.Repeat("=", 50))
	
	// Create test suite
	testSuite := NewTestSuite()
	defer testSuite.Close()
	
	testSuite.RunDemoTests()
	
	fmt.Println("\nTests completed successfully!")
	fmt.Println("Press Enter to exit...")
	fmt.Scanln() 
}
