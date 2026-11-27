package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Impatience — {2}{R} Enchantment (Urza's Destiny, S82 OLD batch 704, #5621):
//
//	"At the beginning of each player's end step, if that player didn't cast a
//	 spell this turn, this enchantment deals 2 damage to that player."
//
// The intervening "if" (CR 603.4) is read for the turn's player as the end
// step begins and again as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97c8f5e1-9fc1-4a62-a757-1c0454754cd3",
		Name:         "Impatience",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(
				effects.CondBoard{Name: "that player didn't cast a spell this turn", Test: effects.S82OldB704DidntCastASpellThisTurn},
				effects.On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor != uuid.Nil && g.CastTallyFor(ev.Actor).Total == 0
				}, "Impatience — it deals 2 damage to that player", effects.S82OldB704ImpatienceDamage)),
		},
	})
}
