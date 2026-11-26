package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ogre Geargrabber — Creature — Ogre Warrior {4}{R}{R}, 4/4 (Scars of
// Mirrodin, #5577):
//
//	"Whenever this creature attacks, gain control of target Equipment an
//	 opponent controls until end of turn. Attach it to this creature. When
//	 you lose control of that Equipment, unattach it."
//
// An attack trigger with one target. On resolution the theft is a layer-2
// effect until end of turn (GainControl, CR 611.2b), the Equipment is
// attached to the Ogre when the Ogre is still on the battlefield (CR 701.3b),
// and a delayed trigger (CR 603.7) watches the Equipment's controller change
// away from the thief, whether the duration ended or another effect took it,
// and unattaches it. Stolen Uniform's shape, without its intervening "if".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a483b1fd-751a-447e-8e4f-a54b1c194d2c",
		Name:         "Ogre Geargrabber",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventAttack, effects.ThisAttacked,
					"Ogre Geargrabber — gain control of target Equipment an opponent controls until end of turn, attach it to this creature",
					ogreGeargrabberSteal),
				effects.TargetPermanent("target Equipment an opponent controls", effects.HasSubtype("Equipment"), effects.OpponentControls())),
		},
	})
}

func ogreGeargrabberSteal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.GainControl{
		Target: id, Controller: item.Controller,
		Label: "Ogre Geargrabber — gain control of the Equipment until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	if info, have := ctx.SourcePermanent(); have && !info.Left {
		if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: item.SourceCardID}); err != nil {
			return err
		}
	}
	return effects.DelayedOnEvent{
		Label:          "Ogre Geargrabber — unattach the Equipment",
		On:             []game.EventKind{game.EventControlChanged},
		Condition:      ogreGeargrabberLostControl,
		Cards:          []uuid.UUID{id},
		Body:           ogreGeargrabberUnattachBody,
		ThroughCleanup: true,
	}.Apply(ctx)
}

// ogreGeargrabberLostControl is "when you lose control of that Equipment": it
// changed controller away from the trigger's controller.
var ogreGeargrabberLostControl = game.DelayedCondition("ogre-geargrabber/lost-control-of-the-equipment",
	func(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
		return ev.Kind == game.EventControlChanged && ev.Target == dt.Controller &&
			len(dt.Cards) > 0 && ev.CardID == dt.Cards[0]
	})

var ogreGeargrabberUnattachBody = game.SimpleDelayedBody("ogre-geargrabber/unattach-the-equipment",
	func(g *game.Game, item *game.StackItem) error {
		for _, t := range item.Targets {
			if t.Kind != game.TargetCard || !effects.OnBattlefield(g, t.ID) {
				continue
			}
			if eq, ok := g.LookupCardForEffect(t.ID); !ok || !eq.IsAttached() {
				continue
			}
			if err := g.UnattachForEffect(t.ID); err != nil {
				return err
			}
		}
		return nil
	})
