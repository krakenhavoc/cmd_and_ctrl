package effects

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b635_helpers.go — Apocalypse, batch 635 (#5614).

// S82OldB635JadedResponse is "Counter target spell if it shares a color with a
// creature you control." The share is judged as it resolves (CR 608.2b): the
// spell's colours as it sits on the stack against the creatures you control
// then. A colourless spell shares none.
func S82OldB635JadedResponse(_ *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	spell, found := ctx.Game.LookupCardForEffect(t.ID)
	if !found {
		return nil
	}
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() && SharesAColorNow(spell, c) {
			return CounterTarget{StackID: t.ID}.Apply(ctx)
		}
	}
	return nil
}

// S82OldB635EvasiveAction is "Domain — Counter target spell unless its
// controller pays {1} for each basic land type among lands you control." The
// count is read as it resolves; with no basic land type the payment is {0}
// and the spell is not countered.
func S82OldB635EvasiveAction(_ *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	n := DomainCount(ctx.Game, ctx.Controller())
	if n <= 0 {
		return nil
	}
	cost := "{" + strconv.Itoa(n) + "}"
	return CounterUnlessPaid{
		StackID:  t.ID,
		Cost:     cost,
		Question: "Evasive Action — pay " + cost + " or your spell is countered",
	}.Apply(ctx)
}

// S82OldB635HauntedAngel is "When this creature dies, exile it and each other
// player creates a 3/3 black Angel creature token with flying." The two
// instructions do not depend on each other: every player but its controller
// makes a token, and the Angel is exiled if it is still in a graveyard.
func S82OldB635HauntedAngel(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == ctx.Controller() {
			continue
		}
		if err := (CreateToken{Controller: p.ID, Template: TokenCard("3/3 black Angel with flying"), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if z := g.FindCardZoneForEffect(item.SourceCardID); z != nil && z.Kind == game.ZoneGraveyard {
		return ExileTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
	}
	return nil
}

// S82OldB635TahngarthsGlare is "Look at the top three cards of target
// opponent's library, then put them back in any order. That player looks at
// the top three cards of your library, then puts them back in any order." Two
// separate placements by two choosers, each over its own library.
func S82OldB635TahngarthsGlare(_ *game.StackItem, ctx *Context) error {
	opp, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	if err := (LookAtLibraryThenPlace{
		Owner: opp, N: 3, Placement: game.LibraryPlaceTop,
		Label: "Tahngarth's Glare — put them back in any order",
	}).Apply(ctx); err != nil {
		return err
	}
	return LookAtLibraryThenPlace{
		Looker: opp, Owner: ctx.Controller(), N: 3, Placement: game.LibraryPlaceTop,
		Label: "Tahngarth's Glare — that player puts your top three cards back in any order",
	}.Apply(ctx)
}

// S82OldB635DestroyTarget is "Destroy target creature …" as an activated
// ability's whole effect: the first target still legal as it resolves.
func S82OldB635DestroyTarget(g *game.Game, item *game.StackItem) error {
	return DestroyTheTargetPermanent(item, NewContext(g, item))
}

// S82OldB635IceCaveCost is the mana cost of the spell as it sits on the
// stack, X read as announced (CR 107.3b); "" when it has none to pay.
func S82OldB635IceCaveCost(g *game.Game, spellID uuid.UUID) string {
	spell, ok := g.LookupCardForEffect(spellID)
	if !ok || strings.TrimSpace(spell.ManaCost) == "" {
		return ""
	}
	x := 0
	if it := g.StackItemForEffect(spellID); it != nil {
		x = it.XValue
	}
	return strings.ReplaceAll(spell.ManaCost, "{X}", "{"+strconv.Itoa(x)+"}")
}

// S82OldB635IceCaveCounter is "If a player does, counter the spell": the spell
// the offer guarded, if it is still on the stack (CR 608.2b). An on-disk key:
// never renamed, never reused.
var S82OldB635IceCaveCounter = AnyPlayerConsequences("s82-old-b635/ice-cave-counter",
	func(ctx *Context, r game.AnyPlayerAnswer) error {
		if r.Guards == uuid.Nil {
			return nil
		}
		if it := ctx.Game.StackItemForEffect(r.Guards); it == nil || it.Kind != game.StackItemSpell {
			return nil
		}
		return CounterTarget{StackID: r.Guards}.Apply(ctx)
	}, nil, nil)

// S82OldB635IceCave is "Whenever a player casts a spell, any other player may
// pay that spell's mana cost. If a player does, counter the spell." The
// offer is made to every player but the caster (the offer's controller is the
// caster, so an Opponents offer leaves them out), in APNAP order, one prompt at
// a time, guarding the spell while it is open.
func S82OldB635IceCave(label string) game.TriggeredAbility {
	return On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return ev.CardID != uuid.Nil && ev.Actor != uuid.Nil
	}, label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		ev := ctx.Trigger().Event
		if g.StackItemForEffect(ev.CardID) == nil {
			return nil
		}
		cost := S82OldB635IceCaveCost(g, ev.CardID)
		if cost == "" {
			return nil
		}
		return g.QueueAnyPlayerPaysKeyedForEffect(game.AnyPlayerPaysKeyed{
			Source:     ctx.Source(),
			Controller: ev.Actor,
			Cost:       cost,
			Question:   "Ice Cave — pay " + cost + " to counter the spell?",
			Opponents:  true,
			Guards:     ev.CardID,
			Then:       S82OldB635IceCaveCounter,
			Asked:      game.NewAskedItem(ctx.Item),
		})
	})
}
