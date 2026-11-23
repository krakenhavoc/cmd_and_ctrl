package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b01_test.go — Innistrad: Midnight Hunt batch 1: plain creatures
// with enters, dies and pump abilities (S82, #5497).

const (
	mb1Rejuvenator = "c6e92911-a100-43eb-ab4b-3778b0db1918"
	mb1Silversmith = "a0ee4faa-1ab0-4b4f-a564-ee2358500daf"
	mb1SnarlWolf   = "2c181479-6df0-44ef-a7ce-d1d592735bc6"
	mb1Trapper     = "d4da20c7-ae4a-44c6-8f03-9e719ccee8c9"
	mb1Bounding    = "257f7c29-8857-429e-a421-9262982804e0"
	mb1Cathars     = "193f9b12-ee34-40c2-9315-6794416a4f62"
	mb1Occultist   = "e51f5b41-5465-4b1d-b96a-544762657ca5"
	mb1Gryff       = "f5c07218-c8bf-4990-a366-f3c065d0b30c"
	mb1Weaver      = "f0a7ef59-7897-4e84-a85e-1fcbe3dc9f6a"
	mb1Pestilent   = "39940633-ef64-408e-ae8f-ade201fe7f06"
	mb1Harrier     = "84c7d60a-7a7a-4344-8db5-d658792e3ac4"
	mb1Intruder    = "4c591977-f2af-4a6e-a0be-3dfb45a3598b"
	mb1Outrider    = "a5a9a314-81a0-4b11-9cf9-daeaea98b33c"
)

func TestS82MidB01CardsAreRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		mb1Rejuvenator: "Dawnhart Rejuvenator", mb1Silversmith: "Gavony Silversmith",
		mb1SnarlWolf: "Snarling Wolf", mb1Trapper: "Gavony Trapper", mb1Bounding: "Bounding Wolf",
		mb1Cathars: "Clarion Cathars", mb1Occultist: "Novice Occultist", mb1Gryff: "Loyal Gryff",
		mb1Weaver: "Brood Weaver", mb1Pestilent: "Pestilent Wolf", mb1Harrier: "Lambholt Harrier",
		mb1Intruder: "Nebelgast Intruder", mb1Outrider: "Odric's Outrider",
	} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// mb1Perm puts a catalog creature on the battlefield, old enough to tap.
func mb1Perm(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// mb1Enter moves a catalog creature from hand onto the battlefield so its
// enters trigger fires.
func mb1Enter(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID})
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("enter the battlefield: %v", err)
	}
	return id
}

func mb1Kill(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{id}, game.DestroyOptions{}) })
	g.RunStateChecksForTest()
}

func mb1Pick(t *testing.T, g *game.Game, who uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.LatestPickTarget(g, who)
	if ch == nil {
		t.Fatal("no target prompt")
	}
	refs := make([]game.TargetRef, len(ids))
	for i, id := range ids {
		refs[i] = game.TargetRef{Kind: game.TargetCard, ID: id}
	}
	if err := g.ResolvePickTargets(ch.ID, who, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func mb1Power(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	e := effectstest.GaLayered(t, g, id).Effective()
	return e.Power, e.Toughness
}

func mb1Mana(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("mana: %v", err)
		}
	})
}

func TestDawnhartRejuvenatorGainsThreeAndTapsForAnyColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	before := me.Life
	mb1Enter(t, g, me, "Dawnhart Rejuvenator", "Creature — Human Warlock", mb1Rejuvenator, 2, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("life %d, want %d", me.Life, before+3)
	}
	id := mb1Perm(g, me.ID, "Dawnhart Rejuvenator", "Creature — Human Warlock", mb1Rejuvenator, 2, 4)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		// A colour prompt is how the engine asks which colour.
		t.Fatalf("mana ability: %v", err)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick != nil {
		if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
			t.Fatalf("colour: %v", err)
		}
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool %d, want 1", me.ManaPool.Total())
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a tapped Rejuvenator made mana again")
	}
}

func TestGavonySilversmithCountersUpToTwoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, g.Seats[1].ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	mb1Enter(t, g, me, "Gavony Silversmith", "Creature — Human Soldier", mb1Silversmith, 2, 3)
	mb1Pick(t, g, me.ID, a, b)
	if effectstest.PlusOnes(g, a) != 1 || effectstest.PlusOnes(g, b) != 1 || effectstest.PlusOnes(g, c) != 0 {
		t.Errorf("counters a=%d b=%d c=%d, want 1 1 0", effectstest.PlusOnes(g, a), effectstest.PlusOnes(g, b), effectstest.PlusOnes(g, c))
	}
}

func TestSnarlingWolfPumpsOnlyOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	w := mb1Perm(g, me.ID, "Snarling Wolf", "Creature — Wolf", mb1SnarlWolf, 1, 1)
	mb1Mana(t, g, me, "{C}{G}{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := mb1Power(t, g, w); p != 3 || tough != 3 {
		t.Errorf("got %d/%d, want 3/3", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a second activation in one turn was allowed")
	}
}

func TestGavonyTrapperTapsATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tr := mb1Perm(g, me.ID, "Gavony Trapper", "Creature — Human Soldier", mb1Trapper, 0, 2)
	foe := effectstest.B12Creature(g, g.Seats[1].ID, "Foe", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, tr, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(foe)}); err == nil {
		t.Fatal("activated without paying {2}")
	}
	mb1Mana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tr, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(foe)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.BattlefieldCardFor(g, foe); !c.Tapped {
		t.Error("the target is not tapped")
	}
	if c := effectstest.BattlefieldCardFor(g, tr); !c.Tapped {
		t.Error("the Trapper did not tap for its own cost")
	}
}

func TestBoundingWolfHasFlashAndReach(t *testing.T) {
	spec, _ := effects.Lookup(mb1Bounding)
	if len(spec.PrintedKeywords) != 2 || spec.PrintedKeywords[0] != "flash" || spec.PrintedKeywords[1] != "reach" {
		t.Errorf("keywords %v", spec.PrintedKeywords)
	}
}

func TestClarionCatharsMakesAHumanToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mb1Enter(t, g, me, "Clarion Cathars", "Creature — Human Knight", mb1Cathars, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Human", me.ID); n != 1 {
		t.Errorf("%d Human tokens, want 1", n)
	}
}

func TestNoviceOccultistDrawsAndLosesOneWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := mb1Perm(g, me.ID, "Novice Occultist", "Creature — Human Wizard", mb1Occultist, 1, 2)
	hand, life := me.Hand.Size(), me.Life
	mb1Kill(g, id)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d→%d, life %d→%d", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestLoyalGryffMayReturnAnotherCreature(t *testing.T) {
	for _, decline := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
		foe := effectstest.B12Creature(g, g.Seats[1].ID, "Foe", "Creature — Bear", 2, 2)
		gryff := mb1Enter(t, g, me, "Loyal Gryff", "Creature — Hippogriff", mb1Gryff, 2, 2)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
		if pick == nil {
			t.Fatal("no creature choice")
		}
		if effectstest.HasID(pick.ChooseCards, gryff) || effectstest.HasID(pick.ChooseCards, foe) || !effectstest.HasID(pick.ChooseCards, other) {
			t.Errorf("offered %v", pick.ChooseCards)
		}
		picks := []uuid.UUID{other}
		if decline {
			picks = nil
		}
		if err := g.ResolveOwnPermanents(pick.ID, me.ID, picks); err != nil {
			t.Fatalf("pick: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Contains(other); got == decline {
			t.Errorf("decline=%v: Other in hand = %v", decline, got)
		}
		if !effects.OnBattlefield(g, gryff) {
			t.Error("the Gryff left")
		}
	}
}

func TestBroodWeaverLeavesASpiderWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := mb1Perm(g, me.ID, "Brood Weaver", "Creature — Spider", mb1Weaver, 2, 4)
	mb1Kill(g, id)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Spider", me.ID); n != 1 {
		t.Fatalf("%d Spider tokens, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spider" {
			if p, tough := mb1Power(t, g, c.InstanceID); p != 1 || tough != 2 || !game.HasKeyword(&c, "reach") {
				t.Errorf("Spider is %d/%d, reach=%v", p, tough, game.HasKeyword(&c, "reach"))
			}
		}
	}
}

func TestPestilentWolfGainsDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	w := mb1Perm(g, me.ID, "Pestilent Wolf", "Creature — Wolf", mb1Pestilent, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without mana")
	}
	mb1Mana(t, g, me, "{G}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, w); !game.HasKeyword(&c, "deathtouch") {
		t.Error("no deathtouch")
	}
}

func TestLambholtHarrierStopsATargetFromBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	h := mb1Perm(g, me.ID, "Lambholt Harrier", "Creature — Wolf", mb1Harrier, 2, 2)
	foe := effectstest.B12Creature(g, g.Seats[1].ID, "Foe", "Creature — Bear", 2, 2)
	mb1Mana(t, g, me, "{R}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, h, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(foe)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AvB13Restrictions(g, foe)&game.CantBlock == 0 {
		t.Error("the target can still block")
	}
}

func TestNebelgastIntruderShrinksAnOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	foe := effectstest.B12Creature(g, g.Seats[1].ID, "Foe", "Creature — Bear", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	mb1Enter(t, g, me, "Nebelgast Intruder", "Creature — Spirit", mb1Intruder, 2, 1)
	mb1Pick(t, g, me.ID, foe)
	if p, _ := mb1Power(t, g, foe); p != 1 {
		t.Errorf("Foe power %d, want 1", p)
	}
	if p, _ := mb1Power(t, g, mine); p != 3 {
		t.Errorf("own creature power %d, want 3", p)
	}
}

func TestOdricsOutriderCountersOnItsOwnAndOthersDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	out := mb1Perm(g, me.ID, "Odric's Outrider", "Creature — Human Knight", mb1Outrider, 2, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	foe := effectstest.B12Creature(g, g.Seats[1].ID, "Foe", "Creature — Bear", 2, 2)
	mb1Kill(g, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, out)+effectstest.PlusOnes(g, bear) != 0 {
		t.Error("an opponent's creature dying triggered it")
	}
	mb1Kill(g, bear)
	mb1Pick(t, g, me.ID, out)
	if effectstest.PlusOnes(g, out) != 1 {
		t.Errorf("Outrider counters %d, want 1", effectstest.PlusOnes(g, out))
	}
	mb1Kill(g, out)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, out) {
		t.Error("the Outrider survived")
	}
}
