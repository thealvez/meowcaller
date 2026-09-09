package meowcaller

import (
	"math"
	"testing"

	"github.com/rs/zerolog"
)

func TestOpusWireCodecRoundTrip60msWideband(t *testing.T) {
	encoder, decoder, err := newWireAudioCodecs(AudioCodecOpus, zerolog.Nop())
	if err != nil {
		t.Fatal(err)
	}

	want := make([]float32, FrameSamples)
	for i := range want {
		want[i] = 0.35 * float32(math.Sin(2*math.Pi*440*float64(i)/SampleRate))
	}
	payload, err := encoder.Encode(want)
	if err != nil {
		t.Fatal(err)
	}
	if len(payload) == 0 || len(payload) > maxOpusPacketBytes {
		t.Fatalf("unexpected opus payload length %d", len(payload))
	}

	got, err := decoder.Decode(payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != FrameSamples {
		t.Fatalf("decoded %d samples, want %d", len(got), FrameSamples)
	}
	if rmsFloat32(got) < 0.05 {
		t.Fatalf("decoded opus frame is effectively silent: rms=%f", rmsFloat32(got))
	}
}

func TestMlowWireCodecPathRemainsAvailable(t *testing.T) {
	encoder, decoder, err := newWireAudioCodecs(AudioCodecMlow, zerolog.Nop())
	if err != nil {
		t.Fatal(err)
	}

	want := make([]float32, FrameSamples)
	for i := range want {
		want[i] = 0.35 * float32(math.Sin(2*math.Pi*550*float64(i)/SampleRate))
	}
	payload, err := encoder.Encode(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := decoder.Decode(payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != FrameSamples {
		t.Fatalf("decoded %d samples, want %d", len(got), FrameSamples)
	}
	if rmsFloat32(got) < 0.05 {
		t.Fatalf("decoded mlow frame is effectively silent: rms=%f", rmsFloat32(got))
	}
}

func TestOpusWireEncoderRejectsWrongFrameSize(t *testing.T) {
	encoder, err := newOpusWireEncoder()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := encoder.Encode(make([]float32, FrameSamples-1)); err == nil {
		t.Fatal("expected wrong frame size to fail")
	}
}

func TestNewWireAudioCodecsRejectsUnknownCodec(t *testing.T) {
	if _, _, err := newWireAudioCodecs(AudioCodec(99), zerolog.Nop()); err == nil {
		t.Fatal("expected unknown codec to fail")
	}
}
