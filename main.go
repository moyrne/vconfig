package vconfig

import (
	fixtures "github.com/go-git/go-git-fixtures/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

type VConfig struct {
	*git.Remote
}

func New(user, host, pass string) *VConfig {
	store := filesystem.NewStorage(fixtures.ByURL("").ByTag("").One().DotGit(), cache.NewObjectLRUDefault())
	return &VConfig{Remote: git.NewRemote(store, &config.RemoteConfig{})}
}

func (v VConfig) Init() {

}
