package testsru

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/r"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Unleash Fury ------------------------------------------------------

// CR 701.10b: +X/+0 with X the power as it resolves, counters included.
// A pump afterwards is not doubled again.
func TestUnleashFuryDoublesThePowerItReads(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FindBattlefieldCardByID(g, bear).Counters = map[string]int{game.CounterPlusOne: 1}

	effectstest.CastCatalogSpell(t, g, "Unleash Fury", "Instant", effectstest.S58p6UnleashFuryOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.S58p6View(t, g, bear)
	if c.CurrentPower() != 6 || c.CurrentToughness() != 3 {
		t.Fatalf("doubled 3/3 is %d/%d, want 6/3", c.CurrentPower(), c.CurrentToughness())
	}
	g.WithWriteLock(func() {
		_ = effects.BoostUntilEOT{Target: bear, Power: 3, Toughness: 3}.Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	if got := effectstest.S58p6View(t, g, bear).CurrentPower(); got != 9 {
		t.Errorf("a later +3 gives %d, want 9 (doubling is a one-time +3)", got)
	}
}

// --- Xenagos, God of Revels ------------------------------------------

// CR 700.5: devotion to red and green counts a {R/G} hybrid once.
func TestXenagosCountsAHybridSymbolOnceTowardSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	xen := effectstest.S58p6Push(g, me.ID, "Xenagos, God of Revels", "Legendary Enchantment Creature — God", effectstest.S58p6XenagosOracle, "{3}{R}{G}", 6, 5)
	effectstest.S58p6Push(g, me.ID, "Red Thing", "Enchantment", "", "{R}{R}{G}", 0, 0)
	effectstest.S58p6Push(g, me.ID, "Hybrid Thing", "Enchantment", "", "{R/G}", 0, 0)
	effectstest.S58p6Push(g, g.Seats[1].ID, "Their Green", "Enchantment", "", "{G}{G}{G}", 0, 0)

	if c := effectstest.S58p6View(t, g, xen); c.IsCreature() {
		t.Fatal("devotion 6 (the hybrid once, the opponent's not at all): Xenagos is a creature")
	}
	if !game.HasKeyword(ptrTo(effectstest.S58p6View(t, g, xen)), "indestructible") {
		t.Error("Xenagos is indestructible whether or not it is a creature")
	}
	effectstest.S58p6Push(g, me.ID, "Green Thing", "Enchantment", "", "{G}", 0, 0)
	if c := effectstest.S58p6View(t, g, xen); !c.IsCreature() || c.CurrentPower() != 6 {
		t.Errorf("devotion 7: creature=%v power=%d, want a 6/5 creature", c.IsCreature(), c.CurrentPower())
	}
}

func ptrTo(c game.Card) *game.Card { return &c }

// The combat trigger targets ANOTHER creature you control: a summoning-
// sick 2/3 becomes a hasty 4/5 and may attack; Xenagos is never offered.
func TestXenagosGivesAnotherCreatureHasteAndItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	xen := effectstest.S58p6Push(g, me.ID, "Xenagos, God of Revels", "Legendary Enchantment Creature — God", effectstest.S58p6XenagosOracle, "{3}{R}{G}", 6, 5)
	effectstest.S58p6Push(g, me.ID, "Devotion", "Enchantment", "", "{R}{R}{G}{G}{G}", 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 3)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	effectstest.FindBattlefieldCardByID(g, bear).SummonedThisTurn = true
	if !effectstest.S58p6View(t, g, xen).IsCreature() {
		t.Fatal("setup: Xenagos should be a creature at devotion 7")
	}

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	c := effectstest.PickTargetPrompt(t, g)
	if slices.Contains(c.PickTargetCards, xen) {
		t.Error("Xenagos was offered as its own target")
	}
	if !slices.Contains(c.PickTargetCards, other) {
		t.Error("another creature you control was not offered")
	}
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)

	got := effectstest.S58p6View(t, g, bear)
	if got.CurrentPower() != 4 || got.CurrentToughness() != 5 {
		t.Errorf("bear is %d/%d, want 4/5", got.CurrentPower(), got.CurrentToughness())
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Errorf("the hasty bear could not attack: %v", err)
	}
}

// --- Anzrag, the Quake-Mole ------------------------------------------

