package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const cunningBanditOracleID = "3d416f83-32b8-4dad-8e4d-3acd182bbe44"

func init() {
	Register(Spec{
		OracleID:     cunningBanditOracleID,
		Name:         "Cunning Bandit",
		Completeness: CompletenessFull,
		Triggered:    kiFlipTriggers("Cunning Bandit"),
	})
	Register(Spec{
		OracleID:     cunningBanditOracleID + "#1",
		Name:         "Azamuki, Treachery Incarnate",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Remove a ki counter from Azamuki: Gain control of target creature until end of turn.",
			Cost:    removeAKiCounter(),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return GainControl{
						Target:   id,
						Duration: DurationUntilEndOfTurn(ctx),
						Label:    "Azamuki, Treachery Incarnate — gain control until end of turn",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
