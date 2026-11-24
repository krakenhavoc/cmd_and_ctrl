package cardsh

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hans Eriksson — Legendary Creature — Human Scout {2}{R}{G}, 1/4 (Commander
// Legends):
//
//	"Whenever Hans Eriksson attacks, reveal the top card of your library. If
//	 it's a creature card, put it onto the battlefield tapped and attacking
//	 defending player or a planeswalker they control. Otherwise, put that card
//	 into your hand. When you put a creature card onto the battlefield this
//	 way, it fights Hans Eriksson."
//
// The trigger reveals the top card (an empty library reveals nothing). A
// non-creature card goes to the hand. A creature card is put onto the
// battlefield tapped and attacking the defending player or one of their
// planeswalkers, as you choose (CR 508.4; one prompt when there is more than
// one, the player Hans attacked, or the controller of the planeswalker he
// attacked, first); it was never declared as an attacker, so it triggers
// nothing as one. Only if it did enter, the reflexive trigger (CR 603.12) has
// it fight Hans, as each last existed if either has gone (CR 701.14b: no
// fight unless both are still on the battlefield).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a37ac5c-341b-48ab-9b0b-5ef8bf097aab",
		Name:         "Hans Eriksson",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.Self, "Hans Eriksson — reveal the top card of your library", hansErikssonReveal),
		},
	})
}

var hansErikssonEntered = effects.EnteredTail("hans-eriksson/entered", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	return effects.ReflexiveTrigger{
		Label: "Hans Eriksson — the creature fights Hans Eriksson",
		Cards: []uuid.UUID{entered},
		Body:  hansErikssonFightBody,
	}.Apply(ctx)
})

var hansErikssonFightBody = game.ReflexiveBody("s82-old-b087/hans-eriksson-fight", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	returned := ctx.PayloadCards()
	if len(returned) == 0 {
		return nil
	}
	return effects.B10Fight(ctx, returned[0], item.SourceCardID)
}), nil)

// hansErikssonDefenders is what the creature may enter attacking: the
// defending player (the player Hans attacked, or the controller of the
// planeswalker he attacked), then the planeswalkers that player controls.
func hansErikssonDefenders(g *game.Game, item *game.StackItem) []uuid.UUID {
	attacked := effects.TriggerAttackTarget(item)
	defender := attacked
	if g.PlayerByIDForEffect(attacked) == nil {
		if c, ok := g.LookupCardForEffect(attacked); ok {
			defender = c.Controller
		}
	}
	var out []uuid.UUID
	for _, ref := range g.AttackTargetsForEffect(item.Controller) {
		if ref.ID == defender {
			out = append([]uuid.UUID{ref.ID}, out...)
			continue
		}
		if c, ok := g.LookupCardForEffect(ref.ID); ok && c.IsPlaneswalker() && c.Controller == defender {
			out = append(out, ref.ID)
		}
	}
	if i := slices.Index(out, attacked); i > 0 {
		out = append([]uuid.UUID{attacked}, slices.Delete(out, i, i+1)...)
	}
	return out
}

func hansErikssonReveal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var revealed []uuid.UUID
	if err := (effects.RevealTopOfLibrary{
		Player:   item.Controller,
		N:        1,
		Reason:   "Hans Eriksson — reveal the top card of your library",
		Revealed: &revealed,
	}).Apply(ctx); err != nil || len(revealed) == 0 {
		return err
	}
	card, ok := g.LookupCardForEffect(revealed[0])
	if !ok {
		return nil
	}
	if !card.IsCreature() {
		return effects.BounceToHand{Target: revealed[0]}.Apply(ctx)
	}
	targets := hansErikssonDefenders(g, item)
	switch len(targets) {
	case 0:
		return hansErikssonPut(g, item, revealed[0], uuid.Nil)
	case 1:
		return hansErikssonPut(g, item, revealed[0], targets[0])
	}
	options := make([]game.ChoiceOption, len(targets))
	for i, id := range targets {
		options[i] = game.ChoiceOption{Label: "Attack " + effects.AttackTargetName(g, id)}
	}
	return effects.PickOption{
		Question: "Hans Eriksson — what does it attack?",
		Options:  options,
		ThenKey:  hansErikssonAttacks,
		WithItem: true,
		Carry:    append([]uuid.UUID{revealed[0]}, targets...),
	}.Apply(ctx)
}

var hansErikssonAttacks = effects.OptionPickThen("option-pick/hans-eriksson-attacks", func(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) < 2 {
		return nil
	}
	card, targets := r.Carry[0], r.Carry[1:]
	index := r.Index
	if index < 0 || index >= len(targets) {
		index = 0
	}
	return hansErikssonPut(ctx.Game, ctx.Item, card, targets[index])
})

// hansErikssonPut puts the revealed creature card onto the battlefield tapped
// and attacking `attacking` (uuid.Nil: not attacking), if it is still on top
// of the library, then runs the reflexive fight.
func hansErikssonPut(g *game.Game, item *game.StackItem, card, attacking uuid.UUID) error {
	if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneLibrary {
		return nil
	}
	return g.PutCardsFromLibraryOntoBattlefieldKeyedForEffect([]uuid.UUID{card}, game.LibraryEntryOptions{
		Controller: item.Controller,
		Tapped:     true,
		Attacking:  attacking,
	}, effects.NewContext(g, item).ItemTail(hansErikssonEntered))
}
