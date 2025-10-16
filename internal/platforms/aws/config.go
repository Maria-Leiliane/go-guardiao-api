package aws

import (
	"context"
	"os"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// LoadConfig carrega a configuração AWS (region, credenciais via env/SharedConfig).
// Suporta region via AWS_REGION. Credenciais podem vir de:
// - AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN
// - ~/.aws/credentials e ~/.aws/config
func LoadConfig(ctx context.Context) (awsv2.Config, error) {
	region := os.Getenv("AWS_REGION")

	var opts []func(*config.LoadOptions) error
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}
	// Carrega cadeia padrão de providers (env > shared > IMDS)
	return config.LoadDefaultConfig(ctx, opts...)
}
