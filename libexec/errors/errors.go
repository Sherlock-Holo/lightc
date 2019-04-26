package errors

type ContainerStopped struct {
	ID string
}

func (cs ContainerStopped) Error() string {
	return "container " + cs.ID + " is stopped"
}
