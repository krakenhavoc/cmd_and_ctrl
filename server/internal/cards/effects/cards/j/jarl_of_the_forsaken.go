package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jarl of the Forsaken — Creature — Zombie Cleric {3}{B}, 3/2 (Kaldheim):
//
//	"Flash
//	 When this creature enters, destroy target creature or planeswalker an
//	 opponent controls that was dealt damage this turn.
//	 Foretell {1}{B} (During your turn, you may pay {2} and exile this card
//	 from your hand face down. Cast it on a later turn for its foretell cost.)"
//
// Rooftop Assassin's trigger over creatures and planeswalkers: the target is
// chosen as the trigger goes on the stack under "was dealt damage this turn"
// (read from the turn's damage events since the permanent last entered), and
// with none the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "294c1fed-9473-4fbe-b10e-71a7a432b98d",
		Name:            "Jarl of the Forsaken",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		SpecialActions:  []game.SpecialAction{effects.Foretell("{1}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Jarl of the Forsaken — destroy target creature or planeswalker an opponent controls that was dealt damage this turn", effects.DestroyChosenPermanent),
				effects.TargetPermanent("target creature or planeswalker an opponent controls that was dealt damage this turn",
					effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls(), effects.DealtDamageThisTurn())),
		},
	})
}
