package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Admiral Beckett Brass — Legendary Creature — Human Pirate {1}{U}{B}{R}, 3/3:
//
//	"Other Pirates you control get +1/+1.
//	 At the beginning of your end step, gain control of target nonland
//	 permanent controlled by a player who was dealt combat damage by three
//	 or more Pirates this turn."
//
// The anthem is the tribal lord. The end-step ability reads the per-turn
// record of which creature objects dealt combat damage to which player,
// each with the creature types it had as it dealt the damage
// (TurnTally.DamageDealers, CountCreaturesOfTypeThatDealtCombatDamageToThisTurn):
// three DIFFERENT Pirates of any controller, still on the battlefield or
// not, and the damaged player is the permanent's controller. The target
// is chosen as the trigger goes on the stack (CR 603.3d) and is checked
// again as it resolves; the control change has no end (CR 611.2).
func init() {
	otherPirates := TribeFilter{Tribes: []string{"Pirate"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:     "43b5e462-d860-473d-828f-6c513fc7768a",
		Name:         "Admiral Beckett Brass",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{TribalAnthem(otherPirates, 1, 1)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourEndStep("Admiral Beckett Brass — gain control of target nonland permanent controlled by a player who was dealt combat damage by three or more Pirates this turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return GainControl{Target: id, Duration: game.IndefiniteDuration(),
							Label: "Admiral Beckett Brass — gain control"}.Apply(ctx)
					}),
				TargetPermanent("target nonland permanent controlled by a player who was dealt combat damage by three or more Pirates this turn",
					Nonland(), cmdr2B05ControlledByPlayerHitByThreePirates),
			),
		},
	})
}

// cmdr2B05ControlledByPlayerHitByThreePirates passes a permanent whose
// controller was dealt combat damage by three or more Pirates this turn.
func cmdr2B05ControlledByPlayerHitByThreePirates(g *game.Game, _ uuid.UUID, c game.Card) bool {
	return g.CountCreaturesOfTypeThatDealtCombatDamageToThisTurn(c.Controller, "Pirate") >= 3
}
