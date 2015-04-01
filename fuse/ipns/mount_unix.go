// +build linux darwin freebsd
// +build !nofuse

package ipns

import (
	core "github.com/ipfs/go-ipfs/core"
	mount "github.com/ipfs/go-ipfs/fuse/mount"
)

// Mount mounts ipns at a given location, and returns a mount.Mount instance.
func Mount(ipfs *core.IpfsNode, ipnsmp, ipfsmp, allow string) (mount.Mount, error) {
	fsys, err := NewFileSystem(ipfs, ipfs.PrivateKey, ipfsmp)
	if err != nil {
		return nil, err
	}

	return mount.NewMount(ipfs, fsys, ipnsmp, allow)
}
