package errors

type ContainerNotExist struct {
	ID string
}

func (ce ContainerNotExist) Error() string {
	return "container " + ce.ID + " not exist"
}
