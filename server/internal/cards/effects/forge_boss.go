package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forge Boss — Creature — Human Warrior {2}{B}{R}, 3/4 (Streets of New
// Capenna, S82):
//
//	"Whenever you sacrifice one or more other creatures, this creature
//	 deals 2 damage to each opponent. This ability triggers only once each
//	 turn."
//
// "One or more" is one trigger per batch of simultaneous sacrifices
// (OncePerBatch), and "only once each turn" shuts every later batch of the
// turn out (TriggersOnlyOnceEachTurn, #5245). The Boss itself does not
// count; the damage comes from the Boss as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c234d2ab-21d5-41a9-9ab8-47a8bd53a686",
		Name:         "Forge Boss",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(OncePerBatch(On(game.EventSacrifice,
				sncB13YouSacrificedAnotherCreature,
				"Forge Boss — 2 damage to each opponent",
				Do(DamageEachOpponent{N: 2})))),
		},
	})
}
