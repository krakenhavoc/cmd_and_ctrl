package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mind Raker — Creature — Eldrazi Processor {3}{B}, 3/3 (Battle for Zendikar,
// #5547):
//
//	"Devoid (This card has no color.)
//	 When this creature enters, you may put a card an opponent owns from exile
//	 into that player's graveyard. If you do, each opponent discards a card."
//
// The processing is Ulamog's Nullifier's, for one card; each opponent then
// chooses their own discard (CR 701.9a) only if a card landed (CR 608.2c).
//
// No simplification.
func init() {
	const name = "Mind Raker"
	effects.Register(effects.Spec{
		OracleID:        "060076e7-3705-4f8e-8fac-59402835902b",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.OldB305WhenThisEntersMayProcess(name, "mind-raker", "each opponent discards a card", nil,
				func(ctx *effects.Context) error {
					effects.EachOpponentDiscards(ctx, 1)
					return nil
				}),
		},
	})
}
