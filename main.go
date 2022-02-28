package vconfig

import (
	"regexp"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/pkg/errors"
)

type VConfig struct {
	remote      string
	user        string
	pass        string
	privateKeys []byte
	publicKeys  *ssh.PublicKeys
	tag         string
	cloneDir    string

	repo     *git.Repository
	workTree *git.Worktree
}

var (
	gitRegex  = regexp.MustCompile(`[a-zA-Z0-9]+@([a-zA-Z0-9\.-]+):([a-zA-Z0-9/-]+).git`)
	httpRegex = regexp.MustCompile(`(https|http)://[\w+.-]+/([\w+.-]+)/([\w+.-]+)`)
)

func New(remote, tag, user, pass string, privateKeys []byte) (config *VConfig, err error) {

	v := &VConfig{
		remote:      remote,
		user:        user,
		pass:        pass,
		privateKeys: privateKeys,
		tag:         tag,
		cloneDir:    remote + "-" + tag,
	}

	if len(privateKeys) != 0 {
		v.publicKeys, err = ssh.NewPublicKeys(user, privateKeys, pass)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return v, nil
}

func (v *VConfig) clone() (err error) {
	v.repo, err = git.PlainOpen(v.cloneDir)
	if err == git.ErrRepositoryNotExists {
		options := &git.CloneOptions{URL: v.remote}
		if v.publicKeys != nil {
			options.Auth = v.publicKeys
		}
		v.repo, err = git.PlainClone(v.cloneDir, false, options)
	}
	return errors.WithStack(err)
}

func (v *VConfig) fetch() (err error) {
	options := &git.FetchOptions{RemoteName: git.DefaultRemoteName}
	if v.publicKeys != nil {
		options.Auth = v.publicKeys
	}
	err = v.repo.Fetch(options)
	if err == nil || err == git.NoErrAlreadyUpToDate {
		return nil
	}

	return errors.WithStack(err)
}

func (v *VConfig) checkout(hash *plumbing.Reference) (err error) {
	head, err := v.repo.Head()
	if err != nil {
		return errors.WithStack(err)
	}

	if head.Hash().String() == hash.Hash().String() {
		return nil
	}

	return errors.WithStack(v.workTree.Checkout(&git.CheckoutOptions{
		Hash: hash.Hash(),
	}))
}

func (v *VConfig) pull() (err error) {
	options := &git.PullOptions{RemoteName: git.DefaultRemoteName}
	if v.publicKeys != nil {
		options.Auth = v.publicKeys
	}
	return errors.WithStack(v.workTree.Pull(options))
}

// Init "https://github.com/moyrne/go-zero"
func (v *VConfig) Init() error {
	if err := v.clone(); err != nil {
		return err
	}
	if err := v.fetch(); err != nil {
		return err
	}

	ref, err := v.repo.Tag(v.tag)
	if err != nil {
		return errors.WithStack(err)
	}

	v.workTree, err = v.repo.Worktree()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := v.checkout(ref); err != nil {
		return err
	}

	return v.pull()
}
