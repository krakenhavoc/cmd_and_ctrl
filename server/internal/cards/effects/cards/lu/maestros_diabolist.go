package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maestros Diabolist — Creature — Vampire Warrior {U}{B}{R}, 1/4 (Streets
// of New Capenna, #5491):
//
//	"Deathtouch, haste
//	 Whenever this creature attacks, if you don't control a Devil token,
//	 create a tapped and attacking 1/1 red Devil creature token with "When
//	 this token dies, it deals 1 damage to any target.""
//
// An intervening "if" (CR 603.4): you must control no Devil token as the
// attack is declared and again as the trigger resolves, so two Diabolists
// attacking together each trigger but only the first to resolve makes a
// Devil. Only a TOKEN Devil counts; a Devil card does not stop it. The
// token is Spiked Corridor's printed Devil (the dies trigger is on the
// token). It enters tapped and attacking with no player named, so its
// controller chooses which player or planeswalker it attacks
// (CR 508.4), preferring what the Diabolist attacked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68d61090-a97f-47e9-bc64-7e5c7f8acbf7",
		Name:            "Maestros Diabolist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.Snc17NoDevilToken, effects.WheneverThisAttacks(
				"Maestros Diabolist — create a tapped and attacking 1/1 red Devil creature token",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokensAttackingYourChoice{
						Template: effects.DevilToken(),
						N:        1,
						Tapped:   true,
						Prefer:   effects.TriggerAttackTarget(item),
						Label:    "Maestros Diabolist",
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
