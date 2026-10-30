package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Armed Assailant — {2}{B} Creature — Human Villain 1/3 (Marvel Super
// Heroes Commander, #4455):
//
//	"Deathtouch
//	 As long as this creature is equipped, it gets +2/+0 and has menace."
//
// Cloud, Planet's Champion's "as long as Cloud is equipped" gate
// (isEquippedCreature, read on every layer recompute) on a layer 7c
// pump and a layer 6 keyword grant, both on the creature alone.
//
// No simplification.
func init() {
	whileEquipped := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfOnly(target, g, source) && isEquippedCreature(g, *target)
	}
	pump := b16Anthem(whileEquipped, 2, 0)
	pump.Label = "As long as this creature is equipped, it gets +2/+0."
	menace := KeywordGrant(whileEquipped, "menace")
	menace.Label = "As long as this creature is equipped, it has menace."
	Register(Spec{
		OracleID:        "53d7722a-b60b-4520-abce-f31b1e6bba00",
		Name:            "Armed Assailant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Static:          []game.StaticAbility{pump, menace},
	})
}
