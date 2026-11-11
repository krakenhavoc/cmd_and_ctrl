package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barbflare Gremlin — Creature — Gremlin {3}{R}, 3/2:
//
//	"First strike, haste
//	 Whenever a player taps a land for mana, if this creature is tapped,
//	 that player adds one mana of any type that land produced. Then that
//	 land deals 1 damage to that player."
//
// Mana Flare's triggered mana ability (CR 605.1b, no stack), gated by the
// intervening "if this creature is tapped", read as the land is tapped.
// The rider is the one new seam (game.ManaTrigger.LandDamagesTapper): after
// the extra mana is added, the tapped land deals 1 damage to the player who
// tapped it, as a source of its own (so a land with a damage prevention or
// protection is read as such).
//
// No simplification.
func init() {
	gremlin := WheneverAPlayerTapsALandForMana(
		"Barbflare Gremlin — that player adds one mana of any type that land produced, then that land deals 1 damage to that player",
		AddsOneManaOfAnyTypeProduced())
	isLand := gremlin.AppliesTo
	gremlin.AppliesTo = func(prod game.ManaProduced, source *game.Card, g *game.Game) bool {
		return isLand(prod, source, g) && source != nil && source.Tapped
	}
	gremlin.LandDamagesTapper = 1
	Register(Spec{
		OracleID:        "bd7accec-ca8a-49c7-9ae4-fe97d99b4a34",
		Name:            "Barbflare Gremlin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		ManaTriggers:    []game.ManaTrigger{gremlin},
	})
}
