package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// harmonize_cards_test.go — #3931, CR 702.180: "You may cast this card
// from your graveyard by paying [cost] and tapping up to one untapped
// creature you control rather than paying this spell's mana cost", reduced
// by an amount of generic mana equal to the tapped creature's power, and
// exiled any time it would leave the stack.

const (
	NaturesRhythmOracle      = "0259f063-3e3f-4e07-9066-f748d4819829"
	SynchronizedChargeOracle = "faa0055b-c1d2-4ce7-875f-41b4be8030bd"
	ChanneledDragonfireOracl = "91179937-6f55-4d1b-8f06-2efe1c8fea7a"
	SongcrafterMageOracle    = "13f0863f-ae54-4bea-8f01-7f3891797577"
	MammothBellowOracle      = "6f2e4607-bece-42ba-a2a3-5dc50f246793"
	WinternightStoriesOracle = "d64687e4-ec37-4db2-9e60-4626c7a3b011"
	UrenisRebuffOracle       = "c0b7e989-8194-484c-93aa-a44df0764a77"
	ZenithFestivalOracle     = "4dabba8c-0490-41ec-a03e-f7d0c9f0edb6"
	GlacialDragonhuntOracle  = "8f976399-29a9-4eb7-844d-3d3f2fdf6ea7"
	WildRideOracle           = "61978202-538b-4f92-acfc-e95f4a6ce466"
	RoamersRoutineOracle     = "cf2355f3-11d2-4b03-b663-9fad6b07e4b2"
	UnendingWhisperOracle    = "c5da5c31-0cf1-4393-8701-eda7880ecebc"
)

// harmonizeTable is a main phase on seat 0's turn with an empty hand.
func HarmonizeTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceToMain(t, g)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return g, me, g.Seats[1]
}
