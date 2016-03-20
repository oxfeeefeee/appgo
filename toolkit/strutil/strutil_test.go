package strutil

import (
	"github.com/oxfeeefeee/appgo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStrutil(t *testing.T) {
	id := appgo.Id(123)
	assert.Equal(t, id, ToId(FromId(id)))
	t.Log(FromIdList([]appgo.Id{appgo.Id(1), appgo.Id(2)}))
}
