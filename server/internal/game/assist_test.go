package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// assist_test.go — ADR 0150 §4 (#1857): assist (CR 702.132a). The cast
// is parked on the helper's pay_amount before anything is paid, the cap
// is the generic part of the TOTAL cost, and the re-run pays both
// players or neither (CR 601.2h).

// assistTable is a three-seat game at the caster's main phase: the
// caster (seat 0, active), the helper (seat 1) and a bystander.
func assistTable(t *testing.T) (*Game, *Player, *Player) {
	t.Helper()
	g := newActiveGameWithSeats(t, 3)
	toMainPhase(t, g)
	return g, g.Seats[0], g.Seats[1]
}

// assistSpell seeds a creature spell with assist in `p`'s hand.
func assistSpell(p *Player, cost string) uuid.UUID {
	c := NewCard("Fan Thing", p.ID)
	c.TypeLine = "Creature — Human Rogue"
	c.ManaCost = cost
	c.Keywords = []string{KeywordAssist}
	p.Hand.PushTop(c)
	return c.InstanceID
}

// lands puts n basic lands of one type under `p`'s control.
func assistLands(g *Game, p *Player, sub string, n int) []uuid.UUID {
	var ids []uuid.UUID
	for range n {
		ids = append(ids, pushBattlefieldForTest(g, p.ID, sub, "Basic Land — "+sub, ""))
	}
	return ids
}

func assistTapped(g *Game, ids []uuid.UUID) int {
	n := 0
	for _, id := range ids {
		if c := findBattlefieldCard(g, id); c != nil && c.Tapped {
			n++
		}
	}
	return n
}

func assistInHand(p *Player, id uuid.UUID) bool {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// assistPrompt is the open assist prompt, or fails.
func assistPrompt(t *testing.T, g *Game) *PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c.Assist != nil {
			return c
		}
	}
	t.Fatalf("no assist prompt open; pending = %d", len(g.PendingChoices))
	return nil
}

func castAssisted(g *Game, caster *Player, id uuid.UUID, helper uuid.UUID) error {
	return g.CastSpell(caster.ID, id, CastSpellParams{Strict: true, AutoTap: true, AssistPlayer: helper})
}

// The park moves and pays nothing, and asks the helper 0..min(G, what
// they can make) of mana, unit assist.
func TestAssistParksTheCastBeforeAnythingIsPaid(t *testing.T) {
	g, me, helper := assistTable(t)
	mine := assistLands(g, me, "Swamp", 1)
	theirs := assistLands(g, helper, "Plains", 2)
	id := assistSpell(me, "{3}{B}")

	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !assistInHand(me, id) || len(g.Stack.Cards) != 0 {
		t.Fatal("the spell moved before the helper answered")
	}
	if assistTapped(g, mine)+assistTapped(g, theirs) != 0 {
		t.Fatal("a land was tapped before the helper answered")
	}
	c := assistPrompt(t, g)
	if c.Kind != PendingChoicePayAmount || c.Chooser != helper.ID || c.Source != id {
		t.Fatalf("prompt = %s to %v about %v", c.Kind, c.Chooser, c.Source)
	}
	pa := c.PayAmount
	if pa.Min != 0 || pa.Max != 2 || pa.Resource != PayResourceMana || pa.Unit != PayAmountAssist {
		t.Errorf("pay_amount = %+v, want 0..2 mana (the helper's two Plains), unit assist", pa)
	}
	if a := c.Assist; a.Caster != me.ID || a.Spell != id || a.SpellName != "Fan Thing" || a.Generic != 3 || a.ManaCost != "{3}{B}" {
		t.Errorf("assist = %+v", a)
	}
}

// The helper pays their share and the caster the rest, both on the
// cast, and the record says who helped.
func TestAssistHelperAndCasterBothPay(t *testing.T) {
	g, me, helper := assistTable(t)
	mine := assistLands(g, me, "Swamp", 2)
	theirs := assistLands(g, helper, "Plains", 3)
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatal(err)
	}
	c := assistPrompt(t, g)
	if err := g.ResolvePayAmount(c.ID, helper.ID, 3); err != nil {
		t.Fatalf("answer: %v", err)
	}
	item := g.StackMeta[id]
	if item == nil || assistInHand(me, id) {
		t.Fatal("the spell is not on the stack")
	}
	if got := assistTapped(g, theirs); got != 3 {
		t.Errorf("helper tapped %d Plains, want 3", got)
	}
	if got := assistTapped(g, mine); got != 1 {
		t.Errorf("caster tapped %d Swamps, want 1 (the {B})", got)
	}
	if item.Paid.AssistedBy != helper.ID || item.Paid.AssistPaid != 3 || len(item.Paid.Mana) != 4 {
		t.Errorf("paid = by %v, %d, %d tokens; want the helper, 3, 4 tokens", item.Paid.AssistedBy, item.Paid.AssistPaid, len(item.Paid.Mana))
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts left open", len(g.PendingChoices))
	}
}

