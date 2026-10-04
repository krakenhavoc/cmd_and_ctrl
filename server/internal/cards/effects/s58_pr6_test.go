package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S58 PR 6: Lercron-Xenagos, red-green combat (#2061).

const (
	s58p6AnzragOracle      = "4adcd967-9ff7-4940-b8b9-0c4215bbcb75"
	s58p6BerserkOracle     = "8b67d192-9a05-4a47-82ae-5fc4b7834d88"
	s58p6GwennaOracle      = "7d95e72a-461e-4993-a190-847466a4b17c"
	s58p6LastMarchOracle   = "3999cd64-ff5c-4e2c-8aa6-d14f9f8a2b4c"
	s58p6MageSlayerOracle  = "db444262-c40f-4c80-9d14-265c32f77cf1"
	s58p6RavineOracle      = "8d38194e-b607-4ff4-9c19-0e8636d463bf"
	s58p6SarythOracle      = "4e07a0b3-f340-4ed1-a8a9-c25fe6f37fb3"
	s58p6ShadowOracle      = "4d81db0d-2f56-42ac-92ce-249a34d262d4"
	s58p6SomberwaldOracle  = "f249db15-09fa-4a1c-9461-9ae803c219b6"
	s58p6SkullsporeOracle  = "ba26ff0a-e714-44f2-95cf-1a5a6088edf9"
	s58p6TraverseOracle    = "0d49cf51-af1f-4c17-9cf4-b82bc7c2c72e"
	s58p6UnleashFuryOracle = "4eaa100c-f1ef-4a2b-9370-5dad7f3a95f2"
	s58p6XenagosOracle     = "cb15a8dd-57fe-466f-847e-66476b690a1f"
	s58p6ZopandrelOracle   = "b168e4e6-f572-4d9f-b98f-95b2611354cb"
	s58p6GrowingRiteOracle = "ea9c459a-6047-43aa-968f-a582be4000e8"
	s58p6DamnationOracle   = "d57a8f0b-7989-4db5-8756-6f2690097252"
)

// s58p6View is the battlefield card after a layer recompute.
func s58p6View(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if c := findBattlefieldCardByID(g, id); c != nil {
			out, found = *c, true
		}
	})
	if !found {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return out
}

// s58p6Push seeds a catalog permanent with a mana cost and a body, not
// summoning sick.
func s58p6Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle, cost string, p, tough int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

// s58p6CastNow casts a card from p's hand at the current step, with no
// advancing, so the cast window is whatever step the table is in.
func s58p6CastNow(g *game.Game, p *game.Player, name, typeLine, oracle string, targets []game.TargetRef) error {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return g.CastSpell(p.ID, id, game.CastSpellParams{Targets: targets})
}

