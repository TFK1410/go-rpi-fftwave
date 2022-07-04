package main

import (
	"math/rand"
	"testing"
	"time"
)

func BenchmarkDrawImage(b *testing.B) {

	ldc := initContext(128, 64)

	ldc.startSqlite()

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

func BenchmarkGlitch(b *testing.B) {

	ldc := initContext(128, 64)

	ldc.startSqlite()

	glitchSeed = time.Now().UTC().UnixNano()
	rand.Seed(glitchSeed)

	l := ldc.getLyric(1)

	prgs := byte(200)

	if l.AlignVisible {
		ldc.drawTextRealign(l, prgs)
	} else {
		ldc.drawText(l, prgs)
	}

	for i := 0; i < b.N; i++ {
		if l.Glitch >= 0.1 {
			glitchImage(ldc.img, l.Glitch, l.GlitchColor)
		}
	}
}