// A partial share: the caster pays what is left of the generic.
func TestAssistPartialShare(t *testing.T) {
	g, me, helper := assistTable(t)
	mine := assistLands(g, me, "Swamp", 4)
	assistLands(g, helper, "Plains", 3)
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.ResolvePayAmount(assistPrompt(t, g).ID, helper.ID, 1); err != nil {
		t.Fatal(err)
	}
	if g.StackMeta[id] == nil {
		t.Fatal("not cast")
	}
	if got := assistTapped(g, mine); got != 3 {
		t.Errorf("caster tapped %d, want 3 ({2}{B})", got)
	}
}

// The helper's 0 is an unassisted cast.
func TestAssistHelperDeclines(t *testing.T) {
	g, me, helper := assistTable(t)
	mine := assistLands(g, me, "Swamp", 4)
	theirs := assistLands(g, helper, "Plains", 3)
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.ResolvePayAmount(assistPrompt(t, g).ID, helper.ID, 0); err != nil {
		t.Fatal(err)
	}
	item := g.StackMeta[id]
	if item == nil || item.Paid.AssistedBy != uuid.Nil || item.Paid.AssistPaid != 0 {
		t.Fatalf("item = %+v", item)
	}
	if assistTapped(g, mine) != 4 || assistTapped(g, theirs) != 0 {
		t.Errorf("tapped caster %d helper %d, want 4 and 0", assistTapped(g, mine), assistTapped(g, theirs))
	}
}

// An answer over the cap, or one the helper can no longer fund, is
// refused with the prompt kept.
func TestAssistAnswerOverTheCapOrUnfundedIsRefused(t *testing.T) {
	g, me, helper := assistTable(t)
	assistLands(g, me, "Swamp", 4)
	theirs := assistLands(g, helper, "Plains", 2)
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatal(err)
	}
	c := assistPrompt(t, g)
	if err := g.ResolvePayAmount(c.ID, helper.ID, 3); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("over the cap: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 1); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("the caster answering: %v, want ErrNotTheChooser", err)
	}
	g.WithWriteLock(func() { findBattlefieldCard(g, theirs[0]).Tapped = true })
	if err := g.ResolvePayAmount(c.ID, helper.ID, 2); !errors.Is(err, ErrInsufficientMana) {
		t.Errorf("unfunded: %v, want ErrInsufficientMana", err)
	}
	if assistPrompt(t, g).ID != c.ID || !assistInHand(me, id) {
		t.Fatal("a refused answer closed the prompt or cast the spell")
	}
	if err := g.ResolvePayAmount(c.ID, helper.ID, 1); err != nil {
		t.Fatalf("a fundable answer after the refusals: %v", err)
	}
	if g.StackMeta[id] == nil {
		t.Error("not cast")
	}
}

// CR 601.2h: the helper can pay but the caster can't, so neither does.
func TestAssistIsAllOrNothing(t *testing.T) {
	g, me, helper := assistTable(t)
	mine := assistLands(g, me, "Plains", 1) // no {B}
	theirs := assistLands(g, helper, "Plains", 3)
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatal(err)
	}
	before := len(g.Events)
	if err := g.ResolvePayAmount(assistPrompt(t, g).ID, helper.ID, 3); err != nil {
		t.Fatalf("the helper's answer is fine, whatever the caster can pay: %v", err)
	}
	if !assistInHand(me, id) || g.StackMeta[id] != nil {
		t.Fatal("cast with the caster short")
	}
	if assistTapped(g, mine)+assistTapped(g, theirs) != 0 {
		t.Fatalf("tapped caster %d helper %d, want nothing", assistTapped(g, mine), assistTapped(g, theirs))
	}
	if len(helper.ManaPool) != 0 || len(me.ManaPool) != 0 {
		t.Error("mana left floating")
	}
	logged := false
	for _, e := range g.Events[before:] {
		if e.Kind == EventEffectError && e.Actor == me.ID && e.Source == id {
			logged = true
		}
	}
	if !logged {
		t.Error("the caster got no log line for the failed cast")
	}
	if len(g.PendingChoices) != 0 {
		t.Error("the prompt survived the answer")
	}
}