func s58p6Activate(t *testing.T, g *game.Game, who uuid.UUID, card uuid.UUID, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(who, card, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
}

// --- Unleash Fury ------------------------------------------------------

// CR 701.10b: +X/+0 with X the power as it resolves, counters included.
// A pump afterwards is not doubled again.
func TestUnleashFuryDoublesThePowerItReads(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	findBattlefieldCardByID(g, bear).Counters = map[string]int{game.CounterPlusOne: 1}

	castCatalogSpell(t, g, "Unleash Fury", "Instant", s58p6UnleashFuryOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	c := s58p6View(t, g, bear)
	if c.CurrentPower() != 6 || c.CurrentToughness() != 3 {
		t.Fatalf("doubled 3/3 is %d/%d, want 6/3", c.CurrentPower(), c.CurrentToughness())
	}
	g.WithWriteLock(func() {
		_ = BoostUntilEOT{Target: bear, Power: 3, Toughness: 3}.Apply(NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	if got := s58p6View(t, g, bear).CurrentPower(); got != 9 {
		t.Errorf("a later +3 gives %d, want 9 (doubling is a one-time +3)", got)
	}
}

// --- Xenagos, God of Revels ------------------------------------------

// CR 700.5: devotion to red and green counts a {R/G} hybrid once.
func TestXenagosCountsAHybridSymbolOnceTowardSeven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	xen := s58p6Push(g, me.ID, "Xenagos, God of Revels", "Legendary Enchantment Creature — God", s58p6XenagosOracle, "{3}{R}{G}", 6, 5)
	s58p6Push(g, me.ID, "Red Thing", "Enchantment", "", "{R}{R}{G}", 0, 0)
	s58p6Push(g, me.ID, "Hybrid Thing", "Enchantment", "", "{R/G}", 0, 0)
	s58p6Push(g, g.Seats[1].ID, "Their Green", "Enchantment", "", "{G}{G}{G}", 0, 0)

	if c := s58p6View(t, g, xen); c.IsCreature() {
		t.Fatal("devotion 6 (the hybrid once, the opponent's not at all): Xenagos is a creature")
	}
	if !game.HasKeyword(ptrTo(s58p6View(t, g, xen)), "indestructible") {
		t.Error("Xenagos is indestructible whether or not it is a creature")
	}
	s58p6Push(g, me.ID, "Green Thing", "Enchantment", "", "{G}", 0, 0)
	if c := s58p6View(t, g, xen); !c.IsCreature() || c.CurrentPower() != 6 {
		t.Errorf("devotion 7: creature=%v power=%d, want a 6/5 creature", c.IsCreature(), c.CurrentPower())
	}
}

func ptrTo(c game.Card) *game.Card { return &c }

// The combat trigger targets ANOTHER creature you control: a summoning-
// sick 2/3 becomes a hasty 4/5 and may attack; Xenagos is never offered.
func TestXenagosGivesAnotherCreatureHasteAndItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	xen := s58p6Push(g, me.ID, "Xenagos, God of Revels", "Legendary Enchantment Creature — God", s58p6XenagosOracle, "{3}{R}{G}", 6, 5)
	s58p6Push(g, me.ID, "Devotion", "Enchantment", "", "{R}{R}{G}{G}{G}", 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 3)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	findBattlefieldCardByID(g, bear).SummonedThisTurn = true
	if !s58p6View(t, g, xen).IsCreature() {
		t.Fatal("setup: Xenagos should be a creature at devotion 7")
	}

	advanceTo(t, g, game.StepBeginCombat)
	c := pickTargetPrompt(t, g)
	if slices.Contains(c.PickTargetCards, xen) {
		t.Error("Xenagos was offered as its own target")
	}
	if !slices.Contains(c.PickTargetCards, other) {
		t.Error("another creature you control was not offered")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)

	got := s58p6View(t, g, bear)
	if got.CurrentPower() != 4 || got.CurrentToughness() != 5 {
		t.Errorf("bear is %d/%d, want 4/5", got.CurrentPower(), got.CurrentToughness())
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Errorf("the hasty bear could not attack: %v", err)
	}
}

// --- Anzrag, the Quake-Mole ------------------------------------------

func TestAnzragBlockedUntapsYourCreaturesAndAddsACombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	anzrag := b12Push(g, me.ID, "Anzrag, the Quake-Mole", "Legendary Creature — Mole God", s58p6AnzragOracle, 8, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	w1 := b12Creature(g, opp.ID, "Wall One", "Creature — Wall", 0, 9)
	w2 := b12Creature(g, opp.ID, "Wall Two", "Creature — Wall", 0, 9)
	start := len(g.Events)

	declareAttack(t, g, opp.ID, anzrag, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, w := range []uuid.UUID{w1, w2} {
		if err := g.DeclareBlocker(w, anzrag); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	if n := triggersOnStackFrom(g, anzrag); n != 1 {
		t.Fatalf("two blockers: %d triggers, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if isTapped(g, anzrag) || isTapped(g, bear) {
		t.Error("the trigger did not untap the creatures you control")
	}
	const label = "Anzrag, the Quake-Mole — untap each creature you control, additional combat"
	if n := countResolved(g, start, label); n != 1 {
		t.Errorf("resolved %d times, want 1", n)
	}
	for i := 0; i < 8 && g.Turn.Step != game.StepEndCombat; i++ {
		b26AnswerDamageAssignments(t, g)
		passPriorityAroundTable(t, g)
		stepOnce(t, g)
	}
	if g.Turn.Step != game.StepEndCombat {
		t.Fatalf("never reached end of combat (at %s)", g.Turn.Step)
	}
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Errorf("after combat: %+v, want an additional combat", next)
	}
}

func TestAnzragMustBeBlockedOnceActivated(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	anzrag := b12Push(g, me.ID, "Anzrag, the Quake-Mole", "Legendary Creature — Mole God", s58p6AnzragOracle, 8, 4)
	wall := reqCreature(g, opp.ID, "Wall", "", 0, 9)
	advanceTo(t, g, game.StepPrecombatMain)
	fillPool(me, 3)
	fillPoolColored(me, "R", 2)
	fillPoolColored(me, "G", 2)
	s58p6Activate(t, g, me.ID, anzrag, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	reqToBlockers(t, g, anzrag)
	reqRefusal(t, g.PassPriority())
	if err := reqBlock(t, g, anzrag, wall); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
}

// --- Berserk -----------------------------------------------------------

// Trample and +X/+0 for X its power, and the delayed trigger destroys
// the creature at the end step because it attacked.
func TestBerserkPumpsAndDestroysAnAttackerAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	castCatalogSpell(t, g, "Berserk", "Instant", s58p6BerserkOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	c := s58p6View(t, g, bear)
	if c.CurrentPower() != 6 || c.CurrentToughness() != 3 || !game.HasKeyword(&c, "trample") {
		t.Fatalf("bear is %d/%d trample=%v, want 6/3 with trample", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "trample"))
	}
	life := opp.Life
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 6 {
		t.Errorf("combat damage %d, want 6", got)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the creature attacked and survived the end step")
	}
}

// A creature that did not attack is not destroyed.
func TestBerserkSparesACreatureThatDidNotAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	castCatalogSpell(t, g, "Berserk", "Instant", s58p6BerserkOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature that never attacked was destroyed")
	}
}

// The cast window: precombat main and the combat steps before damage
// yes; once combat damage has begun, and in the postcombat main phase,
// no.
func TestBerserkCastOnlyBeforeTheCombatDamageStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)

	advanceTo(t, g, game.StepBeginCombat)
	if err := s58p6CastNow(g, me, "Berserk", "Instant", s58p6BerserkOracle, cardTarget(bear)); err != nil {
		t.Fatalf("beginning of combat: %v", err)
	}
	passPriorityAroundTable(t, g)
	attackWith(t, g, opp.ID, bear)
	if err := s58p6CastNow(g, me, "Berserk", "Instant", s58p6BerserkOracle, cardTarget(bear)); err == nil {
		t.Error("cast during the combat damage step")
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	if err := s58p6CastNow(g, me, "Berserk", "Instant", s58p6BerserkOracle, cardTarget(bear)); err == nil {
		t.Error("cast in the postcombat main phase")
	}
}

// --- Gwenna, Eyes of Gaea --------------------------------------------

func TestGwennaManaPaysCreatureSpellsAndCreatureAbilitiesOnly(t *testing.T) {
	g, me, _ := spendTable(t)
	gwenna := pushCatalogPermanent(g, me.ID, "Gwenna, Eyes of Gaea", "Legendary Creature — Elf Druid Scout", s58p6GwennaOracle, false)
	if err := g.ActivateManaAbility(me.ID, gwenna, 0, game.ManaAbilityParams{Colors: []string{"R", "U"}}); err != nil {
		t.Fatalf("tap Gwenna: %v", err)
	}
	if got := poolColors(me); len(got) != 2 {
		t.Fatalf("pool = %v, want two mana", got)
	}
	cost, _ := game.ParseCost("{R}{U}")
	creature := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{R}{U}"}
	sorcery := game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{R}{U}"}
	rock := game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{2}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(creature)) {
		t.Error("refused a creature spell")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(creature)) {
		t.Error("refused an ability of a creature source")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(sorcery)) {
		t.Error("paid for a sorcery")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(rock)) {
		t.Error("paid for an artifact's ability")
	}
}

func TestGwennaGrowsAndUntapsOnAFivePowerCreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gwenna := b12Push(g, me.ID, "Gwenna, Eyes of Gaea", "Legendary Creature — Elf Druid Scout", s58p6GwennaOracle, 2, 3)
	findBattlefieldCardByID(g, gwenna).Tapped = true
	advanceTo(t, g, game.StepPrecombatMain)

	small := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: small, Name: "Four", TypeLine: "Creature — Beast", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, small, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if plusOneCounters(g, gwenna) > 0 || !isTapped(g, gwenna) {
		t.Fatal("a power-4 creature spell triggered Gwenna")
	}
	big := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: big, Name: "Five", TypeLine: "Creature — Beast", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if plusOneCounters(g, gwenna) != 1 || isTapped(g, gwenna) {
		t.Errorf("counters=%d tapped=%v, want one counter and untapped", plusOneCounters(g, gwenna), isTapped(g, gwenna))
	}
}

// --- Last March of the Ents ------------------------------------------

func TestLastMarchDrawsTheGreatestToughnessThenPutsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 3)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 5; i++ {
		stapleLibraryCard(me, "Spare", "Sorcery")
	}
	elk := handCardForTest(me, "Elk", "Creature — Elk", "")
	ox := handCardForTest(me, "Ox", "Creature — Ox", "")
	handCardForTest(me, "Bolt", "Instant", "")
	if spec, ok := Lookup(s58p6LastMarchOracle); !ok || !spec.CantBeCountered {
		t.Error("Last March of the Ents must not be counterable")
	}
	castCatalogSpell(t, g, "Last March of the Ents", "Sorcery", s58p6LastMarchOracle, nil)
	before := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 3 {
		t.Fatalf("drew %d, want 3 (the greatest toughness)", got)
	}
	answerChooseCards(t, g, me.ID, elk, ox)
	if !g.Battlefield.Contains(elk) || !g.Battlefield.Contains(ox) {
		t.Error("the chosen creature cards are not on the battlefield")
	}
}

