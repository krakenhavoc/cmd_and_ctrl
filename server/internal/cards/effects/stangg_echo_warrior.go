package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stangg, Echo Warrior — Legendary Creature — Human Warrior {2}{R}{G}, 3/4
// (Dominaria United Commander):
//
//	"Whenever Stangg attacks, create Stangg Twin, a legendary 3/4 red and
//	 green Human Warrior creature token. It enters tapped and attacking. For
//	 each Aura and Equipment attached to Stangg, create a token that's a copy
//	 of it attached to Stangg Twin. Sacrifice all tokens created this way at
//	 the beginning of the next end step."
//
// One trigger, three steps, each the keyed tail of the one before (so a table
// waiting on a token doubler's order prompt is a restore point). The Twin
// enters tapped and attacking the player Stangg attacked or a planeswalker
// they control (the attack split Myriad's copies use; it was never declared as
// an attacker, CR 508.4, so nothing that watches attacks sees it, Stangg's own
// trigger included). Then, as the Twin is made, each Aura and Equipment
// attached to Stangg right now is copied (CR 707.2: its copiable values, a
// token copy of a token too): an Aura's copy enters attached to the Twin
// (CR 303.4f, as a copy of an Aura spell does) and is not created if it can't
// enchant it (CR 303.4i); an Equipment's copy is attached to the Twin once
// it has entered. Finally the Twin and every copy are sacrificed at the
// beginning of the next end step by one delayed trigger naming them
// (CR 603.7): a token that has left by then is not found again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aaf2a587-ced7-4615-b10f-164e7384b073",
		Name:         "Stangg, Echo Warrior",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, Self,
				"Stangg, Echo Warrior — create Stangg Twin tapped and attacking, a token copy of each Aura and Equipment attached to Stangg attached to it; sacrifice them at the beginning of the next end step",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					defender := b17DefendingPlayer(g, item.Trigger.Event)
					if defender == uuid.Nil {
						return nil
					}
					ctx := NewContext(g, item)
					tmpl := ctx.TextToken(TokenCard("3/4 red and green Legendary Human Warrior named Stangg Twin"))
					return g.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller:         item.Controller,
						Source:             item.SourceCardID,
						Groups:             []game.TokenGroup{{Template: tmpl, Count: 1, Entry: game.TokenEntryOptions{Tapped: true}, AttackOpponent: defender}},
						AttackChoicesLabel: "Stangg, Echo Warrior — Stangg Twin",
					}, ctx.ItemTail(stanggTwinCreated))
				}),
		},
	})
}

// stanggTwinCreated copies each Aura and Equipment attached to Stangg onto the
// Twin. An on-disk key: never renamed, never reused.
var stanggTwinCreated = CardTail("stangg-echo-warrior/twin-created", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	g := ctx.Game
	twin := game.TargetRef{Kind: game.TargetCard, ID: o.IDs[0]}
	var groups []game.TokenGroup
	for _, id := range g.AttachmentsOf(ctx.Source()) {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !(c.HasSubtype("Aura") || c.HasSubtype("Equipment")) {
			continue
		}
		tmpl, ok := TokenCopyTemplate(g, id)
		if !ok {
			continue
		}
		grp := game.TokenGroup{Template: copiedToken(tmpl), Count: 1}
		if c.HasSubtype("Aura") {
			grp.Entry = game.TokenEntryOptions{AttachedTo: twin}
		}
		groups = append(groups, grp)
	}
	if len(groups) == 0 {
		return stanggSacrificeAtEnd(ctx, o.IDs)
	}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: ctx.Controller(), Source: ctx.Source(), Groups: groups, Together: true,
	}, withCarry(ctx.ItemTail(stanggCopiesCreated), o.IDs...))
})

// stanggCopiesCreated attaches each Equipment copy to the Twin and schedules the
// sacrifice of the Twin and every copy. The carried IDs are the Twin(s). An
// on-disk key.
var stanggCopiesCreated = CardTail("stangg-echo-warrior/copies-created", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	g := ctx.Game
	twin := game.TargetRef{Kind: game.TargetCard, ID: o.Tail.Carry[0]}
	for _, id := range o.IDs {
		if c, ok := g.LookupCardForEffect(id); ok && c.HasSubtype("Equipment") {
			if err := g.AttachForEffect(id, twin); err != nil {
				return err
			}
		}
	}
	all := append(append([]uuid.UUID(nil), o.Tail.Carry...), o.IDs...)
	return stanggSacrificeAtEnd(ctx, all)
})

// stanggSacrificeAtEnd is "Sacrifice all tokens created this way at the beginning
// of the next end step."
func stanggSacrificeAtEnd(ctx *Context, tokens []uuid.UUID) error {
	return ScheduleDelayedTrigger{
		Label: "Stangg, Echo Warrior — sacrifice the tokens",
		Cards: tokens,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
}
