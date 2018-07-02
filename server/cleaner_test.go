package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCleaner(t *testing.T) {
	c := NewCleaner()
	added := map[time.Duration]string{}
	for _, ttl := range []time.Duration{3 * time.Minute, time.Minute, 2 * time.Minute} {
		added[ttl] = ttl.String()
		c.Add(time.Time{}.Add(ttl), ttl.String())
	}
	assert.Equal(t, added[time.Minute], c.PopOneSince(time.Time{}.Add(90*time.Second)))
	assert.Len(t, c.heap, 2)
	assert.Empty(t, c.PopOneSince(time.Time{}.Add(119*time.Second)))
	assert.Equal(t, added[2*time.Minute], c.PopOneSince(time.Time{}.Add(121*time.Second)))
	assert.Equal(t, added[3*time.Minute], c.PopOneSince(time.Time{}.Add(200*time.Second)))
	assert.Empty(t, c.PopOneSince(time.Time{}.Add(500*time.Second)))
}