// --- Mage Slayer -------------------------------------------------------

// The equipped creature deals damage equal to its power to the player
// it attacks, as the trigger resolves, before combat damage.
func TestMageSlayerHitsTheDefenderForThePower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	slayer := b12Push(g, me.ID, "Mage Slayer", "Artifact — Equipment", s58p6MageSlayerOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	fillPool(me, 3)
	s58p6Activate(t, g, me.ID, slayer, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	passPriorityAroundTable(t, g)

	life := opp.Life
	declareAttack(t, g, opp.ID, bear)
	if triggerOnStack(g, slayer) == nil {
		t.Fatal("no Mage Slayer trigger")
	}
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("trigger dealt %d, want 3", got)
	}
}

// --- Raging Ravine -------------------------------------------------------

func s58p6Ravine(g *game.Game, owner uuid.UUID) uuid.UUID {
	return s58p6Push(g, owner, "Raging Ravine", "Land", s58p6RavineOracle, "", 0, 0)
}

func s58p6Animate(t *testing.T, g *game.Game, me *game.Player, ravine uuid.UUID) {
	t.Helper()
	fillPool(me, 2)
	fillPoolColored(me, "R", 1)
	fillPoolColored(me, "G", 1)
	s58p6Activate(t, g, me.ID, ravine, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
}

func TestRagingRavineEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	id := b12PlayFromHand(t, g, "Raging Ravine", "Land", s58p6RavineOracle, game.CastSpellParams{})
	if !isTapped(g, id) {
		t.Error("Raging Ravine entered untapped")
	}
}

