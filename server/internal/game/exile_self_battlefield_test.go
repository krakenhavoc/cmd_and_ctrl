package game

import (
	"testing"

	"github.com/google/uuid"
)

// exile_self_battlefield_test.go — #1404 / ADR 0020's 2026-09-24
// amendment: AbilityCost.ExileSelf paid from the BATTLEFIELD ("Exile
// this artifact:", Perpetual Timepiece). The graveyard leg (#1221) is
// in other_zone_ability_test.go and is unchanged; the catalog half is
// cards/effects/exile_this_permanent_test.go and the enumerator's is
// legal/exile_self_battlefield_test.go.

// exileThisPermanentAbility is "{cost}, Exile this artifact: draw a
// card" — no Zones, so it functions from the battlefield.
func exileThisPermanentAbility(cost string, tap bool) ActivatedAbilityShape {
	return ActivatedAbilityShape{
		Label: "Exile-this-permanent test " + cost,
		Cost:  AbilityCost{Mana: cost, Tap: tap, ExileSelf: true},
		Effect: func(g *Game, item *StackItem) error {
			return g.DrawNForEffect(item.Controller, 1)
		},
	}
}

// pushExileThisPermanent puts an artifact carrying the ability onto
// the battlefield under p's control.
func pushExileThisPermanent(g *Game, p *Player, ab ActivatedAbilityShape) uuid.UUID {
	c := NewCard("Timepiece Test", p.ID)
	c.TypeLine = "Artifact"
	c.Controller = p.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{ab}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// The battlefield leg: the permanent is in exile when the activation
// returns — before anyone can respond, so before the ability resolves
// — and it left as a permanent LEAVING THE BATTLEFIELD: one EventLTB
// naming exile, no EventSacrifice, and nothing in a graveyard, which
// is what keeps every dies trigger off it.
func TestExileSelfFromTheBattlefieldExilesAtAnnounce(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushExileThisPermanent(g, me, exileThisPermanentAbility("{1}", false))
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	before := len(g.Events)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(src) {
		t.Fatal("the source is still on the battlefield after paying its exile cost")
	}
	if !g.Exile.Contains(src) {
		t.Fatal("the source is not in exile")
	}
	if me.Graveyard.Contains(src) {
		t.Error("the source went to the graveyard — exile-as-cost is not a death")
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("%d stack items, want the ability", len(g.StackMeta))
	}
	ltb := 0
	for _, ev := range g.Events[before:] {
		switch {
		case ev.Kind == EventLTB && ev.CardID == src:
			ltb++
			if ev.NewZone != ZoneExile {
				t.Errorf("EventLTB names %s, want exile — a dies watcher reads this field", ev.NewZone)
			}
		case ev.Kind == EventSacrifice && ev.CardID == src:
			t.Error("EventSacrifice fired — exiling a permanent as a cost is not sacrificing it")
		}
	}
	if ltb != 1 {
		t.Errorf("%d EventLTB for the source, want exactly one", ltb)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want the {1} spent", me.ManaPool)
	}
}

// A refused activation pays NOTHING (ADR 0020 §3): a "{2}, {T}, Exile
// this artifact" whose mana cannot be paid leaves the source on the
// battlefield, untapped, and the pool as it was. The exile is the last
// component paid, so no earlier failure can strand the card in exile.
func TestExileSelfFromTheBattlefieldRefusedActivationPaysNothing(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushExileThisPermanent(g, me, exileThisPermanentAbility("{2}", true))
	me.ManaPool.AddMana(ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("an unpayable {2} was accepted")
	}
	if !g.Battlefield.Contains(src) || g.Exile.Contains(src) {
		t.Fatal("a refused activation moved the source")
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == src && c.Tapped {
			t.Error("a refused activation tapped the source")
		}
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool = %v, want the mana untouched", me.ManaPool)
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("%d stack items after a refusal, want none", len(g.StackMeta))
	}
}

// CR 903.9 on the battlefield leg, through #1397's ask-first gate
// (cost_commander_choice.go): a commander exiled to its own ability is
// asked BEFORE anything is paid — the announcement parks with the
// commander still on the battlefield, the mana unspent and nothing on
// the stack — and the owner's answer makes the announcement, which then
// pays in one indivisible step with the answer on the move. Both
// answers are checked: the prompt is only worth offering if the card
// lands where the answer says.
func TestExileSelfFromTheBattlefieldAsksACommandersOwnerFirst(t *testing.T) {
	for _, takeCommandZone := range []bool{true, false} {
		g := newActiveGame(t)
		advanceTo(t, g, StepPrecombatMain)
		me := g.Seats[0]
		id := seatCommander(t, g.Battlefield, me)
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].ActivatedAbilities = []ActivatedAbilityShape{exileThisPermanentAbility("{1}", false)}
			}
		}
		me.ManaPool.AddMana(ManaToken{Color: "C"})

		if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		prompt := expectCommanderPrompt(t, g, me)
		if !g.Battlefield.Contains(id) || len(me.ManaPool) != 1 || len(g.StackMeta) != 0 {
			t.Fatalf("something was paid before the owner answered: on battlefield %v, pool %v, %d stack items",
				g.Battlefield.Contains(id), me.ManaPool, len(g.StackMeta))
		}

		if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, takeCommandZone); err != nil {
			t.Fatalf("ResolveOptionalReplacement: %v", err)
		}
		if takeCommandZone {
			assertOnlyIn(t, id, me.Command, g.Battlefield, g.Exile)
		} else {
			assertOnlyIn(t, id, g.Exile, g.Battlefield, me.Command)
		}
		if len(g.PendingChoices) != 0 {
			t.Errorf("%d pending choices after the answer — the payment must not pause again", len(g.PendingChoices))
		}
		if len(g.StackMeta) != 1 {
			t.Errorf("%d stack items after the answer, want the ability announced", len(g.StackMeta))
		}
		if len(me.ManaPool) != 0 {
			t.Errorf("pool = %v after the answer, want the {1} spent", me.ManaPool)
		}
	}
}

// The auto-tapper must not spend a source the cost is about to exile
// (#1242's rule for a sacrifice, CR 118.3): a planner that cracked it
// for the mana half would leave the exile with nothing to pay.
func TestAbilityAutoTapExclusionsIncludeAnExiledSource(t *testing.T) {
	src := uuid.New()
	got := AbilityAutoTapExclusions(src, AbilityCost{Mana: "{2}", ExileSelf: true}, nil, nil, nil, nil)
	if !got[src] {
		t.Errorf("exclusions = %v, want the source excluded — the cost exiles it", got)
	}
}
