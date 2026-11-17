package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Churning Reservoir — Artifact {R}:
//
//	"At the beginning of your upkeep, put an oil counter on another target
//	 nontoken artifact or creature you control.
//	 {2}, {T}: Create a 1/1 red Phyrexian Goblin creature token. Activate
//	 only if an oil counter was removed from a permanent you controlled this
//	 turn or a permanent with an oil counter on it was put into a graveyard
//	 this turn."
//
// The gate reads this turn's event log (s82OneB12OilEventsThisTurn): an oil
// removal is judged by the controller of the permanent it came off, a
// permanent put into a graveyard by the oil it had as it left, whoever
// controlled it. Both are read before any cost is paid (CR 602.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d1c5987-3572-4bf9-b01c-eb6e98a7da77",
		Name:         "Churning Reservoir",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourUpkeep("Churning Reservoir — put an oil counter on another target nontoken artifact or creature you control",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 {
							return nil
						}
						return AddCounter{Target: item.Targets[0].ID, Kind: s82OneB12Oil, N: 1}.Apply(NewContext(g, item))
					}),
				Another(TargetPermanent("another target nontoken artifact or creature you control",
					Or(Artifact(), Creature()), Not(IsTokenPredicate()), YouControl()))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Create a 1/1 red Phyrexian Goblin creature token. Activate only if an oil counter was removed from a permanent you controlled this turn or a permanent with an oil counter on it was put into a graveyard this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1, CreatureTokens: 1},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return s82OneB12OilEventsThisTurn(g, controller)
			},
			Effect: Do(CreateToken{Template: TokenCard("1/1 red Phyrexian Goblin"), N: 1}),
		}},
	})
}