// The animation, layer by layer: a 3/3 red and green Elemental creature
// that is still a land, with the quoted trigger and its mana ability.
func TestRagingRavineBecomesAThreeThreeElementalLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ravine := s58p6Ravine(g, me.ID)
	advanceTo(t, g, game.StepPrecombatMain)
	if s58p6View(t, g, ravine).IsCreature() {
		t.Fatal("an idle Raging Ravine is a creature")
	}
	s58p6Animate(t, g, me, ravine)

	c := s58p6View(t, g, ravine)
	eff := c.Effective()
	if !c.IsCreature() || !c.IsLand() {
		t.Fatalf("creature=%v land=%v, want both", c.IsCreature(), c.IsLand())
	}
	if !slices.Contains(eff.Subtypes, "Elemental") {
		t.Errorf("subtypes %v, want Elemental", eff.Subtypes)
	}
	if !slices.Equal(c.EffectiveColors(), []string{"R", "G"}) {
		t.Errorf("colours %v, want red and green", c.EffectiveColors())
	}
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("%d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if len(eff.GrantedAbilities) != 1 || eff.GrantedAbilities[0].Key != game.GrantKey(ragingRavineAttackCounter) {
		t.Errorf("granted %v, want the attack trigger once", eff.GrantedAbilities)
	}
	if len(game.ManaAbilitiesForCard(c)) == 0 {
		t.Error("the animated land lost its mana ability")
	}
}

