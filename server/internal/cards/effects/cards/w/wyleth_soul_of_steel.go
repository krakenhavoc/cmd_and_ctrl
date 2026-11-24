package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wyleth, Soul of Steel — Legendary Creature — Human Warrior {1}{R}{W},
// 2/2 (Commander Legends):
//
//	"Trample
//	 Whenever Wyleth attacks, draw a card for each Aura and Equipment
//	 attached to it."
//
// The attack trigger counts the Auras and Equipment attached to Wyleth as it
// resolves (CR 608.2c): every attachment of those two kinds, whoever controls
// it. With none attached, or with Wyleth gone, it draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "35bef4d3-4473-4265-86b0-f1393c82fd1d",
		Name:            "Wyleth, Soul of Steel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Wyleth, Soul of Steel — draw a card for each Aura and Equipment attached to it",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					if !effects.B09SourceStillOnBattlefield(ctx.Game, ctx.Item) {
						return nil
					}
					n := 0
					for _, id := range ctx.Game.AttachmentsOf(ctx.Source()) {
						if c, ok := ctx.Game.LookupCardForEffect(id); ok && (c.HasSubtype("Aura") || c.HasSubtype("Equipment")) {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return effects.DrawCards{Player: ctx.Controller(), N: n}.Apply(ctx)
				}))),
		},
	})
}
