package aws

import (
	"context"
	"errors"
	"fmt"
	"os"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// QueueMessage representa uma mensagem recebida de uma fila.
type QueueMessage struct {
	MessageID     string
	Body          string
	ReceiptHandle string
	Attributes    map[string]string
}

// QueueClient define operações essenciais de uma fila (SQS).
type QueueClient interface {
	Send(ctx context.Context, body string, attrs map[string]string) (string, error)
	Receive(ctx context.Context, max int32, waitSeconds int32, visibilityTimeout int32) ([]QueueMessage, error)
	Delete(ctx context.Context, receiptHandle string) error
	Purge(ctx context.Context) error
}

type SQSClient struct {
	client   *sqs.Client
	queueURL string
}

// NewSQS cria um cliente SQS usando o cfg carregado.
// Suporta endpoint override via AWS_SQS_ENDPOINT (ex.: LocalStack).
func NewSQS(cfg awsv2.Config, queueURL string) (*SQSClient, error) {
	if queueURL == "" {
		return nil, errors.New("queueURL não pode ser vazio")
	}

	endpoint := os.Getenv("AWS_SQS_ENDPOINT")
	var svc *sqs.Client
	if endpoint != "" {
		// Recomendado nas versões recentes: BaseEndpoint
		svc = sqs.NewFromConfig(cfg, func(o *sqs.Options) {
			o.BaseEndpoint = awsv2.String(endpoint)
			// Alternativa (se a versão do SDK não tiver BaseEndpoint):
			// o.EndpointResolverV2 = sqs.EndpointResolverFromURL(endpoint)
		})
	} else {
		svc = sqs.NewFromConfig(cfg)
	}

	return &SQSClient{client: svc, queueURL: queueURL}, nil
}

func (q *SQSClient) Send(ctx context.Context, body string, attrs map[string]string) (string, error) {
	msgAttrs := map[string]sqstypes.MessageAttributeValue{}
	for k, v := range attrs {
		msgAttrs[k] = sqstypes.MessageAttributeValue{
			DataType:    awsv2.String("String"),
			StringValue: awsv2.String(v),
		}
	}

	out, err := q.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:          awsv2.String(q.queueURL),
		MessageBody:       awsv2.String(body),
		MessageAttributes: msgAttrs,
	})
	if err != nil {
		return "", fmt.Errorf("sqs send message: %w", err)
	}
	if out.MessageId == nil {
		return "", errors.New("sqs send message: MessageId vazio")
	}
	return *out.MessageId, nil
}

func (q *SQSClient) Receive(ctx context.Context, max int32, waitSeconds int32, visibilityTimeout int32) ([]QueueMessage, error) {
	if max <= 0 || max > 10 {
		max = 10
	}
	out, err := q.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            awsv2.String(q.queueURL),
		MaxNumberOfMessages: max,
		WaitTimeSeconds:     waitSeconds,
		VisibilityTimeout:   visibilityTimeout,
		// Mantemos apenas os atributos customizados (MessageAttributes)
		MessageAttributeNames: []string{"All"},
		// AttributeNames removido para evitar depreciação e porque não é usado
	})
	if err != nil {
		return nil, fmt.Errorf("sqs receive: %w", err)
	}

	msgs := make([]QueueMessage, 0, len(out.Messages))
	for _, m := range out.Messages {
		attr := map[string]string{}
		for k, v := range m.MessageAttributes {
			if v.StringValue != nil {
				attr[k] = *v.StringValue
			}
		}
		msgs = append(msgs, QueueMessage{
			MessageID:     awsv2.ToString(m.MessageId),
			Body:          awsv2.ToString(m.Body),
			ReceiptHandle: awsv2.ToString(m.ReceiptHandle),
			Attributes:    attr,
		})
	}
	return msgs, nil
}

func (q *SQSClient) Delete(ctx context.Context, receiptHandle string) error {
	_, err := q.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      awsv2.String(q.queueURL),
		ReceiptHandle: awsv2.String(receiptHandle),
	})
	if err != nil {
		return fmt.Errorf("sqs delete: %w", err)
	}
	return nil
}

func (q *SQSClient) Purge(ctx context.Context) error {
	_, err := q.client.PurgeQueue(ctx, &sqs.PurgeQueueInput{
		QueueUrl: awsv2.String(q.queueURL),
	})
	if err != nil {
		return fmt.Errorf("sqs purge: %w", err)
	}
	return nil
}
