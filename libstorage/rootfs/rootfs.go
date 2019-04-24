package rootfs

import (
	"math/rand"
	"time"
)

const (
	idLength = 10
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type RootFS struct {
	ID        string
	ImageName string
	LowerDir  string
	UpperDir  string
	WorkDir   string
	MergedDir string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateInfoID() string {
	b := make([]rune, idLength)
	for i := 0; i < idLength; i++ {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
