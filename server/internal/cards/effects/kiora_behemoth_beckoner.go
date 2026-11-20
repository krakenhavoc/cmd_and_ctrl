package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kiora, Behemoth Beckoner — Legendary Planeswalker — Kiora {2}{G/U}, loyalty 7 (War of the Spark, #5520):
//
//	"Whenever a creature you control with power 4 or greater enters, draw a card.
//	 −1: Untap target permanent."
//
// Starting loyalty is printed data (ADR 0032). The trigger is Garruk's Uprising's draw for a
// creature with power 4 or greater entering under your control; the −1
// untaps any permanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1efe485e-6874-4f4c-b981-e10f41670865",
		Name:         "Kiora, Behemoth Beckoner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, kioraBigCreatureEntered, "Kiora, Behemoth Beckoner — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "−1: Untap target permanent.",
			Cost:    LoyaltyCost(-1),
			Targets: TargetPermanent("target permanent"),
			Effect:  untapFirstLegalTarget,
		}},
	})
}

// kioraBigCreatureEntered is "whenever a creature you control with power
// 4 or greater enters", the power read as it enters.
func kioraBigCreatureEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsCreature() && c.CurrentPower() >= 4
}