func TestAnzragBlockedUntapsYourCreaturesAndAddsACombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	anzrag := effectstest.B12Push(g, me.ID, "Anzrag, the Quake-Mole", "Legendary Creature — Mole God", effectstest.S58p6AnzragOracle, 8, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	w1 := effectstest.B12Creature(g, opp.ID, "Wall One", "Creature — Wall", 0, 9)
	w2 := effectstest.B12Creature(g, opp.ID, "Wall Two", "Creature — Wall", 0, 9)
	start := len(g.Events)

	effectstest.DeclareAttack(t, g, opp.ID, anzrag, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, w := range []uuid.UUID{w1, w2} {
		if err := g.DeclareBlocker(w, anzrag); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	if n := effectstest.TriggersOnStackFrom(g, anzrag); n != 1 {
		t.Fatalf("two blockers: %d triggers, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, anzrag) || effectstest.IsTapped(g, bear) {
		t.Error("the trigger did not untap the creatures you control")
	}
	const label = "Anzrag, the Quake-Mole — untap each creature you control, additional combat"
	if n := effectstest.CountResolved(g, start, label); n != 1 {
		t.Errorf("resolved %d times, want 1", n)
	}
	for i := 0; i < 8 && g.Turn.Step != game.StepEndCombat; i++ {
		effectstest.B26AnswerDamageAssignments(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.StepOnce(t, g)
	}
	if g.Turn.Step != game.StepEndCombat {
		t.Fatalf("never reached end of combat (at %s)", g.Turn.Step)
	}
	if next := effectstest.StepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Errorf("after combat: %+v, want an additional combat", next)
	}
}

func TestAnzragMustBeBlockedOnceActivated(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	anzrag := effectstest.B12Push(g, me.ID, "Anzrag, the Quake-Mole", "Legendary Creature — Mole God", effectstest.S58p6AnzragOracle, 8, 4)
	wall := effectstest.ReqCreature(g, opp.ID, "Wall", "", 0, 9)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "R", 2)
	effectstest.FillPoolColored(me, "G", 2)
	effectstest.S58p6Activate(t, g, me.ID, anzrag, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.ReqToBlockers(t, g, anzrag)
	effectstest.ReqRefusal(t, g.PassPriority())
	if err := effectstest.ReqBlock(t, g, anzrag, wall); err != nil {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Berserk", "Instant", effectstest.S58p6BerserkOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.S58p6View(t, g, bear)
	if c.CurrentPower() != 6 || c.CurrentToughness() != 3 || !game.HasKeyword(&c, "trample") {
		t.Fatalf("bear is %d/%d trample=%v, want 6/3 with trample", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "trample"))
	}
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 6 {
		t.Errorf("combat damage %d, want 6", got)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the creature attacked and survived the end step")
	}
}

// A creature that did not attack is not destroyed.
func TestBerserkSparesACreatureThatDidNotAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Berserk", "Instant", effectstest.S58p6BerserkOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature that never attacked was destroyed")
	}
}