// Two activations, two triggers, two counters; the counters stay when
// the turn ends and the land stops being a creature, and count again
// the next time it animates.
func TestRagingRavineCountersStayAndStack(t *testing.T) {
	g := newCatalogGame(t)
	mySeat := g.Turn.ActiveSeat
	me := g.Seats[mySeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ravine := s58p6Ravine(g, me.ID)
	advanceTo(t, g, game.StepPrecombatMain)
	s58p6Animate(t, g, me, ravine)
	s58p6Animate(t, g, me, ravine)

	declareAttack(t, g, opp.ID, ravine)
	if n := triggersOnStackFrom(g, ravine); n != 2 {
		t.Fatalf("%d attack triggers, want 2", n)
	}
	passPriorityAroundTable(t, g)
	if n := plusOneCounters(g, ravine); n != 2 {
		t.Fatalf("%d counters, want 2", n)
	}

	advanceToNextSeatsTurn(t, g)
	c := s58p6View(t, g, ravine)
	if c.IsCreature() || !c.IsLand() {
		t.Fatalf("after the turn: creature=%v land=%v, want a land only", c.IsCreature(), c.IsLand())
	}
	if len(c.EffectiveColors()) != 0 {
		t.Errorf("after the turn the land is %v, want colourless", c.EffectiveColors())
	}
	if n := plusOneCounters(g, ravine); n != 2 {
		t.Errorf("counters %d after it stopped being a creature, want 2", n)
	}
	advanceToMainOf(t, g, mySeat)
	s58p6Animate(t, g, me, ravine)
	if got := s58p6View(t, g, ravine).CurrentPower(); got != 5 {
		t.Errorf("animated again with two counters: power %d, want 5", got)
	}
}

// CR 302.6: animated the turn it entered, it can't attack, and its {T}
// mana ability can't be activated while it is a creature.
func TestRagingRavineAnimatedTheTurnItEntersIsSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ravine := s58p6Ravine(g, me.ID)
	findBattlefieldCardByID(g, ravine).SummonedThisTurn = true
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateManaAbility(me.ID, ravine, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("a new land taps for mana: %v", err)
	}
	findBattlefieldCardByID(g, ravine).Tapped = false
	me.ManaPool.EmptyPool()
	s58p6Animate(t, g, me, ravine)
	if err := g.ActivateManaAbility(me.ID, ravine, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Error("a summoning-sick creature land tapped for mana")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ravine, opp.ID); err == nil {
		t.Error("a summoning-sick creature land attacked")
	}
}

// --- Saryth, the Viper's Fang ----------------------------------------

func TestSarythGivesTappedDeathtouchAndUntappedHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saryth := b12Push(g, me.ID, "Saryth, the Viper's Fang", "Legendary Creature — Human Warlock", s58p6SarythOracle, 3, 4)
	tapped := b12Creature(g, me.ID, "Tapped", "Creature — Bear", 2, 2)
	untapped := b12Creature(g, me.ID, "Untapped", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	findBattlefieldCardByID(g, tapped).Tapped = true
	g.BumpLayerVersionForTest()

	if c := s58p6View(t, g, tapped); !game.HasKeyword(&c, "deathtouch") || game.HasKeyword(&c, "hexproof") {
		t.Error("a tapped creature you control: want deathtouch, not hexproof")
	}
	if c := s58p6View(t, g, untapped); !game.HasKeyword(&c, "hexproof") || game.HasKeyword(&c, "deathtouch") {
		t.Error("an untapped creature you control: want hexproof, not deathtouch")
	}
	if c := s58p6View(t, g, saryth); game.HasKeyword(&c, "hexproof") || game.HasKeyword(&c, "deathtouch") {
		t.Error("Saryth gave itself a keyword")
	}
	if c := s58p6View(t, g, theirs); game.HasKeyword(&c, "hexproof") {
		t.Error("an opponent's creature got hexproof")
	}

	advanceTo(t, g, game.StepPrecombatMain)
	fillPool(me, 1)
	s58p6Activate(t, g, me.ID, saryth, game.ActivateAbilityParams{Targets: cardTarget(tapped)})
	passPriorityAroundTable(t, g)
	if isTapped(g, tapped) {
		t.Error("the untap ability did not untap the target")
	}
}

