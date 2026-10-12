package main

// The ollie at the end of the direct load: tail snap, then the board
// landing, made from noise and a low thump (no sound files from the game).

import (
	"encoding/binary"
	"math"
	"math/rand"
)

func ollieWav() []byte {
	const rate = 44100
	n := rate * 45 / 100
	pcm := make([]float64, n)
	hit := func(at float64, noiseMs, thumpHz, thumpMs, gain float64) {
		s := int(at * rate)
		for i := 0; s+i < n; i++ {
			t := float64(i) / rate
			v := 0.0
			if t < noiseMs/1000*4 {
				v += (rand.Float64()*2 - 1) * math.Exp(-t/(noiseMs/1000))
			}
			v += math.Sin(2*math.Pi*thumpHz*t) * math.Exp(-t/(thumpMs/1000)) * 0.9
			pcm[s+i] += v * gain
		}
	}
	hit(0, 6, 95, 45, 0.9)     // pop
	hit(0.27, 9, 70, 60, 0.75) // landing
	out := make([]byte, 44+n*2)
	copy(out, "RIFF")
	binary.LittleEndian.PutUint32(out[4:], uint32(36+n*2))
	copy(out[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(out[16:], 16)
	binary.LittleEndian.PutUint16(out[20:], 1)
	binary.LittleEndian.PutUint16(out[22:], 1)
	binary.LittleEndian.PutUint32(out[24:], rate)
	binary.LittleEndian.PutUint32(out[28:], rate*2)
	binary.LittleEndian.PutUint16(out[32:], 2)
	binary.LittleEndian.PutUint16(out[34:], 16)
	copy(out[36:], "data")
	binary.LittleEndian.PutUint32(out[40:], uint32(n*2))
	for i, v := range pcm {
		v = math.Max(-1, math.Min(1, v*0.6))
		binary.LittleEndian.PutUint16(out[44+i*2:], uint16(int16(v*32767)))
	}
	return out
}
