module github.com/jaffee/aicli

go 1.21.1

replace github.com/sashabaranov/go-openai => github.com/jaffee/go-openai v0.0.0-20231121153610-1c05908c31a0

require (
	github.com/aws/aws-sdk-go-v2/config v1.25.5
	github.com/aws/aws-sdk-go-v2/service/bedrock v1.3.3
	github.com/aws/aws-sdk-go-v2/service/bedrockruntime v1.3.3
	github.com/jaffee/commandeer v0.6.0
	github.com/pkg/errors v0.9.1
	github.com/sashabaranov/go-openai v1.17.0
	github.com/stretchr/testify v1.2.2
	github.com/wader/readline v0.0.0-20230307172220-bcb7158e7448
)

require (
	github.com/aws/aws-sdk-go-v2 v1.23.1 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.5.1 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.16.4 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.14.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.2.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.5.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.10.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.10.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.17.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.20.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.25.4 // indirect
	github.com/aws/smithy-go v1.17.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.1.0 // indirect
	golang.org/x/text v0.3.7 // indirect
)