// --- Shadow in the Warp --------------------------------------------

func TestShadowInTheWarpDiscountsOnlyTheFirstCreatureSpell(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Shadow in the Warp", "Enchantment", s58p6ShadowOracle, false)
	first := handSpell(me, "First", "Creature — Bear", "{2}{G}")
	fillPoolColored(me, "G", 1)
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the first creature spell, {2} less: %v", err)
	}
	passPriorityAroundTable(t, g)
	second := handSpell(me, "Second", "Creature — Bear", "{2}{G}")
	fillPoolColored(me, "G", 1)
	refusedForMana(t, g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true}), "the second creature spell")
}

func TestShadowInTheWarpPingsAnOpponentsFirstNoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shadow := pushCatalogPermanent(g, me.ID, "Shadow in the Warp", "Enchantment", s58p6ShadowOracle, false)
	advanceToMainOf(t, g, 1)
	opp := g.Seats[1]
	life := opp.Life
	creature := handSpell(opp, "Their Bear", "Creature — Bear", "")
	if err := g.CastSpell(opp.ID, creature, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	for i := 0; i < 2; i++ {
		spell := handSpell(opp, "Their Sorcery", "Sorcery", "")
		if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
			t.Fatal(err)
		}
		if i == 0 && triggerOnStack(g, shadow) == nil {
			t.Fatal("no trigger on the first noncreature spell")
		}
		passPriorityAroundTable(t, g)
	}
	if got := life - opp.Life; got != 2 {
		t.Errorf("opponent lost %d, want 2 (the first noncreature spell only)", got)
	}
}

// --- Somberwald Sage ---------------------------------------------------

func TestSomberwaldSageManaCastsOnlyCreatureSpells(t *testing.T) {
	g, me, _ := spendTable(t)
	sage := pushCatalogPermanent(g, me.ID, "Somberwald Sage", "Creature — Human Druid", s58p6SomberwaldOracle, false)
	if err := g.ActivateManaAbility(me.ID, sage, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap the Sage: %v", err)
	}
	if got := poolColors(me); !slices.Equal(got, []string{"G", "G", "G"}) {
		t.Fatalf("pool = %v, want three green", got)
	}
	cost, _ := game.ParseCost("{G}")
	creature := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{G}"}
	sorcery := game.Card{Name: "Growth", TypeLine: "Sorcery", ManaCost: "{G}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(creature)) {
		t.Error("refused a creature spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(sorcery)) {
		t.Error("paid for a sorcery")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(creature)) {
		t.Error("paid for a creature's ability")
	}
}

// --- The Skullspore Nexus ----------------------------------------------

func TestSkullsporeNexusCostsLessByTheGreatestPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range []int{2, 6, 3} {
		b12Creature(g, me.ID, "Beater", "Creature — Beast", p, p)
	}
	spec, _ := Lookup(s58p6SkullsporeOracle)
	if got := spec.SelfCostModifiers[0].Amount(game.CostQuery{Game: g, Controller: me.ID}); got != 6 {
		t.Errorf("reduction = %d, want 6", got)
	}
}

