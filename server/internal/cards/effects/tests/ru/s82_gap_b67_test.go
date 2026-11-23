package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b67_test.go — Fallout enchantments and instants (GAP batch 67,
// #5948), played through the real catalog entries.

const (
	gapB67WildWasteland = "4297ab35-9f16-4d9e-af77-010469ec4f59"
	gapB67NerdRage      = "4601e4c7-74bc-44ce-9274-68ba79f8aae1"
	gapB67HooverDam     = "91596da5-5b1a-430c-a878-7757ae366b6b"
	gapB67Idolized      = "8cbf0173-2349-4205-b720-378643c347a4"
	gapB67Overencumber  = "54400ec3-f2ab-4cdf-9731-e04281bf7e90"
	gapB67Vault101      = "77322d69-330e-4b00-b984-62c6d45c9540"
	gapB67BottleCap     = "e9c78601-4bec-4d16-8fd9-88c6053bd9c4"
	gapB67Inventory     = "b20eae3d-d0ff-42c3-a060-d06629506781"
)

func TestGapB67CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB67WildWasteland: "Wild Wasteland", gapB67NerdRage: "Nerd Rage", gapB67HooverDam: "Battle of Hoover Dam",
		gapB67Idolized: "Idolized", gapB67Overencumber: "Overencumbered", gapB67Vault101: "Vault 101: Birthday Party",
		gapB67BottleCap: "Bottle-Cap Blast", gapB67Inventory: "Inventory Management",
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

func gapB67Aura(t *testing.T, g *game.Game, name, oracle string, target game.TargetRef) uuid.UUID {
	t.Helper()
	aura := effectstest.CastCatalogSpell(t, g, name, "Enchantment — Aura", oracle, []game.TargetRef{target})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatalf("%s did not enter", name)
	}
	return aura
}

func gapB67MaxHandSize(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = g.EffectiveMaxHandSizeLocked(p) })
	return n
}

// --- Wild Wasteland -------------------------------------------------------

func TestGapB67WildWastelandSkipsTheDrawAndExilesTwoEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.SeedLibrary(opp, "A", "B", "C", "D", "E")
	effectstest.PushCatalogPermanent(g, opp.ID, "Wild Wasteland", "Enchantment", gapB67WildWasteland, false)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	exiled := g.Exile.Size()
	hand := opp.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := g.Exile.Size() - exiled; got != 2 {
		t.Fatalf("%d cards exiled at upkeep, want 2", got)
	}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if got := opp.Hand.Size(); got != hand {
		t.Errorf("hand %d at the main phase, want %d (the draw step is skipped)", got, hand)
	}
}

// --- Nerd Rage ------------------------------------------------------------

func TestGapB67NerdRageDrawsTwoAndRemovesTheEnchantedCreaturesControllersHandLimit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	if got := gapB67MaxHandSize(g, opp); got != 7 {
		t.Fatalf("a baseline maximum hand size of %d, want 7", got)
	}
	gapB67Aura(t, g, "Nerd Rage", gapB67NerdRage, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards on entry, want 2", got)
	}
	if got := gapB67MaxHandSize(g, opp); got != game.NoMaxHandSize {
		t.Errorf("the enchanted creature's controller has a maximum hand size of %d, want none", got)
	}
	if got := gapB67MaxHandSize(g, me); got != 7 {
		t.Errorf("the Aura's controller has a maximum hand size of %d, want 7", got)
	}
}

