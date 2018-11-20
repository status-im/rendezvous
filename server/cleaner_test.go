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
	c.Add(time.Time{}.Add(140*time.Second), time.Minute.String())
	assert.Len(t, c.heap, 3)
	assert.Equal(t, []string{added[2*time.Minute]}, c.PopSince(time.Time{}.Add(121*time.Second)))
	assert.Len(t, c.heap, 2)
	assert.Equal(t, []string{added[time.Minute]}, c.PopSince(time.Time{}.Add(141*time.Second)))
	assert.Equal(t, []string{added[3*time.Minute]}, c.PopSince(time.Time{}.Add(200*time.Second)))
	assert.Empty(t, c.PopSince(time.Time{}.Add(500*time.Second)))
}
