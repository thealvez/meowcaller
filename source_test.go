package meowcaller

import (
	"errors"
	"io"
	"testing"
)

func TestNextOpusAudioPacketSkipsTagsAndEmptyPackets(t *testing.T) {
	packets := [][]byte{nil, []byte("OpusTags\x00vendor"), {0x48, 0x83, 0x01}}
	index := 0
	next := func() ([]byte, error) {
		if index == len(packets) {
			return nil, io.EOF
		}
		packet := packets[index]
		index++
		return packet, nil
	}

	got, err := nextOpusAudioPacket(next)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(packets[2]) {
		t.Fatalf("got packet %x, want %x", got, packets[2])
	}
}

func TestNextOpusAudioPacketPropagatesReadError(t *testing.T) {
	want := errors.New("read failed")
	_, got := nextOpusAudioPacket(func() ([]byte, error) { return nil, want })
	if !errors.Is(got, want) {
		t.Fatalf("got error %v, want %v", got, want)
	}
}