func TestGapB67NerdRageGivesPlusTenOnlyWithTenCardsInHand(t *testing.T) {
	for _, tc := range []struct {
		cards, want int
	}{{9, 2}, {10, 12}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		gapB67Aura(t, g, "Nerd Rage", gapB67NerdRage, game.TargetRef{Kind: game.TargetCard, ID: bear})
		g.WithWriteLock(func() { me.Hand.Cards = nil })
		for i := 0; i < tc.cards; i++ {
			effectstest.PushCatalogHandCard(me, "Filler", "Sorcery", "")
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(bear, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if p, _ := effectstest.PtOf(t, g, bear); p != tc.want {
			t.Errorf("with %d cards in hand the creature has power %d, want %d", tc.cards, p, tc.want)
		}
	}
}

// --- Idolized -------------------------------------------------------------

func TestGapB67IdolizedPumpsOnlyAnAttackerAttackingAlone(t *testing.T) {
	for _, alone := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		friend := effectstest.B12Push(g, me.ID, "Friend", "Creature — Bear", "", 1, 1)
		effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
		effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
		gapB67Aura(t, g, "Idolized", gapB67Idolized, game.TargetRef{Kind: game.TargetCard, ID: bear})
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		attackers := []uuid.UUID{bear}
		if !alone {
			attackers = append(attackers, friend)
		}
		for _, a := range attackers {
			if err := g.DeclareAttacker(a, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		// Nonland permanents: Bear, Friend, Rock, the Aura = 4; the Forest is a land.
		want := 2
		if alone {
			want = 2 + 4
		}
		if p, _ := effectstest.PtOf(t, g, bear); p != want {
			t.Errorf("attacking alone=%v: the creature has power %d, want %d", alone, p, want)
		}
	}
}

// --- Bottle-Cap Blast -----------------------------------------------------

func gapB67CastBlast(t *testing.T, g *game.Game, me *game.Player, target game.TargetRef) {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	blast := effectstest.HandCardFull(me, "Bottle-Cap Blast", "Instant", "{4}{R}", gapB67BottleCap, []string{"R"})
	effectstest.FloatMana(t, g, me, "{R}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, blast, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{target}}); err != nil {
		t.Fatalf("cast Bottle-Cap Blast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapB67BottleCapBlastMakesTappedTreasuresForExcessDamageToAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	gapB67CastBlast(t, g, me, game.TargetRef{Kind: game.TargetCard, ID: victim})
	if g.Battlefield.Contains(victim) {
		t.Error("the creature survived 5 damage")
	}
	treasures := effectstest.TmtB1Named(g, "Treasure")
	if len(treasures) != 3 {
		t.Fatalf("%d Treasures, want 3 (5 damage to a 2/2)", len(treasures))
	}
	for _, id := range treasures {
		if c := effectstest.LayeredCard(t, g, id); c.Controller != me.ID || !c.Tapped {
			t.Errorf("a Treasure is controlled by %v and tapped=%v, want my tapped Treasure", c.Controller, c.Tapped)
		}
	}
}

func TestGapB67BottleCapBlastMakesNothingForDamageToAPlayerOrAnExactKill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	gapB67CastBlast(t, g, me, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != life-5 {
		t.Errorf("life %d, want %d", opp.Life, life-5)
	}
	if got := len(effectstest.TmtB1Named(g, "Treasure")); got != 0 {
		t.Errorf("%d Treasures from damage to a player, want 0", got)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	exact := effectstest.B12Creature(g2, opp2.ID, "Their Giant", "Creature — Giant", 5, 5)
	gapB67CastBlast(t, g2, me2, game.TargetRef{Kind: game.TargetCard, ID: exact})
	if g2.Battlefield.Contains(exact) {
		t.Error("a 5/5 survived 5 damage")
	}
	if got := len(effectstest.TmtB1Named(g2, "Treasure")); got != 0 {
		t.Errorf("%d Treasures from exactly lethal damage, want 0", got)
	}
}

// --- Battle of Hoover Dam -------------------------------------------------

func TestGapB67HooverDamNCRReturnsACheapCreatureWithAFinalityCounterAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheap := effectstest.PushGraveyardPermanent(me, "Cheap Bear", "Creature — Bear", "{3}")
	pricey := effectstest.PushGraveyardPermanent(me, "Pricey Bear", "Creature — Bear", "{4}")
	effectstest.CastSiege(t, g, "Battle of Hoover Dam", gapB67HooverDam, "NCR")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt at the end step")
	}
	if effectstest.HasID(pick.PickTargetCards, pricey) {
		t.Error("a creature card with mana value 4 is offered")
	}
	effectstest.PickCard(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Fatal("the creature card did not return")
	}
	if got := effectstest.CountersOn(g, cheap, game.CounterFinality); got != 1 {
		t.Errorf("%d finality counters, want 1", got)
	}
}

func TestGapB67HooverDamLegionPutsTwoCountersOnATargetWhenYourCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, me.ID, "Victim", "Creature — Bear", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	effectstest.CastSiege(t, g, "Battle of Hoover Dam", gapB67HooverDam, "Legion")
	effectstest.B27Kill(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, bear); got != 0 {
		t.Fatalf("%d counters after an opponent's creature died, want 0", got)
	}
	effectstest.B27Kill(g, victim)
	effectstest.B27Settle(t, g, me.ID, bear)
	if got := effectstest.PlusOnes(g, bear); got != 2 {
		t.Errorf("%d +1/+1 counters after my creature died, want 2", got)
	}
}

// --- Overencumbered -------------------------------------------------------

func TestGapB67OverencumberedGivesTheEnchantedOpponentThreeTokensAndStopsAttacksUnlessPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Overencumbered", "Enchantment — Aura", gapB67Overencumber, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, name := range []string{"Clue", "Food", "Junk"} {
		if got := effectstest.TmtB1Count(g, opp.ID, name); got != 1 {
			t.Errorf("the enchanted opponent has %d %s tokens, want 1", got, name)
		}
		if got := effectstest.TmtB1Count(g, me.ID, name); got != 0 {
			t.Errorf("I have %d %s tokens, want 0", got, name)
		}
	}
	attacker := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepBeginCombat)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatalf("no payment prompt for the enchanted opponent: %+v", g.PendingChoices)
	}
	prompt := effectstest.AnswerPayUnless(t, g, opp.ID, false)
	_ = prompt
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, me.ID); err == nil {
		t.Error("a creature attacked although the cost was not paid")
	}
}

