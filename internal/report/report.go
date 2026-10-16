package report

import (
	"api-health-monitor/internal/model"
	"log"
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