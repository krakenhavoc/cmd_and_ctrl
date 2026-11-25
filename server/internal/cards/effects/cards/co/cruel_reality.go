package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cruel Reality — Enchantment — Aura Curse {5}{B}{B} (Amonkhet, S82 OLD batch 244, #5535):
//
//	"Enchant player
//	 At the beginning of enchanted player's upkeep, that player sacrifices a
//	 creature or planeswalker of their choice. If the player can't, they lose 5
//	 life."
//
// The trigger fires at the enchanted player's upkeep only. That player is asked
// for a creature or planeswalker of their own choice (Deadly Brew's sacrifice
// run); if they have none to give, or none was sacrificed, they lose 5 life
// (CR 701.21a: a sacrifice that can't happen isn't paid).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e3cc9b6-7849-4385-a487-c8f5364174e7",
		Name:         "Cruel Reality",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Actor
			}, "Cruel Reality — that player sacrifices a creature or planeswalker; if they can't, they lose 5 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					victim := item.Trigger.Event.Actor
					return g.PlayersSacrificeThenKeyedForEffect(ctx.Source(), []uuid.UUID{victim},
						effects.SacrificeSpec("a creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
						"Cruel Reality — sacrifice a creature or planeswalker", cruelRealityLose, ctx.RunData([]uuid.UUID{victim}))
				}),
		},
	})
}

// cruelRealityLose is "if the player can't, they lose 5 life": the carry's
// first ID is the enchanted player, who loses the life when the run sacrificed
// nothing for them. An on-disk key: never renamed, never reused.
var cruelRealityLose = effects.RunThen("prompt-run/s82-old-b244-cruel-reality-lose", func(ctx *effects.Context, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	victim := r.Carry[0]
	if len(r.Sacrifices().By(victim)) > 0 {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(r.Source, victim, -5)
})
