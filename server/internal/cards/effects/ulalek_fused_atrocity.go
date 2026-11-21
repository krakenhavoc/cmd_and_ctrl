package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ulalek, Fused Atrocity — Legendary Creature — Eldrazi {C/W}{C/U}{C/B}{C/R}{C/G},
// 2/5:
//
//	"Devoid (This card has no color.)
//	 Whenever you cast an Eldrazi spell, you may pay {C}{C}. If you do,
//	 copy all spells you control, then copy all other activated and
//	 triggered abilities you control. You may choose new targets for the
//	 copies. (Mana abilities can't be copied.)"
//
// The trigger fires for an Eldrazi spell its controller casts (the type
// read through the card's text, CR 612.2) and asks for {C}{C}, colourless
// mana only. When it is paid, every spell its controller controls on the
// stack is copied in stack order, the Eldrazi spell just cast included (from
// last-known information if it has already left the stack, CR 608.2h), then
// every other activated and triggered ability it controls, this trigger
// excluded, again in stack order (CR 707.10: the copies are controlled by
// that player and go on the stack above everything). Each copy may have new
// targets (CR 707.10c). An item that has left the stack is skipped, and a
// mana ability never used the stack.
func init() {
	const name = "Ulalek, Fused Atrocity"
	Register(Spec{
		OracleID:        "9f6828e3-39d9-45d9-9bf1-0e3737a0321e",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventCast || ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && spell.HasSubtype(source.TextSubtype("Eldrazi"))
			}, name+" — you may pay {C}{C}; if you do, copy all spells you control, then copy all other activated and triggered abilities you control",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{C}{C}",
						Question: name + " — pay {C}{C} to copy all spells you control, then all other activated and triggered abilities you control?",
						Then:     ulalekPaid,
						Carry:    []uuid.UUID{item.Trigger.Event.CardID},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// ulalekPaid makes the copies once {C}{C} is paid. An on-disk key.
var ulalekPaid = PayUnlessBranches("pay-unless/ulalek-copy-everything", func(ctx *Context, r game.PayUnlessAnswer) error {
	g, you := ctx.Game, ctx.Controller()
	var self uuid.UUID
	if ctx.Item != nil {
		self = ctx.Item.ID
	}
	var spells, abilities []*game.StackItem
	for _, it := range g.StackMeta {
		if it == nil || it.ID == self || it.Controller != you {
			continue
		}
		if it.Kind == game.StackItemSpell {
			spells = append(spells, it)
		} else {
			abilities = append(abilities, it)
		}
	}
	bySeq := func(s []*game.StackItem) {
		sort.Slice(s, func(i, j int) bool { return s[i].Seq < s[j].Seq })
	}
	bySeq(spells)
	bySeq(abilities)
	// The Eldrazi spell that triggered this is a spell its controller
	// controls even when the answer comes after it has left the stack (the
	// keyed question does not hold it there): it is copied from last-known
	// information (CR 608.2h), first, as the topmost spell.
	var ids []uuid.UUID
	cast := uuid.Nil
	if len(r.Carry) > 0 {
		cast = r.Carry[0]
		ids = append(ids, cast)
	}
	for _, it := range spells {
		if it.ID != cast {
			ids = append(ids, it.ID)
		}
	}
	for _, id := range ids {
		if err := (CopySpell{StackID: id, Controller: you, ChooseNewTargets: true, FromLastKnown: true}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, it := range abilities {
		if err := (CopyAbility{ItemID: it.ID, Controller: you, ChooseNewTargets: true}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}, nil)