// One wrath, one token: the total last-known power of your nontoken
// creatures that died (a +1/+1 counter included), not a token's, not an
// opponent's.
func TestSkullsporeNexusMakesOneTokenFromTheTotalPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "The Skullspore Nexus", "Legendary Artifact", s58p6SkullsporeOracle, false)
	a := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	findBattlefieldCardByID(g, a).Counters = map[string]int{game.CounterPlusOne: 1}
	b12Creature(g, me.ID, "Ox", "Creature — Ox", 4, 4)
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 9, 9)

	castCatalogSpell(t, g, "Damnation", "Sorcery", s58p6DamnationOracle, nil)
	passPriorityAroundTable(t, g)

	var tokens []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Fungus Dinosaur" {
			tokens = append(tokens, c)
		}
	}
	if len(tokens) != 1 {
		t.Fatalf("%d Fungus Dinosaurs, want 1", len(tokens))
	}
	c := s58p6View(t, g, tokens[0].InstanceID)
	if c.CurrentPower() != 7 || c.CurrentToughness() != 7 || c.Controller != me.ID {
		t.Errorf("token %d/%d controlled by me=%v, want a 7/7 of mine", c.CurrentPower(), c.CurrentToughness(), c.Controller == me.ID)
	}
	if !slices.Equal(c.EffectiveColors(), []string{"G"}) {
		t.Errorf("token colours %v, want green", c.EffectiveColors())
	}
}

func TestSkullsporeNexusDoublesTargetCreaturesPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nexus := pushCatalogPermanent(g, me.ID, "The Skullspore Nexus", "Legendary Artifact", s58p6SkullsporeOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	fillPool(me, 2)
	s58p6Activate(t, g, me.ID, nexus, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	passPriorityAroundTable(t, g)
	if c := s58p6View(t, g, bear); c.CurrentPower() != 6 || c.CurrentToughness() != 3 {
		t.Errorf("bear %d/%d, want 6/3", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Traverse the Outlands ------------------------------------------

func TestTraverseTheOutlandsFetchesUpToTheGreatestPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	var forests []uuid.UUID
	for i := 0; i < 3; i++ {
		forests = append(forests, stapleLibraryCard(me, "Forest", "Basic Land — Forest"))
	}
	stapleLibraryCard(me, "Stomping Ground", "Land — Mountain Forest")

	castCatalogSpell(t, g, "Traverse the Outlands", "Sorcery", s58p6TraverseOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Stomping Ground") != uuid.Nil {
		t.Error("a nonbasic land was offered")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, forests); err == nil {
		t.Error("took three basics with a greatest power of two")
	}
	answerSearchByID(t, g, me.ID, forests[0], forests[1])
	for _, id := range forests[:2] {
		got, ok := battlefieldCard(g, id)
		if !ok || !got.Tapped {
			t.Errorf("fetched Forest on battlefield=%v tapped=%v, want both", ok, got.Tapped)
		}
	}
}

func TestTraverseTheOutlandsWithNoCreatureFindsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Traverse the Outlands", "Sorcery", s58p6TraverseOracle, nil)
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		answerSearchByID(t, g, me.ID)
	}
	if g.Battlefield.Contains(forest) {
		t.Error("X = 0 put a land onto the battlefield")
	}
}

// --- Zopandrel, Hunger Dominus ------------------------------------------

func TestZopandrelDoublesYourCreaturesAtEachCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	zop := b12Push(g, me.ID, "Zopandrel, Hunger Dominus", "Legendary Creature — Phyrexian Horror", s58p6ZopandrelOracle, 4, 6)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if c := s58p6View(t, g, bear); c.CurrentPower() != 4 || c.CurrentToughness() != 6 {
		t.Errorf("bear %d/%d, want 4/6", c.CurrentPower(), c.CurrentToughness())
	}
	if c := s58p6View(t, g, zop); c.CurrentPower() != 8 || c.CurrentToughness() != 12 {
		t.Errorf("Zopandrel %d/%d, want 8/12", c.CurrentPower(), c.CurrentToughness())
	}
	if c := s58p6View(t, g, theirs); c.CurrentPower() != 2 {
		t.Errorf("an opponent's creature was doubled to %d", c.CurrentPower())
	}

	advanceToStepOf(t, g, 1, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if c := s58p6View(t, g, bear); c.CurrentPower() != 4 {
		t.Errorf("on the opponent's turn the bear is %d, want 4 (each combat)", c.CurrentPower())
	}
}

func TestZopandrelSacrificesTwoOthersForAnIndestructibleCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	zop := b12Push(g, me.ID, "Zopandrel, Hunger Dominus", "Legendary Creature — Phyrexian Horror", s58p6ZopandrelOracle, 4, 6)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	advanceTo(t, g, game.StepPrecombatMain)
	fillPoolColored(me, "G", 2)
	if err := g.ActivateCatalogAbility(me.ID, zop, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{zop, a}}); err == nil {
		t.Fatal("Zopandrel sacrificed itself to its own ability")
	}
	s58p6Activate(t, g, me.ID, zop, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, b}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the two creatures were not sacrificed")
	}
	c := s58p6View(t, g, zop)
	if c.Counters[game.CounterIndestructible] != 1 || !game.HasKeyword(&c, "indestructible") {
		t.Errorf("counters %v, want one indestructible counter", c.Counters)
	}
}