// The cap is the TOTAL cost's generic: commander tax raises it.
func TestAssistCapIncludesCommanderTax(t *testing.T) {
	g, me, helper := assistTable(t)
	assistLands(g, helper, "Plains", 9)
	c := NewCard("Fan Commander", me.ID)
	c.TypeLine = "Legendary Creature — Human Rogue"
	c.ManaCost = "{3}{B}"
	c.Keywords = []string{KeywordAssist}
	c.IsCommander = true
	me.Command.PushTop(c)
	me.CommanderCasts[c.InstanceID] = 1
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{FromZone: "command", Strict: true, AutoTap: true, AssistPlayer: helper.ID}); err != nil {
		t.Fatal(err)
	}
	p := assistPrompt(t, g)
	if p.Assist.Generic != 5 || p.PayAmount.Max != 5 {
		t.Errorf("generic %d max %d, want 5 and 5 ({3} + {2} tax)", p.Assist.Generic, p.PayAmount.Max)
	}
}

// ... and a reduction lowers it.
func TestAssistCapIsAfterReductions(t *testing.T) {
	const medallion = "test-1857-medallion"
	g, me, helper := assistTable(t)
	assistLands(g, helper, "Plains", 9)
	withCatalogCostModifiers(t, modifiersFor(medallion, CostModifier{
		Kind: CostReduction, Label: "Spells you cast cost {1} less to cast.", Amount: fixed(1),
	}))
	modifierSource(t, g, me, "Test Medallion", medallion)
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatal(err)
	}
	if got := assistPrompt(t, g).Assist.Generic; got != 2 {
		t.Errorf("generic %d, want 2", got)
	}
}

// What an announcement may not ask.
func TestAssistRefusals(t *testing.T) {
	g, me, helper := assistTable(t)
	assistLands(g, helper, "Plains", 3)
	plain := pushTypedCardToHandWithCost(me, "Plain Bear", "Creature — Bear", "{1}{G}")
	if err := castAssisted(g, me, plain, helper.ID); !errors.Is(err, ErrAssistNotOffered) {
		t.Errorf("no assist: %v", err)
	}
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, me.ID); !errors.Is(err, ErrAssistPlayer) {
		t.Errorf("the caster: %v", err)
	}
	if err := castAssisted(g, me, id, uuid.New()); !errors.Is(err, ErrAssistPlayer) {
		t.Errorf("a stranger: %v", err)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, ForceCast: true, AssistPlayer: helper.ID}); !errors.Is(err, ErrAssistNotOffered) {
		t.Errorf("a forced cast: %v", err)
	}
	gone := g.Seats[2]
	if err := g.Concede(gone.ID); err != nil {
		t.Fatal(err)
	}
	if err := castAssisted(g, me, id, gone.ID); !errors.Is(err, ErrAssistPlayer) {
		t.Errorf("a player who left: %v", err)
	}
	pips := assistSpell(me, "{B}{B}")
	if err := castAssisted(g, me, pips, helper.ID); !errors.Is(err, ErrAssistNoGeneric) {
		t.Errorf("no generic: %v", err)
	}
	if len(g.PendingChoices) != 0 || !assistInHand(me, id) {
		t.Error("a refused announcement parked or moved something")
	}
}

// A helper who can make nothing isn't asked.
func TestAssistHelperWithNothingIsNotAsked(t *testing.T) {
	g, me, helper := assistTable(t)
	assistLands(g, me, "Swamp", 4)
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingChoices) != 0 || g.StackMeta[id] == nil {
		t.Fatalf("pending %d, cast %v", len(g.PendingChoices), g.StackMeta[id] != nil)
	}
}

