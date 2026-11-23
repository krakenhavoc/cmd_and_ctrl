package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dismantling Wave — Sorcery {2}{W} (Marvel Super Heroes Commander):
//
//	"For each opponent, destroy up to one target artifact or enchantment
//	 that player controls.
//	 Cycling {6}{W}{W}
//	 When you cycle this card, destroy all artifacts and enchantments."
//
// Windgrace's Judgment's per-opponent "up to one" clauses, one per seat in
// turn order, each bound to its opponent. The cycling trigger is Fractured
// Sanity's (a graveyard trigger on the cycle event) and destroys the whole
// board's artifacts and enchantments, yours included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25012c1f-c069-43d7-b930-a3a5a7eb1389",
		Name:         "Dismantling Wave",
		Completeness: effects.CompletenessFull,
		Targets:      effects.PerOpponentClauses(game.MaxPlayers-1, "artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve:    effects.DestroyEachLegalTarget,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{6}{W}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.InGraveyard(effects.On(game.EventCycle, effects.Self,
				"Dismantling Wave — cycled: destroy all artifacts and enchantments",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyAllMatching{Match: effects.Or(effects.Artifact(), effects.Enchantment())}.Apply(effects.NewContext(g, item))
				})),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifactsAndEnchantments, How: game.SweepDestroy}}),
		},
	})
}
