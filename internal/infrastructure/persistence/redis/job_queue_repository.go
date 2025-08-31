package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"web-page-analyzer/internal/domain/entities"
	"web-page-analyzer/internal/domain/repositories"

	"github.com/go-redis/redis/v8"
)

const (
	JobQueueKey      = "analysis:jobs:pending"
	RetryQueueKey    = "analysis:jobs:retry"
	FailedQueueKey   = "analysis:jobs:failed"
	ProcessingSetKey = "analysis:jobs:processing"
)

type jobQueueRepository struct {
	client *redis.Client
}

func NewJobQueueRepository(client *redis.Client) repositories.JobQueueRepository {
	return &jobQueueRepository{client: client}
}

func (r *jobQueueRepository) Enqueue(ctx context.Context, job *entities.AnalysisJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	score := float64(time.Now().Unix()) - float64(job.Priority*1000)
	return r.client.ZAdd(ctx, JobQueueKey, &redis.Z{
		Score:  score,
		Member: data,
	}).Err()
}

func (r *jobQueueRepository) Dequeue(ctx context.Context) (*entities.AnalysisJob, error) {
	result, err := r.client.ZPopMin(ctx, JobQueueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) == 0 {
		return nil, nil
	}

	var job entities.AnalysisJob
	if err := json.Unmarshal([]byte(result[0].Member.(string)), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	jobData, _ := json.Marshal(job)
	r.client.SAdd(ctx, ProcessingSetKey, jobData)

	return &job, nil
}

func (r *jobQueueRepository) EnqueueWithDelay(ctx context.Context, job *entities.AnalysisJob, delay int) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	executeAt := time.Now().Add(time.Duration(delay) * time.Second)
	score := float64(executeAt.Unix()) - float64(job.Priority*1000)

	return r.client.ZAdd(ctx, RetryQueueKey, &redis.Z{
		Score:  score,
		Member: data,
	}).Err()
}

func (r *jobQueueRepository) GetQueueLength(ctx context.Context) (int64, error) {
	return r.client.ZCard(ctx, JobQueueKey).Result()
}

func (r *jobQueueRepository) GetFailedJobs(ctx context.Context, limit int) ([]*entities.AnalysisJob, error) {
	result, err := r.client.ZRevRange(ctx, FailedQueueKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get failed jobs: %w", err)
	}

	jobs := make([]*entities.AnalysisJob, 0, len(result))
	for _, data := range result {
		var job entities.AnalysisJob
		if err := json.Unmarshal([]byte(data), &job); err != nil {
			continue
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

func (r *jobQueueRepository) MarkJobCompleted(ctx context.Context, job *entities.AnalysisJob) error {
	jobData, _ := json.Marshal(job)
	return r.client.SRem(ctx, ProcessingSetKey, jobData).Err()
}

func (r *jobQueueRepository) MarkJobFailed(ctx context.Context, job *entities.AnalysisJob) error {
	jobData, _ := json.Marshal(job)

	pipe := r.client.Pipeline()
	pipe.SRem(ctx, ProcessingSetKey, jobData)
	pipe.ZAdd(ctx, FailedQueueKey, &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: jobData,
	})

	_, err := pipe.Exec(ctx)
	return err
}

func (r *jobQueueRepository) RequeueRetryJobs(ctx context.Context) error {
	now := time.Now().Unix()

	result, err := r.client.ZRangeByScore(ctx, RetryQueueKey, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%d", now),
	}).Result()

	if err != nil {
		return err
	}

	if len(result) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	for _, jobData := range result {
		var job entities.AnalysisJob
		if err := json.Unmarshal([]byte(jobData), &job); err != nil {
			continue
		}

		score := float64(time.Now().Unix()) - float64(job.Priority*1000)
		pipe.ZAdd(ctx, JobQueueKey, &redis.Z{
			Score:  score,
			Member: jobData,
		})
		pipe.ZRem(ctx, RetryQueueKey, jobData)
	}

	_, err = pipe.Exec(ctx)
	return err
}
