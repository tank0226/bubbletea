package tea

import "time"

const (
	// defaultFramerate specifies the maximum interval at which timer-based
	// renderers should update the view.
	defaultFramerate = time.Second / 60
)

type renderer interface{}
