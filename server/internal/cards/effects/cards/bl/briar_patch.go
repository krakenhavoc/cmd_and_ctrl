package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Briar Patch — Enchantment {1}{G}{G} (Mercadian Masques, S82 OLD, #5619):
//
//	"Whenever a creature attacks you, it gets -1/-0 until end of turn."
//
// "Attacks you" is an attack that names the enchantment's controller as a player,
// not one of their planeswalkers or battles. The shrink is the attacker's, and
// only while it is still the permanent that attacked (CR 400.7); a creature that
// has left is not affected.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "247793af-be9e-47c7-8887-e99e49478b0b",
		Name:         "Briar Patch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, briarPatchAttacksYou, "Briar Patch — it gets -1/-0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
						return nil
					}
					return effects.BoostUntilEOT{Target: ctx.Trigger().Event.CardID, Power: -1, Label: "Briar Patch — -1/-0 until end of turn"}.Apply(ctx)
				}),
		},
	})
}

// briarPatchAttacksYou is "a creature attacks you": the attack names the
// enchantment's controller.
func briarPatchAttacksYou(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventAttack && ev.CardID != uuid.Nil && ev.Target != uuid.Nil && ev.Target == source.Controller
}
