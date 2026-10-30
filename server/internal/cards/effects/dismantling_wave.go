package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "25012c1f-c069-43d7-b930-a3a5a7eb1389",
		Name:         "Dismantling Wave",
		Completeness: CompletenessFull,
		Targets:      perOpponentClauses(game.MaxPlayers-1, "artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve:    destroyEachLegalTarget,
		Activated:    []ActivatedAbility{Cycling("{6}{W}{W}")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(InGraveyard(On(game.EventCycle, Self,
				"Dismantling Wave — cycled: destroy all artifacts and enchantments",
				func(g *game.Game, item *game.StackItem) error {
					return DestroyAllMatching{Match: Or(Artifact(), Enchantment())}.Apply(NewContext(g, item))
				})),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifactsAndEnchantments, How: game.SweepDestroy}}),
		},
	})
}
