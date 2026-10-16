package report

import (
	"api-health-monitor/internal/model"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func Display(report model.MonitorReport) {
	for _, r:= range report.Results {
		log.Println("=================================")
		log.Println("ID:", r.EndpointID)
		log.Println("URL:", r.URL)
		log.Println("Status:", r.Status)
		log.Println("Status Code:", r.HTTPStatusCode)
		log.Println("Last Attempt Duration:", r.LastAttemptDuration)
		log.Println("Total Elapsed Time:", r.TotalElapsed)
		log.Println("Error:", r.Error)
		log.Println("=================================")
	}
	log.Println("=================================")
	log.Println("RESULTS")
	log.Println("Success Count:", report.SuccessCount)
	log.Println("Cancellation Count:", report.CancelledCount)
	log.Println("Failure Count", report.FailureCount)
	log.Println("Total Run Duration:", report.TotalRunDuration)
	log.Println("=================================")
}

	func Save(report model.MonitorReport, dirpath string) error {
		info, err := os.Stat(dirpath)
		if err != nil {
			return fmt.Errorf("stat report directory %q: %w", dirpath, err)
		}

		if !info.IsDir() {
			return fmt.Errorf("report path %q is not a directory", dirpath)
		}

		data, err := json.MarshalIndent(report, "", "	")
		if err != nil {
			return fmt.Errorf("marshal report: %w", err)
		}

		filename := filepath.Join(dirpath, time.Now().Format("20060102_150405.000000000")+".json")
		if err := os.WriteFile(filename, data, 0644); err != nil {
			return fmt.Errorf("write file %s: %w", filename, err)
		}
		return nil
	}