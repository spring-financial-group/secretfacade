# Facade for reading and writing to secret stores for Go

## Introduction

In order to work with an ever increasing number of secret stores and providers this library provides a basic interface
to allow client applications to read and write secrets. It optionally provides idiomatic authentication mechanisms for
each of the various secret stores.

## Installation and usage

To install, run:

```
$ go get github.com/chrismellard/secretfacade
```

And import using:

```
import "github.com/chrismellard/secretfacade"
```

Usage:

```go
package main

import (
	"fmt"

	"github.com/chrismellard/secretfacade/pkg/secretstore"
	"github.com/chrismellard/secretfacade/pkg/secretstore/factory"
)

func main() {
	factory := factory.SecretManagerFactory{}
	mgr, err := factory.NewSecretManager(secretstore.SecretStoreTypeGoogle)
	if err != nil {
		panic("error creating google secret manager from factory")
	}

	err = mgr.SetSecret("projectId", "myDatabaseConnectionString", &secretstore.SecretValue{Value: "superSecret"})
	if err != nil {
		panic("error setting myDatabaseConnectionString secret")
	}

	connectionStringSecret, err := mgr.GetSecret("projectId", "myDatabaseConnectionString", "")
	if err != nil {
		panic("error getting myDatabaseConnectionString secret")
	}

	fmt.Printf("Please don't print out secrets like %s", connectionStringSecret)
}

```
