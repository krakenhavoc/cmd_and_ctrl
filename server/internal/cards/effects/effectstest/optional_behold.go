package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// optional_behold_test.go — #4642: "As an additional cost to cast this
// spell, you may behold a <type>" (CR 701.4), tested on its two cards:
// Hulk's Thunderclap (a Gamma; the third target exists only when it was
// paid) and Hyperion's Atomic Vision (a Hero; "if a Hero was beheld").

const (
	ThunderclapOracle    = "e377059d-a918-462f-920a-6b062bd5fcfd"
	AtomicVisionOracle   = "e1e82283-01ec-447f-b0d4-5bd4619ab3d8"
	GammaTypeLine        = "Creature — Gamma Human"
	HeroTypeLine         = "Creature — Hero Human"
	BeholdOfferIdx       = 0
	ThunderclapManaCost  = "{1}{G}"
	AtomicVisionManaCost = "{1}{W}"
)

func Slotted(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, len(ids))
	for i, id := range ids {
		out[i] = game.TargetRef{Kind: game.TargetCard, ID: id, Slot: i}
	}
	return out
}

func StackLen(g *game.Game) int {
	n := 0
	g.ReadSnapshot(func() { n = len(g.Stack.Cards) })
	return n
}
