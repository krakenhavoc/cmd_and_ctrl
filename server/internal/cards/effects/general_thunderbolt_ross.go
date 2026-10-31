package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// General Thunderbolt Ross — Legendary Creature — Human Soldier {W},
// 1/1 (Marvel Super Heroes Commander, #4455):
//
//	"First strike
//	 Battalion — Whenever General Thunderbolt Ross and at least two
//	 other creatures attack, attacking creatures get +1/+0 until end of
//	 turn."
//
// Battalion is Boros Strike-Captain's trigger (battalionAttack). The
// pump is applied to the creatures attacking as the trigger resolves
// (CR 611.2c), Ross himself included, and does not reach a creature that
// joins the attack later.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "57ffb558-3a1c-43ea-a79c-bcfacad22ca4",
		Name:            "General Thunderbolt Ross",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, battalionAttack,
				"General Thunderbolt Ross — attacking creatures get +1/+0 until end of turn (battalion)",
				Do(BoostUntilEOT{Match: AttackingCreature(), Power: 1, Label: "General Thunderbolt Ross — battalion"})),
		},
	})
}
