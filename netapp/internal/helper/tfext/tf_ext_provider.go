package tfext

import (
	"context"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

// Provider represents Terraform resoure provider overwritten
// with StopFunc called during stop if defined
type Provider struct {
	*schema.Provider

	// StopFunc is a function for stopping the provider. If the
	// provider doesn't need to be stopped, this can be omitted.
	//
	// See the StopFunc documentation for more information.
	StopFunc StopFunc
}

// StopFunc is the function used to stop a Provider.
//
type StopFunc func(meta interface{}) error

// Stop overwrite of the schema.Provider Stop call
//
// Will call schema.Provider.Stop() after call to StopFunc
func (p *Provider) Stop() error {
	log.Printf("[DEBUG] (TFEXTPROV) stop provider called")
	var err error
	if p.StopFunc != nil {
		err = p.StopFunc(p.Meta())
	}

	if err == nil {
		err = p.Provider.Stop()
	}

	return err
}

func (p *Provider) InternalValidate() error {
	log.Printf("[DEBUG] (TFEXTPROV) internal validate called")
	return p.Provider.InternalValidate()
}

func (p *Provider) SetMeta(v interface{}) {
	log.Printf("[DEBUG] (TFEXTPROV) set meta called")
	p.Provider.SetMeta(v)
}

// Stopped reports whether the provider has been stopped or not.
func (p *Provider) Stopped() bool {
	log.Printf("[DEBUG] (TFEXTPROV) stopped called")
	return p.Provider.Stopped()
}

// StopContext returns a channel that is closed once the provider is stopped.
func (p *Provider) StopContext() context.Context {
	log.Printf("[DEBUG] (TFEXTPROV) StopContext called")
	return p.Provider.StopContext()
}
