// Copyright (c) HashiCorp, Inc.

package mql

type options struct {
	withSkipWhitespace bool
}

// Option - how options are passed as args
type Option func(*options) error

func getDefaultOptions() options {
	return options{}
}

func getOpts(opt ...Option) (options, error) {
	opts := getDefaultOptions()

	for _, o := range opt {
		if err := o(&opts); err != nil {
			return opts, err
		}
	}
	return opts, nil
}

// withSkipWhitespace provides an option to request that whitespace be skipped
func withSkipWhitespace() Option {
	return func(o *options) error {
		o.withSkipWhitespace = true
		return nil
	}
}
