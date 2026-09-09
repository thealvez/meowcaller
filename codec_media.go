package meowcaller

import (
	"fmt"
	"math"

	"github.com/pion/opus"
	"github.com/purpshell/meowcaller/mlow"
	"github.com/rs/zerolog"
)

const maxOpusPacketBytes = 1275

type wireAudioEncoder interface {
	Encode([]float32) ([]byte, error)
}

type wireAudioDecoder interface {
	Decode([]byte) ([]float32, error)
}

type mlowWireEncoder struct {
	encoder *mlow.MlowEncoder
}

func (e *mlowWireEncoder) Encode(frame []float32) ([]byte, error) {
	return e.encoder.Encode(frame)
}

type mlowWireDecoder struct {
	decoder *mlow.MlowDecoder
}

func (d *mlowWireDecoder) Decode(payload []byte) ([]float32, error) {
	return d.decoder.Decode(payload), nil
}

type opusWireEncoder struct {
	encoder *opus.Encoder
	pcm     []int16
	packet  []byte
}

func newOpusWireEncoder() (*opusWireEncoder, error) {
	encoder, err := opus.NewEncoder(
		opus.WithApplication(opus.ApplicationVoIP),
		opus.WithBitrate(24000),
	)
	if err != nil {
		return nil, fmt.Errorf("meowcaller: create opus encoder: %w", err)
	}
	return &opusWireEncoder{
		encoder: encoder,
		pcm:     make([]int16, FrameSamples),
		packet:  make([]byte, maxOpusPacketBytes),
	}, nil
}

func (e *opusWireEncoder) Encode(frame []float32) ([]byte, error) {
	if len(frame) != FrameSamples {
		return nil, fmt.Errorf("meowcaller: opus frame has %d samples, want %d", len(frame), FrameSamples)
	}
	for i, sample := range frame {
		sample = max(-1, min(1, sample))
		e.pcm[i] = int16(math.Round(float64(sample * 32767)))
	}
	n, err := e.encoder.EncodeSILK(e.pcm, opus.BandwidthWideband, e.packet)
	if err != nil {
		return nil, fmt.Errorf("meowcaller: opus encode: %w", err)
	}
	payload := make([]byte, n)
	copy(payload, e.packet[:n])
	return payload, nil
}

type opusWireDecoder struct {
	decoder opus.Decoder
	pcm     []float32
}

func newOpusWireDecoder() (*opusWireDecoder, error) {
	decoder, err := opus.NewDecoderWithOutput(SampleRate, 1)
	if err != nil {
		return nil, fmt.Errorf("meowcaller: create opus decoder: %w", err)
	}
	return &opusWireDecoder{decoder: decoder, pcm: make([]float32, FrameSamples*2)}, nil
}

func (d *opusWireDecoder) Decode(payload []byte) ([]float32, error) {
	n, err := d.decoder.DecodeToFloat32(payload, d.pcm)
	if err != nil {
		return nil, fmt.Errorf("meowcaller: opus decode: %w", err)
	}
	frame := make([]float32, n)
	copy(frame, d.pcm[:n])
	return frame, nil
}

func newWireAudioCodecs(codec AudioCodec, log zerolog.Logger) (wireAudioEncoder, wireAudioDecoder, error) {
	switch codec {
	case AudioCodecMlow:
		return &mlowWireEncoder{encoder: mlow.NewMlowEncoder(mlow.WithLogger(log))},
			&mlowWireDecoder{decoder: mlow.NewMlowDecoder(mlow.WithLogger(log))}, nil
	case AudioCodecOpus:
		encoder, err := newOpusWireEncoder()
		if err != nil {
			return nil, nil, err
		}
		decoder, err := newOpusWireDecoder()
		if err != nil {
			return nil, nil, err
		}
		return encoder, decoder, nil
	default:
		return nil, nil, fmt.Errorf("meowcaller: unsupported audio codec %d", codec)
	}
}
