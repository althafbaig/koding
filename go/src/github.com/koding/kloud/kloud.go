package kloud

import (
	"io/ioutil"
	"log"

	"github.com/koding/kloud/eventer"
	"github.com/koding/kloud/idlock"
	"github.com/koding/kloud/protocol"
	"github.com/koding/kloud/provider/digitalocean"
	"github.com/koding/kloud/provider/openstack"

	"github.com/koding/logging"
)

const (
	VERSION = "0.0.1"
	NAME    = "kloud"
)

type Kloud struct {
	Log logging.Logger

	// Providers is responsible for creating machines and handling them.
	Providers map[string]protocol.Provider

	// Storage is used to store persistent data which is used by the Provider
	// during certain actions
	Storage Storage

	// Eventers is providing an event mechanism for each method.
	Eventers map[string]eventer.Eventer

	// Deployer is executed after a successfull build
	Deployer protocol.Deployer

	// idlock provides multiple locks per id
	idlock *idlock.IdLock
}

// NewKloud creates a new Kloud instance with default providers.
func NewKloud() *Kloud {
	kld := &Kloud{
		idlock:   idlock.New(),
		Log:      logging.NewLogger(NAME),
		Eventers: make(map[string]eventer.Eventer),
	}

	kld.initializeProviders()
	return kld
}

func (k *Kloud) initializeProviders() {
	// Our digitalocean api uses lots of logs, the only way to supress them is
	// to disable std log package.
	log.SetOutput(ioutil.Discard)

	k.Providers = map[string]protocol.Provider{
		"digitalocean": &digitalocean.Provider{
			Log: logging.NewLogger("digitalocean"),
		},
		"rackspace": &openstack.Provider{
			Log:          logging.NewLogger("rackspace"),
			AuthURL:      "https://identity.api.rackspacecloud.com/v2.0",
			ProviderName: "rackspace",
		},
	}
}

func (k *Kloud) GetProvider(providerName string) (protocol.Provider, error) {
	provider, ok := k.Providers[providerName]
	if !ok {
		return nil, NewError(ErrProviderNotFound)
	}

	return provider, nil
}
