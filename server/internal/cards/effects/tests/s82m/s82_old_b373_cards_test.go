package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o373Oracle     = "0503fc7e-6819-4979-b0df-7de23de4558b"
	o373Ornitharch = "19639595-c59b-4233-a167-f5cffcb0e38b"
	o373Demolok    = "2341b59b-ce46-4efe-be94-c3851d3cd9e8"
	o373Phoenix    = "24b9204f-ecc8-472e-9296-f999adf1ab1b"
	o373Pharagax   = "69a134f2-c0c4-496f-b354-8325437af8ca"
	o373Harpy      = "9c38b337-487c-4cef-843b-b0a2ed3f1b27"
	o373Siren      = "c5e7ba95-c9b9-45e8-9551-dba892d90683"
	o373Brute      = "ee1d97e9-847d-4295-a9c1-7ce8006fb0bd"
	o373Snake      = "ef0c8193-8cbe-409e-9c2e-96b46ffcc420"
	o373Fanatic    = "f1072af9-9e6f-45ce-8c4b-8e678e7da8ce"
	o373Ravager    = "fa1a5797-3076-48a4-b64b-ce3dfe2a3e7b"
)

func TestO373CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		o373Oracle: "Oracle of Bones", o373Ornitharch: "Ornitharch", o373Demolok: "Nessian Demolok",
		o373Phoenix: "Flame-Wreathed Phoenix", o373Pharagax: "Pharagax Giant", o373Harpy: "Shrike Harpy",
		o373Siren: "Siren of the Fanged Coast", o373Brute: "Thunder Brute", o373Snake: "Snake of the Golden Grove",
		o373Fanatic: "Fanatic of Xenagos", o373Ravager: "Nessian Wilds Ravager",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// o373Enter casts the creature from the active seat's hand, has the caster
// choose the seat after them as the tribute opponent, and has that opponent
// pay or decline. It stops before the "if tribute wasn't paid" trigger.
func o373Enter(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int, pay bool) uuid.UUID {
	t.Helper()
	caster := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	id := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: caster.ID, Controller: caster.ID})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var c *game.PendingChoice
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceEntryController {
			c = pc
		}
	}
	if c == nil {
		t.Fatalf("%s: no tribute opponent prompt", name)
	}
	for i, opt := range c.PickOptions {
		if opt.Player == opp.ID {
			if err := g.ResolveEntryController(c.ID, c.Chooser, i); err != nil {
				t.Fatalf("ResolveEntryController: %v", err)
			}
		}
	}
	var p *game.PendingChoice
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceOptionalReplacement {
			p = pc
		}
	}
	if p == nil {
		t.Fatalf("%s: no tribute payment prompt", name)
	}
	if err := g.ResolveOptionalReplacement(p.ID, opp.ID, pay); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	return id
}

// o373Settle resolves the stack and answers an open target prompt with `target`.
func o373Settle(t *testing.T, g *game.Game, who uuid.UUID, target uuid.UUID) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	if target != uuid.Nil && effectstest.LatestPickTarget(g, who) != nil {
		effectstest.AnswerPickTarget(t, g, target)
		effectstest.PassPriorityAroundTable(t, g)
	}
}

func o373Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCardByID(g, id)
	if !ok {
		t.Fatal("the creature is not on the battlefield")
	}
	return c.Counters[game.CounterPlusOne]
}

