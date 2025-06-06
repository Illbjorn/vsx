package gallery

import (
	_ "embed"
	"testing"

	"github.com/illbjorn/echo"
	"github.com/stretchr/testify/assert"
)

//go:embed sample.json
var sample string

func TestExtension(t *testing.T) {
	g := NewGallery("https", "marketplace.visualstudio.com")
	extensions, err := g.Query(t.Context(), "golang")
	assert.NoError(t, err)
	assert.True(t, len(extensions) > 0)

	
	for _, result := range extensions {
		for _, ext := range result.Extensions {
			echo.Infof("Found extension [%#v].", ext.DisplayName)
		}
	}
}
