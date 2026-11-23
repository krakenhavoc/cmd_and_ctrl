package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cunning Bandit // Azamuki, Treachery Incarnate — {1}{R}{R} Creature —
// Human Warrior 2/2 // Legendary Creature — Spirit 5/2, a flip card
// (#3216, ADR 0151):
//
//	Top:     "Whenever you cast a Spirit or Arcane spell, you may put a
//	          ki counter on this creature.
//	          At the beginning of the end step, if there are two or more
//	          ki counters on this creature, you may flip it."
//	Flipped: "Remove a ki counter from Azamuki: Gain control of target
//	          creature until end of turn."
//
// The theft is Act of Treason's layer-2 until-end-of-turn control
// effect (ADR 0063), without the untap and haste that card adds: Azamuki
// does not print them. Azamuki is red and {1}{R}{R} (CR 710.1c).
//
// No simplification.
const CunningBanditOracleID = "3d416f83-32b8-4dad-8e4d-3acd182bbe44"

func init() {
	effects.Register(effects.Spec{
		OracleID:     CunningBanditOracleID,
		Name:         "Cunning Bandit",
		Completeness: effects.CompletenessFull,
		Triggered:    effects.KiFlipTriggers("Cunning Bandit"),
	})
	effects.Register(effects.Spec{
		OracleID:     CunningBanditOracleID + "#1",
		Name:         "Azamuki, Treachery Incarnate",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove a ki counter from Azamuki: Gain control of target creature until end of turn.",
			Cost:    effects.RemoveAKiCounter(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
					return effects.GainControl{
						Target:   id,
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Azamuki, Treachery Incarnate — gain control until end of turn",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
