package nonOpWriteCloser

type NonOpWriteCloser struct{}

func (NonOpWriteCloser) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (NonOpWriteCloser) Close() error {
	return nil
}
