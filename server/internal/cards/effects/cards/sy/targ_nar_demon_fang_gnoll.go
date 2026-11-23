package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Targ Nar, Demon-Fang Gnoll — Legendary Creature — Gnoll {R}{G}, 2/2
// (Adventures in the Forgotten Realms):
//
//	"Pack tactics — Whenever Targ Nar attacks, if you attacked with
//	 creatures with total power 6 or greater this combat, attacking
//	 creatures get +1/+0 until end of turn.
//	 {2}{R}{G}: Double Targ Nar's power and toughness until end of turn."
//
// The set of attacking creatures is fixed as the trigger resolves (CR 611.2c).
// The activated ability is doubling (CR 701.10): +X/+Y with X and Y read as it
// resolves, a one-time reading (DoublePowerAndToughnessOfEachUntilEOT).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d53ef172-8c26-4469-94c6-bf7fe4d48dbf",
		Name:         "Targ Nar, Demon-Fang Gnoll",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82AfrB09PackTactics("Targ Nar — pack tactics: attacking creatures get +1/+0 until end of turn", func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.AttackingCreature(),
					Power: 1,
					Label: "Targ Nar — attacking creatures get +1/+0 until end of turn",
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}{G}: Double Targ Nar's power and toughness until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if effects.SourceIsNewObject(g, item) {
					return nil
				}
				return effects.DoublePowerAndToughnessOfEachUntilEOT(effects.NewContext(g, item), []uuid.UUID{item.SourceCardID},
					"Targ Nar — double power and toughness until end of turn")
			},
		}},
	})
}