// --- Growing Rites of Itlimoc ------------------------------------------

func s58p6RitesRow() cards.Card {
	return transformRow(growingRitesOfItlimocOracleID,
		"Growing Rites of Itlimoc", "Legendary Enchantment", "{2}{G}",
		"Itlimoc, Cradle of the Sun", "Legendary Land", "", "", []string{"G"})
}

// s58p6SettleLook settles the enters trigger's look at four: takes
// `take` (or nothing) and puts the rest on the bottom in the order
// offered.
func s58p6SettleLook(t *testing.T, g *game.Game, me *game.Player, take ...uuid.UUID) {
	t.Helper()
	passPriorityAroundTable(t, g)
	if c := chooseCardsChoiceFor(g, me.ID); c != nil {
		answerChooseCards(t, g, me.ID, take...)
	}
	if order := putInLibraryChoiceFor(g, me.ID); order != nil {
		if err := g.ResolvePutInLibrary(order.ID, me.ID, order.ScryCards, nil); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
}

// CR 603.4: with three creatures the ability doesn't trigger; with four
// it transforms into a land that taps for {G} per creature.
func TestGrowingRitesTransformsWithFourCreaturesAtYourEndStep(t *testing.T) {
	for _, n := range []int{3, 4} {
		g := newCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me := g.Seats[seat]
		rites := importToBattlefield(t, g, s58p6RitesRow(), me)
		s58p6SettleLook(t, g, me)
		for i := 0; i < n; i++ {
			b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
		}
		advanceToEndStepOf(t, g, seat)
		if (triggerOnStack(g, rites) != nil) != (n == 4) {
			t.Errorf("%d creatures: triggered=%v", n, triggerOnStack(g, rites) != nil)
		}
		passPriorityAroundTable(t, g)
		c := s58p6View(t, g, rites)
		if transformed := c.ActiveFace == 1 && c.IsLand(); transformed != (n == 4) {
			t.Errorf("%d creatures: transformed=%v", n, transformed)
		}
		if n != 4 {
			continue
		}
		if c.Name != "Itlimoc, Cradle of the Sun" {
			t.Errorf("name %q", c.Name)
		}
		me.ManaPool.EmptyPool()
		if err := g.ActivateManaAbility(me.ID, rites, 1, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("tap Itlimoc for {G} per creature: %v", err)
		}
		if got := poolColors(me); !slices.Equal(got, []string{"G", "G", "G", "G"}) {
			t.Errorf("pool %v, want four green", got)
		}
	}
}

func TestGrowingRitesLooksAtFourAndTakesACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var top []uuid.UUID
	for _, tl := range []string{"Creature — Deep", "Creature — Elf", "Land", "Instant", "Creature — Ox"} {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: tl, TypeLine: tl, Owner: me.ID, Controller: me.ID})
		top = append([]uuid.UUID{id}, top...)
	}
	castCatalogSpell(t, g, "Growing Rites of Itlimoc", "Legendary Enchantment", s58p6GrowingRiteOracle, nil)
	passPriorityAroundTable(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no look-at-four prompt")
	}
	if !slices.Contains(c.ChooseCards, top[0]) || !slices.Contains(c.ChooseCards, top[3]) ||
		slices.Contains(c.ChooseCards, top[1]) || slices.Contains(c.ChooseCards, top[4]) {
		t.Errorf("offered %v, want the two creature cards among the top four only", c.ChooseCards)
	}
	elf := top[3]
	s58p6SettleLook(t, g, me, elf)
	if !me.Hand.Contains(elf) {
		t.Error("the creature card did not go to hand")
	}
	if n := len(me.Library.Cards); n < 4 || me.Library.Cards[0].InstanceID == top[4] {
		t.Error("the rest did not go to the bottom")
	}
}
