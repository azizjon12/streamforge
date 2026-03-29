package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/azizjon12/streamforge/internal/api"
	"github.com/nats-io/nats.go"
)

const DefaultSubject = "streamforge.jobs"

type NATSQueue struct {
	conn    *nats.Conn
	subject string
	sub     *nats.Subscription
}

func NewNATSQueue(url, subject string) (*NATSQueue, error) {
	if subject == "" {
		subject = DefaultSubject
	}

	conn, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}

	return &NATSQueue{
		conn:    conn,
		subject: subject,
	}, nil
}

func (q *NATSQueue) Enqueue(ctx context.Context, job api.StreamJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- q.conn.Publish(q.subject, payload)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("publish job: %w", err)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *NATSQueue) Dequeue(ctx context.Context) (api.StreamJob, error) {
	if q.sub == nil {
		sub, err := q.conn.SubscribeSync(q.subject)
		if err != nil {
			return api.StreamJob{}, fmt.Errorf("subscribe: %w", err)
		}
		q.sub = sub
	}

	for {
		msg, err := q.sub.NextMsgWithContext(ctx)
		if err != nil {
			return api.StreamJob{}, err
		}

		var job api.StreamJob
		if err := json.Unmarshal(msg.Data, &job); err != nil {
			return api.StreamJob{}, fmt.Errorf("unmarshal: %w", err)
		}

		return job, nil
	}
}

func (q *NATSQueue) Close() {
	if q.sub != nil {
		_ = q.sub.Unsubscribe()
	}
	if q.conn != nil {
		q.conn.Close()
	}
}
