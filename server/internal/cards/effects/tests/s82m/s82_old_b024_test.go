package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB024Muscle   = "06e0459f-d691-4470-884a-27c09b6612d2"
	oldB024Wrangler = "4021b6b4-9b87-4dd5-a5df-79540bedbe87"
	oldB024Nylea    = "43674fdf-496f-42dc-8a0a-5c42a7c7d910"
	oldB024Seven    = "526ca4a9-3f50-4f7a-8169-2bda95792401"
	oldB024Rashel   = "c0c39f80-cf12-482e-a7c4-757db85125b0"
	oldB024Depala   = "cc913f28-18ef-485e-8d2b-ae0115523020"
	oldB024Jurin    = "de384867-8bb4-4edf-b542-970538718fdd"
)

func TestOldB024CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB024Muscle:   "Muscle Sliver",
		oldB024Wrangler: "Squirrel Wrangler",
		oldB024Nylea:    "Nylea, God of the Hunt",
		oldB024Seven:    "Seven Dwarves",
		oldB024Rashel:   "Rashel, Fist of Torm",
		oldB024Depala:   "Depala, Pilot Exemplar",
		oldB024Jurin:    "Jurin, Leading the Charge",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB024Power(g *game.Game, id uuid.UUID) (int, int) {
	c := effectstest.ApaLive(g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

func TestOldB024MuscleSliverPumpsEverySliver(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	muscle := effectstest.B12Push(g, me.ID, "Muscle Sliver", "Creature — Sliver", oldB024Muscle, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Sliver", "Creature — Sliver", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, th := oldB024Power(g, muscle); p != 2 || th != 2 {
		t.Errorf("Muscle Sliver is %d/%d, want 2/2", p, th)
	}
	if p, th := oldB024Power(g, theirs); p != 2 || th != 2 {
		t.Errorf("an opponent's Sliver is %d/%d, want 2/2", p, th)
	}
	if p, th := oldB024Power(g, bear); p != 2 || th != 2 {
		t.Errorf("a non-Sliver is %d/%d, want unchanged 2/2", p, th)
	}
}

func TestOldB024SquirrelWranglerMakesTwoSquirrelsAndPumpsEverySquirrel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	wrangler := effectstest.B12Push(g, me.ID, "Squirrel Wrangler", "Creature — Human Druid", oldB024Wrangler, 2, 2)
	lands := []uuid.UUID{
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest"),
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest"),
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Squirrel", "Creature — Squirrel", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, wrangler, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a creature was accepted as the land to sacrifice")
	}
	effectstest.ApaActivate(t, g, me, wrangler, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{lands[0]}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lands[0]) {
		t.Error("the land was not sacrificed")
	}
	if n := effectstest.B27CountNamed(g, me.ID, "Squirrel"); n != 2 {
		t.Fatalf("%d Squirrel tokens, want 2", n)
	}
	effectstest.ApaMana(me, "G", "G")
	effectstest.ApaActivate(t, g, me, wrangler, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{lands[1]}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := oldB024Power(g, theirs); p != 2 || th != 2 {
		t.Errorf("an opponent's Squirrel is %d/%d, want 2/2", p, th)
	}
	if p, th := oldB024Power(g, bear); p != 2 || th != 2 {
		t.Errorf("a Bear is %d/%d, want 2/2", p, th)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Squirrel" {
			if p, th := oldB024Power(g, c.InstanceID); p != 2 || th != 2 {
				t.Errorf("a Squirrel token is %d/%d, want 2/2", p, th)
			}
		}
	}
}

func TestOldB024NyleaIsACreatureOnlyAtFiveDevotionAndGivesTrampleAndPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	nylea := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Nylea, God of the Hunt", OracleID: oldB024Nylea, ManaCost: "{3}{G}",
		TypeLine: "Legendary Enchantment Creature — God", Power: 6, Toughness: 6})
	bear := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if effectstest.ApaLive(g, nylea).IsCreature() {
		t.Fatal("Nylea is a creature with devotion 1")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("another creature has no trample")
	}
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Green Rock", TypeLine: "Artifact", ManaCost: "{G}{G}{G}{G}"})
	if !effectstest.ApaLive(g, nylea).IsCreature() {
		t.Fatal("Nylea is not a creature with devotion 5")
	}
	if effectstest.HasEffectiveKeyword(t, g, nylea, "trample") {
		t.Error("Nylea gave herself trample")
	}
	effectstest.ApaMana(me, "G", "G", "G", "G")
	effectstest.ApaActivate(t, g, me, nylea, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := oldB024Power(g, bear); p != 4 || th != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, th)
	}
}

func TestOldB024SevenDwarvesCountsOnlyYourOtherCopies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Push(g, me.ID, "Seven Dwarves", "Creature — Dwarf", oldB024Seven, 2, 2)
	if p, th := oldB024Power(g, a); p != 2 || th != 2 {
		t.Fatalf("alone it is %d/%d, want 2/2", p, th)
	}
	b := effectstest.B12Push(g, me.ID, "Seven Dwarves", "Creature — Dwarf", oldB024Seven, 2, 2)
	effectstest.B12Push(g, opp.ID, "Seven Dwarves", "Creature — Dwarf", oldB024Seven, 2, 2)
	effectstest.B12Creature(g, me.ID, "Other Dwarf", "Creature — Dwarf", 2, 2)
	for _, id := range []uuid.UUID{a, b} {
		if p, th := oldB024Power(g, id); p != 3 || th != 3 {
			t.Errorf("with one other copy of mine it is %d/%d, want 3/3", p, th)
		}
	}
}

