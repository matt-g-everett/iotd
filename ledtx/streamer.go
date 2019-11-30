package ledtx

import (
	"math"
	"time"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/eclipse/paho.mqtt.golang"
)

// GradientTable stores a look-up table of colours interpolated by hue.
type GradientTable []struct {
	Hue float64
	Pos float64
}

// GetColor gets a colour at the specified point on the look-up table.
func (g GradientTable) GetColor(t, s, l float64) colorful.Color {
	for i := 0; i < len(g)-1; i++ {
		c1 := g[i]
		c2 := g[i+1]
		if c1.Pos <= t && t <= c2.Pos {
			// We are in between c1 and c2. Go blend them!
			h := (((t - c1.Pos) / (c2.Pos - c1.Pos)) * (c2.Hue - c1.Hue)) + c1.Hue
			return colorful.Hcl(h, s, l)
		}
	}

	// Nothing found? Means we're at (or past) the last gradient keypoint.
	return colorful.Hcl(g[len(g)-1].Hue, 1.0, 0.05)
}


// Streamer that streams RGB data frames to an ledrx device.
type Streamer struct {
	client mqtt.Client
	hue float64
	rainbow GradientTable
}

// NewStreamer creates an instance of a Streamer.
func NewStreamer(client mqtt.Client) *Streamer {
	s := new(Streamer)
	s.client = client
	s.hue = 0.0
	s.rainbow = GradientTable{
		{0.0, 0.0},
		{6.0, 0.04}, // Pink
		{87.0, 0.14}, // Red
		{88.0, 0.28}, // Orange
		{98.0, 0.42}, // Yellow
		{180.0, 0.56}, // Green
		{185.0, 0.70}, // Turquiose
		{320.0, 0.84}, // Blue
		{332.0, 0.91}, // Violet
		{360.0, 1.0}, // Pink wrap
	}

	return s
}

// SendFrame sends a frame as binary over MQTT to an ledrx device.
func (s *Streamer) SendFrame() {
	f := NewFrame()

	hue := 0.0
	saturation := 1.0
	luminance := 0.05
	// hueIncrement := 1.0 / float64(len(f.pixels))
	for i := 0; i < len(f.pixels); i++ {
		c := s.rainbow.GetColor(hue, saturation, luminance)
		hue += 0.02
		hue = math.Mod(hue, 1.0)
		f.pixels[i] = c
	}

	b, _ := f.MarshalBinary()
	token := s.client.Publish("home/xmastree/stream", 0, false, b)
	token.Wait()
}

// Run causes the Streamer to send Frames continuously.
func (s *Streamer) Run() {
	publishTimer := time.NewTicker(33 * time.Millisecond)
	for {
		<-publishTimer.C
		s.SendFrame()
	}
}
