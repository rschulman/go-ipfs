// +build linux darwin freebsd
// +build !nofuse

package readonly

import (
	core "github.com/ipfs/go-ipfs/core"
	mount "github.com/ipfs/go-ipfs/fuse/mount"
)

// Mount mounts ipfs at a given location, and returns a mount.Mount instance.
func Mount(ipfs *core.IpfsNode, mountpoint, allow string) (mount.Mount, error) {
	fsys := NewFileSystem(ipfs)
	return mount.NewMount(ipfs, fsys, mountpoint, allow)
}