// Paid or not, each creature enters with its printed counters or none.
func TestO373TributePaidPutsTheCountersAndDoesNothingElse(t *testing.T) {
	for _, c := range []struct {
		name, oracle string
		n            int
	}{
		{"Ornitharch", o373Ornitharch, 2}, {"Snake of the Golden Grove", o373Snake, 3},
		{"Pharagax Giant", o373Pharagax, 2}, {"Thunder Brute", o373Brute, 3},
		{"Fanatic of Xenagos", o373Fanatic, 1}, {"Flame-Wreathed Phoenix", o373Phoenix, 2},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		life := me.Life
		oppLife := g.Seats[(g.Turn.ActiveSeat+1)%4].Life
		id := o373Enter(t, g, c.name, "Creature — Test", c.oracle, 3, 3, true)
		o373Settle(t, g, me.ID, uuid.Nil)
		if got := o373Counters(t, g, id); got != c.n {
			t.Errorf("%s: %d counters, want %d", c.name, got, c.n)
		}
		if me.Life != life || g.Seats[(g.Turn.ActiveSeat+1)%4].Life != oppLife {
			t.Errorf("%s: life changed although tribute was paid", c.name)
		}
		if effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%s: has haste although tribute was paid", c.name)
		}
	}
}

func TestO373OrnitharchMakesTwoFlyingBirdsWhenTributeIsNotPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	o373Enter(t, g, "Ornitharch", "Creature — Archon", o373Ornitharch, 3, 3, false)
	o373Settle(t, g, me.ID, uuid.Nil)
	birds := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Bird" && c.Controller == me.ID && c.Power == 1 && c.Toughness == 1 {
			birds++
		}
	}
	if birds != 2 {
		t.Fatalf("%d Bird tokens, want 2", birds)
	}
}

func TestO373SnakeGainsFourLifeWhenTributeIsNotPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	o373Enter(t, g, "Snake of the Golden Grove", "Creature — Snake", o373Snake, 4, 4, false)
	o373Settle(t, g, me.ID, uuid.Nil)
	if me.Life != life+4 {
		t.Fatalf("life %d, want %d", me.Life, life+4)
	}
}

func TestO373PharagaxDealsFiveToEachOpponentWhenTributeIsNotPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		before[s.ID] = s.Life
	}
	o373Enter(t, g, "Pharagax Giant", "Creature — Giant", o373Pharagax, 3, 3, false)
	o373Settle(t, g, me.ID, uuid.Nil)
	for _, s := range g.Seats {
		want := before[s.ID]
		if s.ID != me.ID {
			want -= 5
		}
		if s.Life != want {
			t.Errorf("seat %s life %d, want %d", s.ID, s.Life, want)
		}
	}
}

func TestO373ThunderBruteGainsHasteUntilEndOfTurnWhenTributeIsNotPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := o373Enter(t, g, "Thunder Brute", "Creature — Cyclops", o373Brute, 5, 5, false)
	o373Settle(t, g, me.ID, uuid.Nil)
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Fatal("no haste")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Error("haste outlasted the turn")
	}
}

func TestO373FanaticGetsPlusOneAndHasteWhenTributeIsNotPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := o373Enter(t, g, "Fanatic of Xenagos", "Creature — Centaur Warrior", o373Fanatic, 3, 3, false)
	o373Settle(t, g, me.ID, uuid.Nil)
	if p, tough := effectstest.B38PT(t, g, id); p != 4 || tough != 4 {
		t.Errorf("%d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Error("no haste")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if p, _ := effectstest.B38PT(t, g, id); p != 3 {
		t.Errorf("power %d after the turn, want 3", p)
	}
}

func TestO373ShrikeHarpyMakesTargetOpponentSacrificeWhenTributeIsNotPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+2)%4]
	bear := effectstest.PushVanillaCreature(g, victim.ID, "Their Bear", 2, 2)
	o373Enter(t, g, "Shrike Harpy", "Creature — Harpy", o373Harpy, 2, 2, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerPickTargetPlayer(t, g, victim.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, victim.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the targeted opponent did not sacrifice their creature")
	}
}

func TestO373SirenStealsACreatureForGoodWhenTributeIsNotPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	theirs := effectstest.PushVanillaCreature(g, g.Seats[(g.Turn.ActiveSeat+2)%4].ID, "Their Bear", 2, 2)
	o373Enter(t, g, "Siren of the Fanged Coast", "Creature — Siren", o373Siren, 1, 1, false)
	o373Settle(t, g, me.ID, theirs)
	c, ok := effectstest.BattlefieldCard(g, theirs)
	if !ok || c.Controller != me.ID {
		t.Fatalf("not stolen: %+v", c)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.Controller != me.ID {
		t.Error("control wore off")
	}
}

