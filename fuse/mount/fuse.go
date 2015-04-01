// +build !nofuse

package mount

import (
	"errors"
	"fmt"
	"time"

	"github.com/ipfs/go-ipfs/Godeps/_workspace/src/bazil.org/fuse"
	"github.com/ipfs/go-ipfs/Godeps/_workspace/src/bazil.org/fuse/fs"
	"github.com/ipfs/go-ipfs/Godeps/_workspace/src/github.com/jbenet/go-ctxgroup"
)

// mount implements go-ipfs/fuse/mount
type mount struct {
	mpoint   string
	filesys  fs.FS
	fuseConn *fuse.Conn
	// closeErr error

	cg ctxgroup.ContextGroup
}

// Mount mounts a fuse fs.FS at a given location, and returns a Mount instance.
// parent is a ContextGroup to bind the mount's ContextGroup to.
func NewMount(p ctxgroup.ContextGroup, fsys fs.FS, mountpoint, allow string) (Mount, error) {
	var conn *fuse.Conn
	var err error

	switch {
	case len(allow) == 0:
		conn, err = fuse.Mount(mountpoint)
	case allow == "root":
		conn, err = fuse.Mount(mountpoint, fuse.AllowRoot())
	case allow == "other":
		conn, err = fuse.Mount(mountpoint, fuse.AllowOther())
	case true:
		return nil, errors.New("Valid mount options for allow: 'root', 'other'")
	}

	if err != nil {
		return nil, err
	}

	m := &mount{
		mpoint:   mountpoint,
		fuseConn: conn,
		filesys:  fsys,
		cg:       ctxgroup.WithParent(p), // link it to parent.
	}
	m.cg.SetTeardown(m.unmount)

	// launch the mounting process.
	if err := m.mount(); err != nil {
		m.Unmount() // just in case.
		return nil, err
	}

	return m, nil
}

func (m *mount) mount() error {
	log.Infof("Mounting %s", m.MountPoint())

	errs := make(chan error, 1)
	go func() {
		err := fs.Serve(m.fuseConn, m.filesys)
		log.Debugf("Mounting %s -- fs.Serve returned (%s)", err)
		if err != nil {
			errs <- err
		}
	}()

	// wait for the mount process to be done, or timed out.
	select {
	case <-time.After(MountTimeout):
		return fmt.Errorf("Mounting %s timed out.", m.MountPoint())
	case err := <-errs:
		return err
	case <-m.fuseConn.Ready:
	}

	// check if the mount process has an error to report
	if err := m.fuseConn.MountError; err != nil {
		return err
	}

	log.Infof("Mounted %s", m.MountPoint())
	return nil
}

// umount is called exactly once to unmount this service.
// note that closing the connection will not always unmount
// properly. If that happens, we bring out the big guns
// (mount.ForceUnmountManyTimes, exec unmount).
func (m *mount) unmount() error {
	log.Infof("Unmounting %s", m.MountPoint())

	// try unmounting with fuse lib
	err := fuse.Unmount(m.MountPoint())
	if err == nil {
		return nil
	}
	log.Debug("fuse unmount err: %s", err)

	// try closing the fuseConn
	err = m.fuseConn.Close()
	if err == nil {
		return nil
	}
	if err != nil {
		log.Debug("fuse conn error: %s", err)
	}

	// try mount.ForceUnmountManyTimes
	if err := ForceUnmountManyTimes(m, 10); err != nil {
		return err
	}

	log.Infof("Seemingly unmounted %s", m.MountPoint())
	return nil
}

func (m *mount) CtxGroup() ctxgroup.ContextGroup {
	return m.cg
}

func (m *mount) MountPoint() string {
	return m.mpoint
}

func (m *mount) Unmount() error {
	// call ContextCloser Close(), which calls unmount() exactly once.
	return m.cg.Close()
}
