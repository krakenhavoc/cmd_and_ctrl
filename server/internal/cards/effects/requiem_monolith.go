package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const requiemMonolithGrant = "requiem-monolith/draw-and-lose"

// Requiem Monolith — Artifact {2}{B}:
//
//	"{T}: Until end of turn, target creature gains "Whenever this
//	 creature is dealt damage, you draw that many cards and lose that
//	 much life." That creature's controller may have this artifact
//	 deal 1 damage to it. Activate only as a sorcery."
//
// A duration grant (ADR 0093 PR 4, Feign Death's shape): the damage
// trigger is a catalog bundle the ability gives the target until end of
// turn, so it is the CREATURE's ability. Its controller controls it and
// "you" is that player, whoever that is, and it fires for every point
// of damage the creature is dealt for the rest of the turn, the
// Monolith's own included. "That many" is the damage the event carries;
// the cards are drawn first and then the life is lost, as printed.
//
// The offer is a real yes/no prompt addressed to the target creature's
// controller, read as the ability resolves (an opponent's creature asks
// the opponent). If they say yes, the Monolith deals 1 damage to the
// creature: the damage is the Monolith's, colourless, and a creature
// that has since gone is simply not damaged. Declining leaves the grant
// in place.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13a8b19b-23bd-4e45-81ff-23371665e4fc",
		Name:         "Requiem Monolith",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: requiemMonolithGrant,
			Triggered: []game.TriggeredAbility{{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return b35SelfWasDealtDamage(ev, source)
				},
				Key:    "Requiem Monolith — draw that many cards and lose that much life",
				Effect: requiemMonolithDrawAndLose,
			}},
			Text: "Whenever this creature is dealt damage, you draw that many cards and lose that much life.",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{T}: Until end of turn, target creature gains \"Whenever this creature is dealt damage, you draw that many cards and lose that much life.\" That creature's controller may have this artifact deal 1 damage to it. Activate only as a sorcery.",
			Cost:         TapCost(),
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature"),
			Effect:       requiemMonolithActivate,
		}},
	})
}

// requiemMonolithDrawAndLose is the granted trigger's body: the
// creature's controller (the trigger's controller) draws as many cards
// as the damage dealt, then loses as much life.
func requiemMonolithDrawAndLose(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	n := item.Trigger.Event.Amount
	if n <= 0 {
		return nil
	}
	if err := (DrawCards{Player: item.Controller, N: n}).Apply(NewContext(g, item)); err != nil {
		return err
	}
	if p := g.PlayerByIDForEffect(item.Controller); p == nil || p.Eliminated {
		return nil
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -n)
}

// requiemMonolithActivate grants the trigger, then offers the 1 damage
// to the creature's controller.
func requiemMonolithActivate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (GrantAbilitiesFor{
		Target: target,
		Keys:   []string{requiemMonolithGrant},
		Label:  "Requiem Monolith — until end of turn, the creature draws and loses when dealt damage",
	}).Apply(ctx); err != nil {
		return err
	}
	creature, ok := g.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	return MayChoice{
		Player:   creature.Controller,
		Question: "Requiem Monolith — have it deal 1 damage to " + creature.Name + "?",
		OnYes:    requiemMonolithPing(target),
	}.Apply(ctx)
}

func requiemMonolithPing(target uuid.UUID) func(ctx *Context) error {
	return func(ctx *Context) error {
		if !onBattlefield(ctx.Game, target) {
			return nil
		}
		// The grant just written is a layer-6 declaration: bring the
		// layers up to date so the damage below already sees the
		// creature's new trigger.
		ctx.Game.RecomputeLayersIfStaleLocked()
		return DealDamage{Source: ctx.Source(), Target: target, Amount: 1}.Apply(ctx)
	}
}
