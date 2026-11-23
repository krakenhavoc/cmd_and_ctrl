package testso

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// orcus_prince_of_undeath_test.go — #5898: a modal enters trigger whose
// chosen mode's target count and total-mana-value cap are the X the
// spell was cast with (CR 107.3m), through the engine.

const orcusOracle = "9d02b5e1-bd29-45ad-b11d-f2cd21c48979"

// Bullet indexes, in printed order.
const (
	orcusSweep = iota
	orcusReturn
)

// orcusCard is Orcus as a card in hand.
func orcusCard(owner uuid.UUID) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: "Orcus, Prince of Undeath", TypeLine: "Legendary Creature — Demon",
		OracleID: orcusOracle, ManaCost: "{X}{2}{B}{R}", Power: 5, Toughness: 3,
		Owner: owner, Controller: owner,
	}
}

// castOrcus casts Orcus for `x` from the active seat's hand and passes
// until its enters trigger asks for its mode.
func castOrcus(t *testing.T, g *game.Game, x int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	c := orcusCard(active.ID)
	active.Hand.PushTop(c)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, c.InstanceID, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("cast for X=%d: %v", x, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(c.InstanceID) {
		t.Fatalf("X=%d: Orcus did not enter", x)
	}
	return c.InstanceID
}

// orcusGraveCreature puts a creature card with mana cost `cost` into
// owner's graveyard.
func orcusGraveCreature(g *game.Game, owner uuid.UUID, name, cost string) uuid.UUID {
	p := g.PlayerByIDForEffect(owner)
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Zombie", ManaCost: cost,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
	return id
}

// orcusChoose answers Orcus's mode prompt with `mode`.
func orcusChoose(t *testing.T, g *game.Game, me uuid.UUID, mode int) {
	t.Helper()
	c := effectstest.ModePickChoiceFor(g, me)
	if c == nil {
		t.Fatal("no mode prompt for Orcus's trigger")
	}
	if c.ModeMin != 1 || c.ModeMax != 1 {
		t.Errorf("choose one: bounds %d..%d", c.ModeMin, c.ModeMax)
	}
	if err := g.ResolveModePick(c.ID, me, []int{mode}); err != nil {
		t.Fatalf("choose mode %d: %v", mode, err)
	}
}

// orcusHasHaste reads haste off the battlefield card through the
// layered view.
func orcusHasHaste(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return game.HasKeyword(&c, "haste")
		}
	}
	return false
}

func TestOrcusIsFull(t *testing.T) {
	spec, ok := effects.Lookup(orcusOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 || len(spec.Triggered) != 1 {
		t.Fatalf("one enters trigger, Full, no caveats: %+v", spec)
	}
	tr := spec.Triggered[0]
	if !tr.XFromCast || tr.Modes == nil || len(tr.Modes.Options) != 2 {
		t.Fatalf("a modal XFromCast trigger of two bullets: %+v", tr)
	}
}

// X = 4: up to four targets with total mana value 4 or less. A set
// totalling 5 is refused; a set totalling 4 returns together, with
// haste.
func TestOrcusReturnsUpToXWithTotalManaValueX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	two := orcusGraveCreature(g, me.ID, "Two Drop", "{1}{B}")
	three := orcusGraveCreature(g, me.ID, "Three Drop", "{2}{B}")
	one := orcusGraveCreature(g, me.ID, "One Drop", "{B}")
	big := orcusGraveCreature(g, me.ID, "Five Drop", "{4}{B}")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := orcusGraveCreature(g, opp.ID, "Their One Drop", "{G}")

	castOrcus(t, g, 4)
	orcusChoose(t, g, me.ID, orcusReturn)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the return")
	}
	if pick.PickTargetMin != 0 || pick.PickTargetMax != 4 {
		t.Errorf("up to X = 4 targets: %d..%d", pick.PickTargetMin, pick.PickTargetMax)
	}
	if effectstest.HasID(pick.PickTargetCards, theirs) {
		t.Error("only your graveyard")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardRefs(big)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a mana value 5 card alone at X = 4: %v, want ErrIllegalTarget", err)
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardRefs(two, three)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("total mana value 5 at X = 4: %v, want ErrIllegalTarget", err)
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardRefs(three, one)); err != nil {
		t.Fatalf("total mana value 4 at X = 4: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{three, one} {
		if !g.Battlefield.Contains(id) {
			t.Fatalf("%v did not return", id)
		}
		if !orcusHasHaste(g, id) {
			t.Errorf("%v has no haste", id)
		}
	}
	if g.Battlefield.Contains(two) || g.Battlefield.Contains(big) {
		t.Error("only the targets return")
	}
	if me.Life != 40 {
		t.Errorf("the return costs no life: %d", me.Life)
	}
}

