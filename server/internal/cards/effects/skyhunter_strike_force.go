package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skyhunter Strike Force — Creature — Cat Knight {2}{W}, 2/2 (March of
// the Machine: Phyrexia: All Will Be One Commander):
//
//	"Flying
//	 Melee (Whenever this creature attacks, it gets +1/+1 until end of turn
//	 for each opponent you attacked this combat.)
//	 Lieutenant — As long as you control your commander, other creatures you
//	 control have melee."
//
// Flying and melee are the engine's keywords (CR 702.121, game/melee.go).
// The lieutenant line is a layer 6 grant to every OTHER creature its
// controller has while they control a commander they own (b18ControlsYourCommander,
// re-read on every layer recompute); the Force keeps its own printed melee
// either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8ade0824-9fd0-417d-94dd-ac399e35b582",
		Name:            "Skyhunter Strike Force",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordMelee},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.InstanceID != source.InstanceID &&
					b18ControlsYourCommander(g, source.Controller)
			}, game.KeywordMelee),
		},
	})
}