func TestOldB024RashelGivesYourAurasExaltedForALoneAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rashel, Fist of Torm", OracleID: oldB024Rashel,
		TypeLine: "Legendary Creature — Human Knight", Power: 2, Toughness: 4, Keywords: []string{"double strike"}})
	attacker := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Test Aura", TypeLine: "Enchantment — Aura"})
	theirAura := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Aura", TypeLine: "Enchantment — Aura"})
	if !effectstest.HasEffectiveKeyword(t, g, aura, "exalted") {
		t.Fatal("my Aura has no exalted")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirAura, "exalted") {
		t.Error("an opponent's Aura has exalted")
	}
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := oldB024Power(g, attacker); p != 3 || th != 3 {
		t.Errorf("the lone attacker is %d/%d, want 3/3 (one exalted instance)", p, th)
	}
}

func TestOldB024DepalaPumpsDwarvesAndVehiclesAndDigsWhenTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	depala := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Depala, Pilot Exemplar", OracleID: oldB024Depala,
		TypeLine: "Legendary Creature — Dwarf Pilot", Power: 3, Toughness: 3})
	dwarf := effectstest.B12Creature(g, me.ID, "Dwarf", "Creature — Dwarf", 1, 1)
	theirDwarf := effectstest.B12Creature(g, opp.ID, "Their Dwarf", "Creature — Dwarf", 1, 1)
	crewed := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Cart", TypeLine: "Artifact Creature — Vehicle", Power: 3, Toughness: 3})
	parked := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Parked Cart", TypeLine: "Artifact — Vehicle"})
	if p, _ := oldB024Power(g, dwarf); p != 2 {
		t.Errorf("my other Dwarf has power %d, want 2", p)
	}
	if p, _ := oldB024Power(g, theirDwarf); p != 1 {
		t.Errorf("an opponent's Dwarf has power %d, want 1", p)
	}
	if p, _ := oldB024Power(g, depala); p != 3 {
		t.Errorf("Depala has power %d, want 3: the bonus is for others", p)
	}
	if p, _ := oldB024Power(g, crewed); p != 4 {
		t.Errorf("a Vehicle that is a creature has power %d, want 4", p)
	}
	if effectstest.ApaLive(g, parked).IsCreature() {
		t.Error("a parked Vehicle is a creature")
	}
	// The library top is: Dwarf, land, Vehicle, land, Bear (top first).
	g.WithWriteLock(func() { me.Library.Cards = nil })
	var want []uuid.UUID
	for _, c := range []game.Card{
		{Name: "Library Bear", TypeLine: "Creature — Bear"},
		{Name: "Library Land", TypeLine: "Basic Land — Forest"},
		{Name: "Library Cart", TypeLine: "Artifact — Vehicle"},
		{Name: "Library Land", TypeLine: "Basic Land — Forest"},
		{Name: "Library Dwarf", TypeLine: "Creature — Dwarf"},
	} {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
		if c.Name == "Library Cart" || c.Name == "Library Dwarf" {
			want = append(want, c.InstanceID)
		}
		g.WithWriteLock(func() { me.Library.PushTop(c) })
	}
	effectstest.AdvanceToMain(t, g)
	handBefore := me.Hand.Size()
	effectstest.PushLands(g, me.ID, 4) // four untapped Forests pay for X
	effectstest.DeclareAttack(t, g, opp.ID, depala)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.PayXPrompt(t, g, me.ID)
	if prompt.PayAmount.Max != 4 {
		t.Fatalf("X may be 0 through %d, want 0 through 4", prompt.PayAmount.Max)
	}
	if err := g.ResolvePayAmount(prompt.ID, me.ID, 4); err != nil {
		t.Fatalf("answer X: %v", err)
	}
	pay := effectstest.PayUnlessFor(g, me.ID)
	if pay == nil {
		t.Fatal("no payment prompt")
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range want {
		if !me.Hand.Contains(id) {
			t.Errorf("a revealed Dwarf or Vehicle %s did not reach the hand", id)
		}
	}
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("hand grew by %d, want 2 (the Dwarf and the Vehicle; X=4 reveals four of the five)", got)
	}
	if me.Library.Size() != 3 {
		t.Errorf("library holds %d cards, want 3 (five minus the two taken)", me.Library.Size())
	}
}

func TestOldB024JurinMustBeBlockedAndPumpsEachAttackerByTheDefendersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jurin := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Jurin, Leading the Charge", OracleID: oldB024Jurin,
		TypeLine: "Legendary Creature — Human Barbarian", Power: 4, Toughness: 6, Keywords: []string{"haste"}})
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	for range 3 {
		effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 1, 1)
	}
	effectstest.B12Creature(g, g.Seats[2].ID, "Bystander", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, jurin, ally)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := oldB024Power(g, jurin); p != 7 {
		t.Errorf("Jurin has power %d, want 4 + 3 creatures the defender controls", p)
	}
	if p, _ := oldB024Power(g, ally); p != 5 {
		t.Errorf("the other attacker has power %d, want 2 + 3", p)
	}
}

func TestOldB024JurinMustBeBlockedIfAble(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	jurin := effectstest.ReqCreature(g, me.ID, "Jurin, Leading the Charge", oldB024Jurin, 4, 6, "haste")
	a := effectstest.ReqCreature(g, opp.ID, "Wall A", "", 0, 4)
	effectstest.ReqCreature(g, opp.ID, "Wall B", "", 0, 4)
	effectstest.ReqToBlockers(t, g, jurin)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Jurin, Leading the Charge must be blocked if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := effectstest.ReqBlock(t, g, jurin, a); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with one blocker: %v", err)
	}
}
