package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tolarian Terror — Creature — Serpent {6}{U}, 5/5 (Foundations, #4465):
//
//	"This spell costs {1} less to cast for each instant and sorcery card
//	 in your graveyard.
//	 Ward {2}"
//
// The Dawning Archaic's self cost reduction, counting each card in your
// graveyard that is an instant or a sorcery; it reduces only the generic
// part, so the Terror never costs less than {U}. Ward is the shared Ward
// trigger (Aboleth Spawn's ward {2}).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9d0f2e1-62c2-44fd-ad38-471daf17bb0a",
		Name:         "Tolarian Terror",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(socB37InstantAndSorceryCardsInYourGraveyard,
				"This spell costs {1} less to cast for each instant and sorcery card in your graveyard."),
		},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Tolarian Terror — ward {2}"),
		},
	})
}
