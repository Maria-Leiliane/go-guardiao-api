package aws

import (
	"context"
	"errors"
	"os"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
)

// Notifier define operações essenciais de notificação (SNS).
type Notifier interface {
	Publish(ctx context.Context, subject, message string, attrs map[string]string) (messageID string, err error)
}

type SNSClient struct {
	client   *sns.Client
	topicArn string
}

// NewSNS cria cliente SNS. Suporta endpoint override via AWS_SNS_ENDPOINT (LocalStack).
func NewSNS(cfg awsv2.Config, topicArn string) (*SNSClient, error) {
	if topicArn == "" {
		return nil, errors.New("topicArn não pode ser vazio")
	}

	endpoint := os.Getenv("AWS_SNS_ENDPOINT")
	var svc *sns.Client
	if endpoint != "" {
		// Preferencial nas versões recentes do SDK
		svc = sns.NewFromConfig(cfg, func(o *sns.Options) {
			o.BaseEndpoint = awsv2.String(endpoint)
			// Fallback se a versão do SDK não tiver BaseEndpoint:
			// o.EndpointResolverV2 = sns.EndpointResolverFromURL(endpoint)
		})
	} else {
		svc = sns.NewFromConfig(cfg)
	}

	return &SNSClient{client: svc, topicArn: topicArn}, nil
}

func (n *SNSClient) Publish(ctx context.Context, subject, message string, attrs map[string]string) (string, error) {
	msgAttrs := map[string]snstypes.MessageAttributeValue{}
	for k, v := range attrs {
		val := v
		msgAttrs[k] = snstypes.MessageAttributeValue{
			DataType:    awsv2.String("String"),
			StringValue: &val,
		}
	}

	out, err := n.client.Publish(ctx, &sns.PublishInput{
		TopicArn:          awsv2.String(n.topicArn),
		Subject:           awsv2.String(subject),
		Message:           awsv2.String(message),
		MessageAttributes: msgAttrs,
	})
	if err != nil {
		return "", err
	}
	return awsv2.ToString(out.MessageId), nil
}