// The cast window: precombat main and the combat steps before damage
// yes; once combat damage has begun, and in the postcombat main phase,
// no.
func TestBerserkCastOnlyBeforeTheCombatDamageStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := effectstest.S58p6CastNow(g, me, "Berserk", "Instant", effectstest.S58p6BerserkOracle, effectstest.CardTarget(bear)); err != nil {
		t.Fatalf("beginning of combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AttackWith(t, g, opp.ID, bear)
	if err := effectstest.S58p6CastNow(g, me, "Berserk", "Instant", effectstest.S58p6BerserkOracle, effectstest.CardTarget(bear)); err == nil {
		t.Error("cast during the combat damage step")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if err := effectstest.S58p6CastNow(g, me, "Berserk", "Instant", effectstest.S58p6BerserkOracle, effectstest.CardTarget(bear)); err == nil {
		t.Error("cast in the postcombat main phase")
	}
}

// --- Gwenna, Eyes of Gaea --------------------------------------------

func TestGwennaManaPaysCreatureSpellsAndCreatureAbilitiesOnly(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	gwenna := effectstest.PushCatalogPermanent(g, me.ID, "Gwenna, Eyes of Gaea", "Legendary Creature — Elf Druid Scout", effectstest.S58p6GwennaOracle, false)
	if err := g.ActivateManaAbility(me.ID, gwenna, 0, game.ManaAbilityParams{Colors: []string{"R", "U"}}); err != nil {
		t.Fatalf("tap Gwenna: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 2 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gwenna := effectstest.B12Push(g, me.ID, "Gwenna, Eyes of Gaea", "Legendary Creature — Elf Druid Scout", effectstest.S58p6GwennaOracle, 2, 3)
	effectstest.FindBattlefieldCardByID(g, gwenna).Tapped = true
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)

	small := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: small, Name: "Four", TypeLine: "Creature — Beast", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, small, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOneCounters(g, gwenna) > 0 || !effectstest.IsTapped(g, gwenna) {
		t.Fatal("a power-4 creature spell triggered Gwenna")
	}
	big := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: big, Name: "Five", TypeLine: "Creature — Beast", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOneCounters(g, gwenna) != 1 || effectstest.IsTapped(g, gwenna) {
		t.Errorf("counters=%d tapped=%v, want one counter and untapped", effectstest.PlusOneCounters(g, gwenna), effectstest.IsTapped(g, gwenna))
	}
}

// --- Last March of the Ents ------------------------------------------

func TestLastMarchDrawsTheGreatestToughnessThenPutsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 3)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 5; i++ {
		effectstest.StapleLibraryCard(me, "Spare", "Sorcery")
	}
	elk := effectstest.HandCardForTest(me, "Elk", "Creature — Elk", "")
	ox := effectstest.HandCardForTest(me, "Ox", "Creature — Ox", "")
	effectstest.HandCardForTest(me, "Bolt", "Instant", "")
	if spec, ok := effects.Lookup(effectstest.S58p6LastMarchOracle); !ok || !spec.CantBeCountered {
		t.Error("Last March of the Ents must not be counterable")
	}
	effectstest.CastCatalogSpell(t, g, "Last March of the Ents", "Sorcery", effectstest.S58p6LastMarchOracle, nil)
	before := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 3 {
		t.Fatalf("drew %d, want 3 (the greatest toughness)", got)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, elk, ox)
	if !g.Battlefield.Contains(elk) || !g.Battlefield.Contains(ox) {
		t.Error("the chosen creature cards are not on the battlefield")
	}
}

// --- Mage Slayer -------------------------------------------------------

// The equipped creature deals damage equal to its power to the player
// it attacks, as the trigger resolves, before combat damage.
func TestMageSlayerHitsTheDefenderForThePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	slayer := effectstest.B12Push(g, me.ID, "Mage Slayer", "Artifact — Equipment", effectstest.S58p6MageSlayerOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FillPool(me, 3)
	effectstest.S58p6Activate(t, g, me.ID, slayer, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)

	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if effectstest.TriggerOnStack(g, slayer) == nil {
		t.Fatal("no Mage Slayer trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("trigger dealt %d, want 3", got)
	}
}

// --- Raging Ravine -------------------------------------------------------

func s58p6Ravine(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.S58p6Push(g, owner, "Raging Ravine", "Land", effectstest.S58p6RavineOracle, "", 0, 0)
}

func s58p6Animate(t *testing.T, g *game.Game, me *game.Player, ravine uuid.UUID) {
	t.Helper()
	effectstest.FillPool(me, 2)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.S58p6Activate(t, g, me.ID, ravine, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestRagingRavineEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12PlayFromHand(t, g, "Raging Ravine", "Land", effectstest.S58p6RavineOracle, game.CastSpellParams{})
	if !effectstest.IsTapped(g, id) {
		t.Error("Raging Ravine entered untapped")
	}
}

// The animation, layer by layer: a 3/3 red and green Elemental creature
// that is still a land, with the quoted trigger and its mana ability.
func TestRagingRavineBecomesAThreeThreeElementalLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ravine := s58p6Ravine(g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if effectstest.S58p6View(t, g, ravine).IsCreature() {
		t.Fatal("an idle Raging Ravine is a creature")
	}
	s58p6Animate(t, g, me, ravine)

	c := effectstest.S58p6View(t, g, ravine)
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
	if len(eff.GrantedAbilities) != 1 || eff.GrantedAbilities[0].Key != game.GrantKey(cardsr.RagingRavineAttackCounter) {
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
	g := effectstest.NewCatalogGame(t)
	mySeat := g.Turn.ActiveSeat
	me := g.Seats[mySeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ravine := s58p6Ravine(g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	s58p6Animate(t, g, me, ravine)
	s58p6Animate(t, g, me, ravine)

	effectstest.DeclareAttack(t, g, opp.ID, ravine)
	if n := effectstest.TriggersOnStackFrom(g, ravine); n != 2 {
		t.Fatalf("%d attack triggers, want 2", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, ravine); n != 2 {
		t.Fatalf("%d counters, want 2", n)
	}

	effectstest.AdvanceToNextSeatsTurn(t, g)
	c := effectstest.S58p6View(t, g, ravine)
	if c.IsCreature() || !c.IsLand() {
		t.Fatalf("after the turn: creature=%v land=%v, want a land only", c.IsCreature(), c.IsLand())
	}
	if len(c.EffectiveColors()) != 0 {
		t.Errorf("after the turn the land is %v, want colourless", c.EffectiveColors())
	}
	if n := effectstest.PlusOneCounters(g, ravine); n != 2 {
		t.Errorf("counters %d after it stopped being a creature, want 2", n)
	}
	effectstest.AdvanceToMainOf(t, g, mySeat)
	s58p6Animate(t, g, me, ravine)
	if got := effectstest.S58p6View(t, g, ravine).CurrentPower(); got != 5 {
		t.Errorf("animated again with two counters: power %d, want 5", got)
	}
}

// CR 302.6: animated the turn it entered, it can't attack, and its {T}
// mana ability can't be activated while it is a creature.
func TestRagingRavineAnimatedTheTurnItEntersIsSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ravine := s58p6Ravine(g, me.ID)
	effectstest.FindBattlefieldCardByID(g, ravine).SummonedThisTurn = true
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateManaAbility(me.ID, ravine, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("a new land taps for mana: %v", err)
	}
	effectstest.FindBattlefieldCardByID(g, ravine).Tapped = false
	me.ManaPool.EmptyPool()
	s58p6Animate(t, g, me, ravine)
	if err := g.ActivateManaAbility(me.ID, ravine, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Error("a summoning-sick creature land tapped for mana")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ravine, opp.ID); err == nil {
		t.Error("a summoning-sick creature land attacked")
	}
}

// --- Saryth, the Viper's Fang ----------------------------------------

func TestSarythGivesTappedDeathtouchAndUntappedHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saryth := effectstest.B12Push(g, me.ID, "Saryth, the Viper's Fang", "Legendary Creature — Human Warlock", effectstest.S58p6SarythOracle, 3, 4)
	tapped := effectstest.B12Creature(g, me.ID, "Tapped", "Creature — Bear", 2, 2)
	untapped := effectstest.B12Creature(g, me.ID, "Untapped", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.FindBattlefieldCardByID(g, tapped).Tapped = true
	g.BumpLayerVersionForTest()

	if c := effectstest.S58p6View(t, g, tapped); !game.HasKeyword(&c, "deathtouch") || game.HasKeyword(&c, "hexproof") {
		t.Error("a tapped creature you control: want deathtouch, not hexproof")
	}
	if c := effectstest.S58p6View(t, g, untapped); !game.HasKeyword(&c, "hexproof") || game.HasKeyword(&c, "deathtouch") {
		t.Error("an untapped creature you control: want hexproof, not deathtouch")
	}
	if c := effectstest.S58p6View(t, g, saryth); game.HasKeyword(&c, "hexproof") || game.HasKeyword(&c, "deathtouch") {
		t.Error("Saryth gave itself a keyword")
	}
	if c := effectstest.S58p6View(t, g, theirs); game.HasKeyword(&c, "hexproof") {
		t.Error("an opponent's creature got hexproof")
	}

	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FillPool(me, 1)
	effectstest.S58p6Activate(t, g, me.ID, saryth, game.ActivateAbilityParams{Targets: effectstest.CardTarget(tapped)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, tapped) {
		t.Error("the untap ability did not untap the target")
	}
}

// --- Shadow in the Warp --------------------------------------------

func TestShadowInTheWarpDiscountsOnlyTheFirstCreatureSpell(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Shadow in the Warp", "Enchantment", effectstest.S58p6ShadowOracle, false)
	first := effectstest.HandSpell(me, "First", "Creature — Bear", "{2}{G}")
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the first creature spell, {2} less: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	second := effectstest.HandSpell(me, "Second", "Creature — Bear", "{2}{G}")
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true}), "the second creature spell")
}

func TestShadowInTheWarpPingsAnOpponentsFirstNoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shadow := effectstest.PushCatalogPermanent(g, me.ID, "Shadow in the Warp", "Enchantment", effectstest.S58p6ShadowOracle, false)
	effectstest.AdvanceToMainOf(t, g, 1)
	opp := g.Seats[1]
	life := opp.Life
	creature := effectstest.HandSpell(opp, "Their Bear", "Creature — Bear", "")
	if err := g.CastSpell(opp.ID, creature, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 2; i++ {
		spell := effectstest.HandSpell(opp, "Their Sorcery", "Sorcery", "")
		if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
			t.Fatal(err)
		}
		if i == 0 && effectstest.TriggerOnStack(g, shadow) == nil {
			t.Fatal("no trigger on the first noncreature spell")
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := life - opp.Life; got != 2 {
		t.Errorf("opponent lost %d, want 2 (the first noncreature spell only)", got)
	}
}

// --- Somberwald Sage ---------------------------------------------------

func TestSomberwaldSageManaCastsOnlyCreatureSpells(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	sage := effectstest.PushCatalogPermanent(g, me.ID, "Somberwald Sage", "Creature — Human Druid", effectstest.S58p6SomberwaldOracle, false)
	if err := g.ActivateManaAbility(me.ID, sage, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap the Sage: %v", err)
	}
	if got := effectstest.PoolColors(me); !slices.Equal(got, []string{"G", "G", "G"}) {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range []int{2, 6, 3} {
		effectstest.B12Creature(g, me.ID, "Beater", "Creature — Beast", p, p)
	}
	spec, _ := effects.Lookup(effectstest.S58p6SkullsporeOracle)
	if got := spec.SelfCostModifiers[0].Amount(game.CostQuery{Game: g, Controller: me.ID}); got != 6 {
		t.Errorf("reduction = %d, want 6", got)
	}
}

// One wrath, one token: the total last-known power of your nontoken
// creatures that died (a +1/+1 counter included), not a token's, not an
// opponent's.
func TestSkullsporeNexusMakesOneTokenFromTheTotalPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "The Skullspore Nexus", "Legendary Artifact", effectstest.S58p6SkullsporeOracle, false)
	a := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FindBattlefieldCardByID(g, a).Counters = map[string]int{game.CounterPlusOne: 1}
	effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 4, 4)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 9, 9)

	effectstest.CastCatalogSpell(t, g, "Damnation", "Sorcery", effectstest.S58p6DamnationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	var tokens []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Fungus Dinosaur" {
			tokens = append(tokens, c)
		}
	}
	if len(tokens) != 1 {
		t.Fatalf("%d Fungus Dinosaurs, want 1", len(tokens))
	}
	c := effectstest.S58p6View(t, g, tokens[0].InstanceID)
	if c.CurrentPower() != 7 || c.CurrentToughness() != 7 || c.Controller != me.ID {
		t.Errorf("token %d/%d controlled by me=%v, want a 7/7 of mine", c.CurrentPower(), c.CurrentToughness(), c.Controller == me.ID)
	}
	if !slices.Equal(c.EffectiveColors(), []string{"G"}) {
		t.Errorf("token colours %v, want green", c.EffectiveColors())
	}
}

func TestSkullsporeNexusDoublesTargetCreaturesPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nexus := effectstest.PushCatalogPermanent(g, me.ID, "The Skullspore Nexus", "Legendary Artifact", effectstest.S58p6SkullsporeOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FillPool(me, 2)
	effectstest.S58p6Activate(t, g, me.ID, nexus, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.S58p6View(t, g, bear); c.CurrentPower() != 6 || c.CurrentToughness() != 3 {
		t.Errorf("bear %d/%d, want 6/3", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Traverse the Outlands ------------------------------------------

func TestTraverseTheOutlandsFetchesUpToTheGreatestPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	var forests []uuid.UUID
	for i := 0; i < 3; i++ {
		forests = append(forests, effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest"))
	}
	effectstest.StapleLibraryCard(me, "Stomping Ground", "Land — Mountain Forest")

	effectstest.CastCatalogSpell(t, g, "Traverse the Outlands", "Sorcery", effectstest.S58p6TraverseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Stomping Ground") != uuid.Nil {
		t.Error("a nonbasic land was offered")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, forests); err == nil {
		t.Error("took three basics with a greatest power of two")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, forests[0], forests[1])
	for _, id := range forests[:2] {
		got, ok := effectstest.BattlefieldCard(g, id)
		if !ok || !got.Tapped {
			t.Errorf("fetched Forest on battlefield=%v tapped=%v, want both", ok, got.Tapped)
		}
	}
}

func TestTraverseTheOutlandsWithNoCreatureFindsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Traverse the Outlands", "Sorcery", effectstest.S58p6TraverseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		effectstest.AnswerSearchByID(t, g, me.ID)
	}
	if g.Battlefield.Contains(forest) {
		t.Error("X = 0 put a land onto the battlefield")
	}
}

// --- Zopandrel, Hunger Dominus ------------------------------------------

func TestZopandrelDoublesYourCreaturesAtEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	zop := effectstest.B12Push(g, me.ID, "Zopandrel, Hunger Dominus", "Legendary Creature — Phyrexian Horror", effectstest.S58p6ZopandrelOracle, 4, 6)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	effectstest.AdvanceToStepOf(t, g, 0, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.S58p6View(t, g, bear); c.CurrentPower() != 4 || c.CurrentToughness() != 6 {
		t.Errorf("bear %d/%d, want 4/6", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.S58p6View(t, g, zop); c.CurrentPower() != 8 || c.CurrentToughness() != 12 {
		t.Errorf("Zopandrel %d/%d, want 8/12", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.S58p6View(t, g, theirs); c.CurrentPower() != 2 {
		t.Errorf("an opponent's creature was doubled to %d", c.CurrentPower())
	}

	effectstest.AdvanceToStepOf(t, g, 1, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.S58p6View(t, g, bear); c.CurrentPower() != 4 {
		t.Errorf("on the opponent's turn the bear is %d, want 4 (each combat)", c.CurrentPower())
	}
}

func TestZopandrelSacrificesTwoOthersForAnIndestructibleCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	zop := effectstest.B12Push(g, me.ID, "Zopandrel, Hunger Dominus", "Legendary Creature — Phyrexian Horror", effectstest.S58p6ZopandrelOracle, 4, 6)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FillPoolColored(me, "G", 2)
	if err := g.ActivateCatalogAbility(me.ID, zop, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{zop, a}}); err == nil {
		t.Fatal("Zopandrel sacrificed itself to its own ability")
	}
	effectstest.S58p6Activate(t, g, me.ID, zop, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, b}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the two creatures were not sacrificed")
	}
	c := effectstest.S58p6View(t, g, zop)
	if c.Counters[game.CounterIndestructible] != 1 || !game.HasKeyword(&c, "indestructible") {
		t.Errorf("counters %v, want one indestructible counter", c.Counters)
	}
}

// --- Growing Rites of Itlimoc ------------------------------------------

func s58p6RitesRow() cards.Card {
	return effectstest.TransformRow(effects.GrowingRitesOfItlimocOracleID,
		"Growing Rites of Itlimoc", "Legendary Enchantment", "{2}{G}",
		"Itlimoc, Cradle of the Sun", "Legendary Land", "", "", []string{"G"})
}

// s58p6SettleLook settles the enters trigger's look at four: takes
// `take` (or nothing) and puts the rest on the bottom in the order
// offered.
func s58p6SettleLook(t *testing.T, g *game.Game, me *game.Player, take ...uuid.UUID) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c != nil {
		effectstest.AnswerChooseCards(t, g, me.ID, take...)
	}
	if order := effectstest.PutInLibraryChoiceFor(g, me.ID); order != nil {
		if err := g.ResolvePutInLibrary(order.ID, me.ID, order.ScryCards, nil); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// CR 603.4: with three creatures the ability doesn't trigger; with four
// it transforms into a land that taps for {G} per creature.
func TestGrowingRitesTransformsWithFourCreaturesAtYourEndStep(t *testing.T) {
	for _, n := range []int{3, 4} {
		g := effectstest.NewCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me := g.Seats[seat]
		rites := effectstest.ImportToBattlefield(t, g, s58p6RitesRow(), me)
		s58p6SettleLook(t, g, me)
		for i := 0; i < n; i++ {
			effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
		}
		effectstest.AdvanceToEndStepOf(t, g, seat)
		if (effectstest.TriggerOnStack(g, rites) != nil) != (n == 4) {
			t.Errorf("%d creatures: triggered=%v", n, effectstest.TriggerOnStack(g, rites) != nil)
		}
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.S58p6View(t, g, rites)
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
		if got := effectstest.PoolColors(me); !slices.Equal(got, []string{"G", "G", "G", "G"}) {
			t.Errorf("pool %v, want four green", got)
		}
	}
}

func TestGrowingRitesLooksAtFourAndTakesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var top []uuid.UUID
	for _, tl := range []string{"Creature — Deep", "Creature — Elf", "Land", "Instant", "Creature — Ox"} {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: tl, TypeLine: tl, Owner: me.ID, Controller: me.ID})
		top = append([]uuid.UUID{id}, top...)
	}
	effectstest.CastCatalogSpell(t, g, "Growing Rites of Itlimoc", "Legendary Enchantment", effectstest.S58p6GrowingRiteOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
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
