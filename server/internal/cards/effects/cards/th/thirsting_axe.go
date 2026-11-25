package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thirsting Axe — Artifact — Equipment {3} (Eldritch Moon, S82 OLD, #5541):
//
//	"Equipped creature gets +4/+0.
//	 At the beginning of your end step, if equipped creature didn't deal
//	 combat damage to a creature this turn, sacrifice it.
//	 Equip {2}"
//
// An intervening "if" (CR 603.4): the Axe must be attached as the end step
// begins and again as the ability resolves, and the creature it names is the
// one equipped then. "Didn't deal combat damage to a creature this turn" is
// read off the turn's damage events, for the creature it is attached to now
// (Giant's Skewer's reading of "equipped creature deals combat damage to a
// creature"): damage to a player or planeswalker does not count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e024870c-af13-4ef3-be27-cdd84dab7d7d",
		Name:         "Thirsting Axe",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(4, 0)},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{Name: "equipped creature didn't deal combat damage to a creature this turn", Test: func(ctx *effects.Context) bool {
				axe, ok := ctx.Game.LookupCardForEffect(ctx.Source())
				return ok && equippedCreatureIdle(ctx.Game, &axe)
			}}, effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && equippedCreatureIdle(g, source)
			}, "Thirsting Axe — sacrifice equipped creature", sacrificeEquippedCreature)),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// equippedCreatureIdle: a creature is attached to `axe` and it has dealt no
// combat damage to a creature this turn.
func equippedCreatureIdle(g *game.Game, axe *game.Card) bool {
	host := g.AttachedHostOf(axe)
	if host == nil {
		return false
	}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventDealDamage && ev.Combat && ev.Amount > 0 && ev.Source == host.InstanceID {
			if c, ok := g.LookupCardForEffect(ev.Target); ok && c.IsCreature() {
				return false
			}
		}
	}
	return true
}

// sacrificeEquippedCreature sacrifices the creature the source is attached to.
func sacrificeEquippedCreature(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	axe, ok := g.LookupCardForEffect(ctx.Source())
	if !ok {
		return nil
	}
	host := g.AttachedHostOf(&axe)
	if host == nil || !equippedCreatureIdle(g, &axe) { // CR 603.4: checked again on resolution
		return nil
	}
	return effects.SacrificePermanent{Target: host.InstanceID}.Apply(ctx)
}