func TestO373SirenDoesNotStealWhenTributeIsPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	theirs := effectstest.PushVanillaCreature(g, g.Seats[(g.Turn.ActiveSeat+2)%4].ID, "Their Bear", 2, 2)
	o373Enter(t, g, "Siren of the Fanged Coast", "Creature — Siren", o373Siren, 1, 1, true)
	o373Settle(t, g, me.ID, theirs)
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.Controller == me.ID {
		t.Fatal("stolen although tribute was paid")
	}
}

func TestO373DemolokDestroysANoncreaturePermanentWhenTributeIsNotPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, g.Seats[(g.Turn.ActiveSeat+2)%4].ID, "Their Rock", "Artifact")
	bear := effectstest.PushVanillaCreature(g, g.Seats[(g.Turn.ActiveSeat+2)%4].ID, "Their Bear", 2, 2)
	o373Enter(t, g, "Nessian Demolok", "Creature — Beast", o373Demolok, 3, 3, false)
	o373Settle(t, g, me.ID, rock)
	if g.Battlefield.Contains(rock) {
		t.Error("the noncreature permanent survived")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was destroyed")
	}
}

func TestO373DemolokOnlyTargetsNoncreaturePermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVanillaCreature(g, g.Seats[(g.Turn.ActiveSeat+2)%4].ID, "Their Bear", 2, 2)
	o373Enter(t, g, "Nessian Demolok", "Creature — Beast", o373Demolok, 3, 3, false)
	o373Settle(t, g, me.ID, bear)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature was destroyed by a noncreature-permanent clause")
	}
}

func TestO373RavagerMayFightAnotherCreatureWhenTributeIsNotPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVanillaCreature(g, g.Seats[(g.Turn.ActiveSeat+2)%4].ID, "Their Bear", 2, 2)
	id := o373Enter(t, g, "Nessian Wilds Ravager", "Creature — Hydra", o373Ravager, 6, 6, false)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Error("the bear survived 6 damage")
	}
	c, ok := effectstest.BattlefieldCardByID(g, id)
	if !ok || c.DamageMarked != 2 {
		t.Errorf("the Ravager took %d damage, want 2 (ok=%v)", c.DamageMarked, ok)
	}
}

func TestO373PhoenixGainsHasteAndReturnsToHandWhenItDiesIfTributeIsNotPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := o373Enter(t, g, "Flame-Wreathed Phoenix", "Creature — Phoenix", o373Phoenix, 3, 3, false)
	o373Settle(t, g, me.ID, uuid.Nil)
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Fatal("no haste")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Fatal("haste wore off: the grant has no stated duration")
	}
	if err := g.DestroyPermanentForEffect(id); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) {
		t.Fatal("the Phoenix did not return to its owner's hand")
	}
}

func TestO373OracleOfBonesOffersAFreeInstantOrSorceryWhenTributeIsNotPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	spell := effectstest.HandCardFull(me, "Test Sorcery", "Sorcery", "{4}{R}", "", []string{"R"})
	o373Enter(t, g, "Oracle of Bones", "Creature — Minotaur Shaman", o373Oracle, 3, 1, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) == nil && effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) == nil {
		t.Fatalf("no free-cast offer: %+v", g.PendingChoices)
	}
	if !me.Hand.Contains(spell) {
		t.Error("the spell left the hand before the offer was answered")
	}
}

func TestO373OracleOfBonesAsksNothingWhenTributeIsPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	effectstest.HandCardFull(me, "Test Sorcery", "Sorcery", "{4}{R}", "", []string{"R"})
	o373Enter(t, g, "Oracle of Bones", "Creature — Minotaur Shaman", o373Oracle, 3, 1, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("a free cast was offered although tribute was paid")
	}
}
