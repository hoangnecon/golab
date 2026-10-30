package golab_test

import (
	"testing"

	"github.com/hoangnecon/golab/internal/domain"
	"github.com/hoangnecon/golab/internal/driver/golab"
)

func TestGolabDriverInterfaceImplementation(t *testing.T) {
	var _ domain.GolabDriver = (*golab.Driver)(nil)
}
