package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oath of Chandra — Legendary Enchantment {1}{R} (Oath of the Gatewatch,
// #5545):
//
//	"When Oath of Chandra enters, it deals 3 damage to target creature an
//	 opponent controls.
//	 At the beginning of each end step, if a planeswalker entered the
//	 battlefield under your control this turn, Oath of Chandra deals 2 damage to
//	 each opponent."
//
// The second ability is an intervening "if" (CR 603.4) on the turn's entry
// record, read as each end step begins and again as the trigger resolves:
// planeswalkers that entered under you this turn, as they entered, even if
// they have since left. It triggers on every end step, not only yours, so it
// can fire on an opponent's turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53a0923f-7da8-4e31-a231-9b275916e2a2",
		Name:         "Oath of Chandra",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Oath of Chandra — 3 damage to target creature an opponent controls", effects.B33DamageChosenTargetFromSource(3)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
			effects.InterveningIfYou("a planeswalker entered the battlefield under your control this turn",
				func(g *game.Game, you uuid.UUID) bool { return g.EnteredWithCardTypeThisTurn(you, "planeswalker") > 0 },
				effects.On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Kind == game.EventBeginEndStep
				}, "Oath of Chandra — 2 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 2}))),
		},
	})
}
