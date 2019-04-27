package pivotRoot

func beforePivotRoot(root string) error {
	/*if err := os.MkdirAll(filepath.Join(root, "/dev"), 0700); err != nil {
		return xerrors.Errorf("mkdir %s failed: %w", filepath.Join(root, "/dev"), err)
	}

	if err := syscall.Mount("tmpfs", filepath.Join(root, "/dev"), "tmpfs", syscall.MS_NOSUID, "mode=755"); err != nil {
		return xerrors.Errorf("mount %s failed: %w", filepath.Join(root, "/dev"), err)
	}*/

	return nil
}
