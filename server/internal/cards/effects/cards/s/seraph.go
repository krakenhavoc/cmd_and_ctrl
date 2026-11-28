package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seraph — Creature — Angel {6}{W}, 4/4 (Ice Age, S82 OLD batch 846, #5636):
//
//	"Flying
//	 Whenever a creature dealt damage by this creature this turn dies, put that
//	 card onto the battlefield under your control at the beginning of the next
//	 end step. Sacrifice the creature when you lose control of this creature."
//
// The first trigger is Sengir Vampire's (the turn's damage record, CR 120.1). It
// schedules a delayed return of that card at the next end step
// (Loyal Cathar's shape): it comes back under the Seraph's controller as the
// ability was put on the stack, only if it is still in the graveyard as the card
// that died (CR 400.7). The Seraph notes each creature it brought back, so a
// second death adds a second creature.
//
// "Lose control of this creature" is Coffin Queen's trigger: the Seraph changing
// controllers, or leaving the battlefield, which ends its controller's control
// of it. It sacrifices every creature noted, each only while it is still the
// object that returned. A Seraph that has already gone when the creature comes
// back never noted it, so nothing is left to trigger: the creature stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a1c50bd1-120d-44ac-af20-13a90fc991e6",
		Name:            "Seraph",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && effects.DealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Seraph — put that card onto the battlefield under your control at the beginning of the next end step",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					obj := ctx.Trigger().Object
					if obj == nil || !effects.DiedCardStillInGraveyard(ctx, obj.Ref().ID) {
						return nil
					}
					return effects.ScheduleDelayedTrigger{
						At:     game.StepEnd,
						Label:  "Seraph — put that card onto the battlefield under your control",
						Cards:  []uuid.UUID{obj.Ref().ID},
						Body:   effects.OldB846SeraphReturnBody,
						Params: game.EffectParams{Object: obj.Ref()},
					}.Apply(ctx)
				}),
			effects.OnAny([]game.EventKind{game.EventControlChanged, game.EventLTB},
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventControlChanged {
						return effects.ThisChangedController(ev, source, lki, g)
					}
					return effects.Self(ev, source, lki, g)
				}, "Seraph — sacrifice the creatures it brought back", effects.OldB846SeraphSacrificeTheCreatures),
		},
	})
}
