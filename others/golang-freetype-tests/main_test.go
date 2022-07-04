package main

import (
	"math/rand"
	"testing"
	"time"
)

func BenchmarkDrawText(b *testing.B) {
	ldc := initContext(128, 64, "RobotoMono-Light.ttf")

	ldc.startSqlite()

	// a separate seed thread makes it so that the glitch shifts are changed only
	// by the defined ticker time instead of with every frame
	glitchSeed = time.Now().UTC().UnixNano()
	rand.Seed(glitchSeed)

	l := ldc.getLyric(1)

	prgs := byte(200)

	for i := 0; i < b.N; i++ {
		if l.AlignVisible {
			ldc.drawTextRealign(l, prgs)
		} else {
			ldc.drawText(l, prgs)
		}
		if l.Glitch >= 0.1 {
			glitchImage(ldc.img, l.Glitch, l.GlitchColor)
		}
	}
}
