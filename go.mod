module github.com/jenkins-x-plugins/secretfacade

go 1.15

require (
	cloud.google.com/go v0.75.0
	github.com/Azure/azure-sdk-for-go v51.1.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.17
	github.com/Azure/go-autorest/autorest/adal v0.9.11
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.7
	github.com/aws/aws-sdk-go v1.37.19
	github.com/hashicorp/vault v1.8.5
	github.com/hashicorp/vault-plugin-auth-kubernetes v0.11.1
	github.com/hashicorp/vault/api v1.1.2-0.20210713235431-1fc8af4c041f
	github.com/hashicorp/vault/sdk v0.2.2-0.20211101151547-6654f4b913f9
	github.com/imdario/mergo v0.3.12
	github.com/jenkins-x/jx-logging/v3 v3.0.6
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/oauth2 v0.0.0-20210113205817-d3ed898aa8a3
	google.golang.org/api v0.36.0
	google.golang.org/genproto v0.0.0-20210108203827-ffc7fda8c3d7
	google.golang.org/grpc v1.45.0
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
)
