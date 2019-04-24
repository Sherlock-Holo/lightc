package errors

import "strings"

type ImageImportConflict struct {
	ConflictName string
}

func (ic ImageImportConflict) Error() string {
	return "import image conflict, " + ic.ConflictName + " has existed"
}

type ImageNotFound struct {
	NotFoundName string
}

func (inf ImageNotFound) Error() string {
	return "image " + inf.NotFoundName + " not found"
}

type RootFSCreateConflict struct {
	ConflictID string
}

func (rc RootFSCreateConflict) Error() string {
	return "create rootfs conflict, " + rc.ConflictID + " has existed"
}

type RootFSNotExist struct {
	ID string
}

func (rne RootFSNotExist) Error() string {
	return "rootfs " + rne.ID + " not exist"
}

const (
	VolumeMountOp   = "mount"
	VolumeUnmountOp = "unmount"
)

type VolumeErr struct {
	Op   string
	Errs []error
}

func (ve VolumeErr) Error() string {
	builder := strings.Builder{}

	builder.WriteString(ve.Op)
	builder.WriteString(" volume failed: ")

	for i, err := range ve.Errs {
		builder.WriteString(err.Error())

		if i < len(ve.Errs)-1 {
			builder.WriteString(", ")
		}
	}

	return builder.String()
}
