package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Shaken Faith — Enchantment — Aura Curse {1}{R}:
//
//	"Enchant player
//	 Whenever enchanted player casts a spell other than the first spell
//	 they cast each turn or copies a spell, this Aura deals 2 damage to
//	 them."
//
// One printed trigger over two events. The cast half reads the per-turn
// cast tally, which the cast path bumps BEFORE it emits EventCast, so a
// total above one means it is not that player's first spell (Breeches'
// clock). The copy half is EventSpellCopied, whose Actor is the copy's
// controller (CR 707.10): a copy is created, not cast, so it never
// counts toward the tally and never triggers the first half. The Aura
// deals the damage, and a Curse that has left deals it as it last
// existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "297494f0-1f79-4f77-b68b-6ed858e21f4e",
		Name:         "Curse of Shaken Faith",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventCast, game.EventSpellCopied},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if source.AttachedTo.Kind != game.TargetPlayer || ev.Actor != source.AttachedTo.ID {
						return false
					}
					switch ev.Kind {
					case game.EventCast:
						return g.CastTallyFor(ev.Actor).Total >= 2
					case game.EventSpellCopied:
						return true
					}
					return false
				},
				"Curse of Shaken Faith — it deals 2 damage to them",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					victim := ctx.Item.Trigger.Event.Actor
					if ctx.Game.PlayerByIDForEffect(victim) == nil {
						return nil
					}
					return ctx.Game.DealDamageToPlayerForEffect(ctx.Source(), victim, 2)
				}))),
		},
	})
}
