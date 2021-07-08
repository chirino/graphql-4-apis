package apis

import (
	"github.com/chirino/graphql/resolvers"
	"reflect"
	"sync"
)

type dataLoaders map[loadKey]*CachedResolution
type loadKey struct {
	path   string
	method string
	args   string
}

type CachedResolution struct {
	once  sync.Once
	apply resolvers.Resolution
	value reflect.Value
	err   error
}

func (load *CachedResolution) resolution() (value reflect.Value, err error) {
	// concurrent calls will wait for the first call to finish..
	load.once.Do(func() {
		load.value, load.err = load.apply()
	})
	return load.value, load.err
}
