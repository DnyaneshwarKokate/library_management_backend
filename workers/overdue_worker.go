package workers

import (
	"sync"
	"sync/atomic"

	"library-management-backend/dto"
	"library-management-backend/model"

	"github.com/sirupsen/logrus"
)

func ProcessOverdueWorkerPool(records []model.BorrowRecord, updateFunc func(recordID uint) error) *dto.ProcessOverdueResponse {
	total := int64(len(records))
	if total == 0 {
		return &dto.ProcessOverdueResponse{
			TotalOverdueFound: 0,
			ProcessedCount:    0,
			FailedCount:       0,
		}
	}

	numWorkers := 5
	if int(total) < numWorkers {
		numWorkers = int(total)
	}

	jobs := make(chan model.BorrowRecord, total)
	var processedCount int64
	var failedCount int64
	var wg sync.WaitGroup

	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for record := range jobs {
				err := updateFunc(record.ID)
				if err != nil {
					logrus.Errorf("Worker %d failed to process overdue record ID %d: %v", workerID, record.ID, err)
					atomic.AddInt64(&failedCount, 1)
				} else {
					logrus.Infof("Worker %d successfully updated overdue record ID %d to OVERDUE", workerID, record.ID)
					atomic.AddInt64(&processedCount, 1)
				}
			}
		}(w)
	}

	for _, record := range records {
		jobs <- record
	}
	close(jobs)

	wg.Wait()

	return &dto.ProcessOverdueResponse{
		TotalOverdueFound: total,
		ProcessedCount:    processedCount,
		FailedCount:       failedCount,
	}
}
