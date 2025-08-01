# Microsoft Feature Management for Go

Feature management provides a way to develop and expose application functionality based on features. Many applications have special requirements when a new feature is developed such as when the feature should be enabled and under what conditions. This library provides a way to define these relationships, and also integrates into common Golang code patterns to make exposing these features possible.

## Installation

#### Feature management

Get and evaluate feature flags.

```bash
go get github.com/microsoft/Featuremanagement-Go/featuremanagement
```

#### Feature flag provider

Built-in feature flag provider for Azure App Configuration.

```bash
go get github.com/microsoft/Featuremanagement-Go/featuremanagement/providers/azappconfig
```

## Get started

[**Quickstart of Go Console app**](https://learn.microsoft.com/azure/azure-app-configuration/quickstart-feature-flag-go-console): A quickstart guide is available to learn how to integrate feature flags from *Azure App Configuration* into your Go console applications.

[**Quickstart of Go Gin web app**](https://learn.microsoft.com/azure/azure-app-configuration/quickstart-feature-flag-go-gin): A quickstart guide is available to learn how to integrate feature flags from *Azure App Configuration* into your Go Gin web applications.

## Examples

- [Console Application](../example/console)
- [Web Application](../example/gin)

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Trademarks

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft 
trademarks or logos is subject to and must follow 
[Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.