func TestGapB67OverencumberedLetsTheEnchantedOpponentAttackWhenTheyPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Overencumbered", "Enchantment — Aura", gapB67Overencumber, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	attacker := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepBeginCombat)
	effectstest.FloatMana(t, g, opp, "{C}{C}{C}")
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, me.ID); err != nil {
		t.Errorf("a creature could not attack after the cost was paid: %v", err)
	}
}

// --- Vault 101: Birthday Party --------------------------------------------

func TestGapB67Vault101MakesTokensThenPutsAnEquipmentOntoTheBattlefieldAndAttachesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	sword := effectstest.TmtB11Hand(g, me, "Test Sword", "Artifact — Equipment", "", "{3}", 0, 0)
	saga := effectstest.CastCatalogSpell(t, g, "Vault 101: Birthday Party", "Enchantment — Saga", gapB67Vault101, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Human Soldier"); got != 1 {
		t.Errorf("%d Human Soldier tokens, want 1", got)
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food tokens, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(saga, game.CounterLore, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatalf("no card choice for chapter II: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{sword}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sword) {
		t.Fatal("the Equipment did not enter the battlefield")
	}
	c = effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatalf("no attach offer: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, sword).AttachedTo.ID; got != bear {
		t.Errorf("the Equipment is attached to %v, want the bear", got)
	}
}

func TestGapB67Vault101ChapterIIIFromTheGraveyardMayBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sword := effectstest.PushGraveyardPermanent(me, "Test Sword", "Artifact — Equipment", "{3}")
	saga := effectstest.CastCatalogSpell(t, g, "Vault 101: Birthday Party", "Enchantment — Saga", gapB67Vault101, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(saga, game.CounterLore, 2) })
	effectstest.SettleOrderingTriggers(t, g)
	for i := 0; i < 2; i++ {
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
		if c == nil {
			t.Fatalf("no card choice for chapter %d: %+v", i+2, g.PendingChoices)
		}
		if !effectstest.HasID(c.ChooseCards, sword) {
			t.Errorf("the graveyard Equipment is not offered at chapter %d", i+2)
		}
		if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
			t.Fatalf("decline: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(sword) {
		t.Error("a declined Equipment entered the battlefield")
	}
}

// --- Inventory Management -------------------------------------------------

func TestGapB67InventoryManagementAttachesEachAuraAndEquipmentYouChoose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	other := effectstest.B12Push(g, me.ID, "Other", "Creature — Bear", "", 2, 2)
	sword := effectstest.PushCatalogPermanent(g, me.ID, "Test Sword", "Artifact — Equipment", "", false)
	axe := effectstest.PushCatalogPermanent(g, me.ID, "Test Axe", "Artifact — Equipment", "", false)
	effectstest.CastCatalogSpell(t, g, "Inventory Management", "Instant", gapB67Inventory, nil)
	effectstest.PassPriorityAroundTable(t, g)
	answered := 0
	for i := 0; i < 4; i++ {
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
		if c == nil {
			break
		}
		pick := []uuid.UUID{bear}
		if answered == 1 {
			pick = nil
		}
		if err := g.ResolveChooseCards(c.ID, me.ID, pick); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		answered++
		effectstest.PassPriorityAroundTable(t, g)
	}
	if answered != 2 {
		t.Fatalf("%d attach offers, want one for each of the two Equipment", answered)
	}
	attached := 0
	for _, id := range []uuid.UUID{sword, axe} {
		if effectstest.LayeredCard(t, g, id).AttachedTo.ID == bear {
			attached++
		}
	}
	if attached != 1 {
		t.Errorf("%d Equipment attached to the bear, want exactly the one I chose", attached)
	}
	_ = other
}

func TestGapB67InventoryManagementHasSplitSecond(t *testing.T) {
	spec, _ := effects.Lookup(gapB67Inventory)
	found := false
	for _, k := range spec.PrintedKeywords {
		if k == game.KeywordSplitSecond {
			found = true
		}
	}
	if !found {
		t.Error("Inventory Management lacks split second")
	}
}
