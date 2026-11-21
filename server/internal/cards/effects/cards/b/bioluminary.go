package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bioluminary — Creature — Merfolk Performer {2}{U}, 1/3:
//
//	"Whenever Bioluminary deals combat damage to a player, you get
//	 {TK}{TK}, then you may put a sticker on a nonland permanent you own."
//
// The tickets are a player counter (CR 107.17); the sticker is the
// engine's keyed questions (ADR 0208 §4), its ticket cost paid as it goes
// on, so the tickets this trigger gives can pay for it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb8b8c58-7bdc-4daa-b48b-728a26e174dd",
		Name:         "Bioluminary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(
				"Bioluminary — you get {TK}{TK}, then you may put a sticker on a nonland permanent you own",
				effects.GetTicketsThenSticker(2)),
		},
	})
}
