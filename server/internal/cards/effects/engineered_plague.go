package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Engineered Plague — Enchantment {2}{B} (Urza's Legacy, #5623):
//
//	"As this enchantment enters, choose a creature type.
//	 All creatures of the chosen type get -1/-1."
//
// Steely Resolve's chosen type with a -1/-1 lord half: every creature of the
// chosen type, any player's, Engineered Plague's controller's included, gets
// -1/-1 in layer 7c. One left at 0 toughness is put into its owner's
// graveyard as a state-based action (CR 704.5f). Until the type is chosen
// it shrinks nothing. It is declared as a partial -1/-1 sweep of creatures
// (ADR 0126): it kills every X/1 of one creature type.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc8210b4-8e17-4156-994b-e95e215e3e34",
		Name:         "Engineered Plague",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, Partial: true}},
		AsEnters:     ChooseCreatureTypeAsEnters("Engineered Plague"),
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Chosen: true}, -1, -1),
		},
	})
}
