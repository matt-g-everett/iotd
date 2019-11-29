package ledtx

import (
	"log"
	"time"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/eclipse/paho.mqtt.golang"
)

// Streamer that streams RGB data frames to an ledrx device.
type Streamer struct {
	client mqtt.Client
}

// NewStreamer creates an instance of a Streamer.
func NewStreamer(client mqtt.Client) *Streamer {
	s := new(Streamer)
	s.client = client
	return s
}

// SendFrame sends a frame as binary over MQTT to an ledrx device.
func (s *Streamer) SendFrame() {
	f := NewFrame()

	for i := 0; i < len(f.pixels); i++ {
		c, _ := colorful.Hex("#FF0000")
		f.pixels[i] = c
	}

	b, _ := f.MarshalBinary()
	token := s.client.Publish("home/xmastree/stream", 0, false, b)
	pubToken := token.(*mqtt.PublishToken)
	log.Println("PUBLISH MSGID: ", pubToken.MessageID())
	token.Wait()
}

// Run causes the Streamer to send Frames continuously.
func (s *Streamer) Run() {
	publishTimer := time.NewTicker(1 * time.Second)
	for {
		<-publishTimer.C
		//s.SendFrame()
		log.Println("Sending frame...")
	}
}