// X = 2: the count caps at two, even of cards whose total fits.
func TestOrcusReturnCountIsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := orcusGraveCreature(g, me.ID, "Zero A", "{0}")
	b := orcusGraveCreature(g, me.ID, "Zero B", "{0}")
	c := orcusGraveCreature(g, me.ID, "Zero C", "{0}")
	castOrcus(t, g, 2)
	orcusChoose(t, g, me.ID, orcusReturn)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil || pick.PickTargetMax != 2 {
		t.Fatalf("up to two targets: %+v", pick)
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardRefs(a, b, c)); err == nil {
		t.Fatal("three targets at X = 2 were accepted")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardRefs(a, b)); err != nil {
		t.Fatalf("two targets at X = 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) || g.Battlefield.Contains(c) {
		t.Error("the two targets return, the third stays")
	}
}

// X survives a save and restore while the trigger is on the stack: the
// re-check sums the targets against the cast X, and the -X/-X bullet
// reads it.
func TestOrcusXSurvivesARestoreOnTheStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	three := orcusGraveCreature(g, me.ID, "Three Drop", "{2}{B}")
	orcus := castOrcus(t, g, 3)
	orcusChoose(t, g, me.ID, orcusReturn)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	// The walk open is a restore point too: X is read off the source
	// again as it is rebuilt.
	mid := effectstest.RestoreRoundTrip(t, g, true)
	midPick := effectstest.LatestPickTarget(mid, me.ID)
	if midPick == nil || midPick.PickTargetMax != 3 || !effectstest.HasID(midPick.PickTargetCards, three) {
		t.Fatalf("restored walk: %+v", midPick)
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardRefs(three)); err != nil {
		t.Fatalf("pick: %v", err)
	}
	item := effectstest.TriggerOnStack(g, orcus)
	if item == nil || item.XValue != 3 {
		t.Fatalf("the trigger carries X = 3: %+v", item)
	}
	restored := effectstest.RestoreRoundTrip(t, g, true)
	if it := effectstest.TriggerOnStack(restored, orcus); it == nil || it.XValue != 3 {
		t.Fatalf("restored trigger X: %+v", it)
	}
	effectstest.PassPriorityAroundTable(t, restored)
	if !restored.Battlefield.Contains(three) {
		t.Error("the restored trigger returns its mana value 3 target at X = 3")
	}
}

// The total is re-summed as the trigger resolves (CR 608.2b): a pick
// whose mana value grew past the total makes the whole selection
// illegal. Here the restored check reads the bound X, not the
// unbound catalog clause (whose Max is 0).
func TestOrcusTotalIsRecheckedAgainstX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := orcusGraveCreature(g, me.ID, "One A", "{B}")
	b := orcusGraveCreature(g, me.ID, "One B", "{B}")
	orcus := castOrcus(t, g, 2)
	orcusChoose(t, g, me.ID, orcusReturn)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardRefs(a, b)); err != nil {
		t.Fatalf("pick: %v", err)
	}
	item := effectstest.TriggerOnStack(g, orcus)
	if item == nil {
		t.Fatal("no trigger on the stack")
	}
	for _, ref := range item.Targets {
		if !g.TargetStillLegalForEffect(item, ref) {
			t.Errorf("total 2 at X = 2 is legal on the stack: %+v", ref)
		}
	}
	// Grow one pick's mana value past the cap while it waits.
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == a {
				me.Graveyard.Cards[i].ManaCost = "{1}{B}"
			}
		}
	})
	for _, ref := range item.Targets {
		if g.TargetStillLegalForEffect(item, ref) {
			t.Errorf("total 3 at X = 2: %v still legal", ref.ID)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the entire selection is illegal; nothing returns")
	}
}

// The other bullet: every other creature gets -X/-X, and you lose X.
func TestOrcusSweepsMinusXAndCostsXLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	small := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life
	orcus := castOrcus(t, g, 2)
	orcusChoose(t, g, me.ID, orcusSweep)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatalf("the sweep has no target: %+v", p)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if me.Life != life-2 {
		t.Errorf("you lose X = 2: %d → %d", life, me.Life)
	}
	if p, tg := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 3 || tg != 3 {
		t.Errorf("my giant %d/%d, want 3/3", p, tg)
	}
	if p, tg := effectstest.EffectivePower(t, g, theirs), effectstest.EffectiveToughness(t, g, theirs); p != 2 || tg != 2 {
		t.Errorf("their giant %d/%d, want 2/2", p, tg)
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 dies to -2/-2")
	}
	if p, tg := effectstest.EffectivePower(t, g, orcus), effectstest.EffectiveToughness(t, g, orcus); p != 5 || tg != 3 {
		t.Errorf("Orcus is not shrunk: %d/%d", p, tg)
	}
}

// Not cast: X is 0 (CR 107.3m), so the return takes no target and the
// sweep does nothing.
func TestOrcusNotCastHasXZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gy := orcusGraveCreature(g, me.ID, "Zero Drop", "{0}")
	effectstest.AdvanceToMain(t, g)
	c := orcusCard(me.ID)
	me.Graveyard.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("put Orcus onto the battlefield: %v", err)
	}
	g.RunStateChecksForTest()
	orcusChoose(t, g, me.ID, orcusReturn)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatalf("X = 0 asks for no target: %+v", p)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(gy) {
		t.Error("X = 0 returns nothing")
	}
}