// A helper's floating mana is theirs to help with: the pool before the
// auto-tapper.
func TestAssistHelperPaysFromTheirPoolFirst(t *testing.T) {
	g, me, helper := assistTable(t)
	assistLands(g, me, "Swamp", 2)
	theirs := assistLands(g, helper, "Plains", 1)
	helper.ManaPool.AddMana(ManaToken{Color: "R"}, ManaToken{Color: "R"})
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatal(err)
	}
	c := assistPrompt(t, g)
	if c.PayAmount.Max != 3 {
		t.Fatalf("max %d, want 3 (two floating, one Plains)", c.PayAmount.Max)
	}
	if err := g.ResolvePayAmount(c.ID, helper.ID, 2); err != nil {
		t.Fatal(err)
	}
	if g.StackMeta[id] == nil {
		t.Fatal("not cast")
	}
	if assistTapped(g, theirs) != 0 || len(helper.ManaPool) != 0 {
		t.Errorf("Plains tapped %v, pool %d; want the floating {R}{R} spent and the Plains untouched", assistTapped(g, theirs) == 1, len(helper.ManaPool))
	}
}

// CR 800.4f: a helper who leaves paid nothing, and the cast goes on.
func TestAssistHelperLeavesPaysNothing(t *testing.T) {
	g, me, helper := assistTable(t)
	assistLands(g, me, "Swamp", 4)
	assistLands(g, helper, "Plains", 3)
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.Concede(helper.ID); err != nil {
		t.Fatal(err)
	}
	item := g.StackMeta[id]
	if item == nil || item.Paid.AssistPaid != 0 {
		t.Fatalf("item = %+v; want the spell cast unassisted", item)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts left", len(g.PendingChoices))
	}
}

// CR 800.4a: a caster who leaves has no cast to make; the prompt goes
// and nothing is paid.
func TestAssistCasterLeavesWithdrawsThePrompt(t *testing.T) {
	g, me, helper := assistTable(t)
	theirs := assistLands(g, helper, "Plains", 3)
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.Concede(me.ID); err != nil {
		t.Fatal(err)
	}
	for _, c := range g.PendingChoices {
		if c.Assist != nil {
			t.Fatal("the helper is still asked for a caster who left")
		}
	}
	if assistTapped(g, theirs) != 0 || g.StackMeta[id] != nil {
		t.Error("something was paid or cast")
	}
}

// Owner decision 5: the caster may withdraw the request.
func TestAssistCancelRequest(t *testing.T) {
	g, me, helper := assistTable(t)
	theirs := assistLands(g, helper, "Plains", 3)
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatal(err)
	}
	c := assistPrompt(t, g)
	if err := g.CancelAssistRequest(helper.ID, c.ID); !errors.Is(err, ErrPendingChoiceNotFound) {
		t.Errorf("the helper cancelling: %v, want ErrPendingChoiceNotFound", err)
	}
	if err := g.CancelAssistRequest(me.ID, uuid.Nil); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if len(g.PendingChoices) != 0 || !assistInHand(me, id) || assistTapped(g, theirs) != 0 {
		t.Error("the withdrawn cast left something behind")
	}
	// The table moves again: the caster can cast it alone now.
	assistLands(g, me, "Swamp", 4)
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("recast: %v", err)
	}
}

// One request per card: the same spell can't be parked twice.
func TestAssistOneRequestPerSpell(t *testing.T) {
	g, me, helper := assistTable(t)
	assistLands(g, helper, "Plains", 3)
	assistLands(g, g.Seats[2], "Plains", 3)
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatal(err)
	}
	if err := castAssisted(g, me, id, g.Seats[2].ID); !errors.Is(err, ErrChoicePending) {
		t.Errorf("a second request for the same spell: %v, want ErrChoicePending", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Errorf("%d prompts, want 1", len(g.PendingChoices))
	}
}

// The undo clone gets its own copy of the prompt's data.
func TestAssistPromptClones(t *testing.T) {
	g, me, helper := assistTable(t)
	assistLands(g, helper, "Plains", 3)
	id := assistSpell(me, "{3}{B}")
	if err := castAssisted(g, me, id, helper.ID); err != nil {
		t.Fatal(err)
	}
	var clone *Game
	g.WithWriteLock(func() { clone = g.cloneLocked() })
	orig := assistPrompt(t, g)
	copied := assistPrompt(t, clone)
	if orig.Assist == copied.Assist || orig.Assist.Caster != copied.Assist.Caster ||
		orig.Assist.Spell != copied.Assist.Spell || orig.Assist.Generic != copied.Assist.Generic {
		t.Error("the clone shares or lost the assist data")
	}
}
