package ledtx

import (
	"encoding/binary"

	"github.com/lucasb-eyer/go-colorful"
)

// Frame represents a frame of RGB pixels to display on an ledrx device.
type Frame struct {
	pixels [50]colorful.Color
}

// NewFrame creates a new Frame instance.
func NewFrame() (*Frame) {
	f := new(Frame)
	return f
}

// MarshalBinary converts a Frame into binary data.
func (f *Frame) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 0, 152)
	binary.LittleEndian.PutUint16(data, 50)
	for _, p := range f.pixels {
		r, g, b := p.Clamped().RGB255()
		data = append(data, r, g, b)
	}

	return data, nil
}
