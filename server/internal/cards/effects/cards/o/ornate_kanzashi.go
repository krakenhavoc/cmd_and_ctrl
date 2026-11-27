package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ornate Kanzashi — Artifact {5} (Betrayers of Kamigawa, #5602):
//
//	"{2}, {T}: Target opponent exiles the top card of their library. You
//	 may play that card this turn."
//
// Knacksaw Clique's impulse exile from another library
// (TargetOpponentExilesTopYouMayPlay): the opponent's top card is exiled and
// you, not they, may play it until end of turn, a land as well as a spell,
// with the mana paid as normal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c41c666-5cf4-4f35-9cb8-c41f1f78f31e",
		Name:         "Ornate Kanzashi",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Target opponent exiles the top card of their library. You may play that card this turn.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect:  effects.TargetOpponentExilesTopYouMayPlay,
		}},
	})
}
