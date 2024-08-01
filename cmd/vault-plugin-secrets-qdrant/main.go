package main

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/plugin"
	qdrant "github.com/migrx-io/vault-plugin-secrets-qdrant/plugin"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{})

	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()
	err := flags.Parse(os.Args[1:])
	if err != nil {
		logger.Error("plugin shutting down", "invalid args", err)
		os.Exit(1)
	}

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

	err = plugin.ServeMultiplex(&plugin.ServeOpts{
		BackendFactoryFunc: qdrant.Factory,
		TLSProviderFunc:    tlsProviderFunc,
	})
	if err != nil {
		logger.Error("plugin shutting down", "serve error", err)
		os.Exit(1)
	}
}
