package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 7, group B: Sarkhan Vol, Ajani Goldmane, Huatli the Sun's
// Heart, Tezzeret Master of the Bridge (Secret Lair Drop, #5462).

const (
	oldB007Tezzeret = "d097991c-41df-47de-a439-57a0693b9581"
	oldB007Sarkhan  = "d902074a-2a2b-42b7-a1e5-750862c64fbe"
	oldB007Ajani    = "dfb5f660-fd8b-4b7b-934c-6a71cd182f15"
	oldB007Huatli   = "e1a13723-a6b1-4160-8eb5-f7a4860898a3"
)

func TestS82OldB007BCardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB007Tezzeret: "Tezzeret, Master of the Bridge", oldB007Sarkhan: "Sarkhan Vol",
		oldB007Ajani: "Ajani Goldmane", oldB007Huatli: "Huatli, the Sun's Heart",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v (found %v)", name, spec, ok)
		}
	}
}

// --- Sarkhan Vol ---------------------------------------------------------

func TestS82OldB007SarkhanPlusOnePumpsAndHastesYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	walker := effectstest.S82Walker(g, me.ID, "Sarkhan Vol", oldB007Sarkhan, 4)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ApaLive(g, mine); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("my Bear is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("my Bear has no haste")
	}
	if c := effectstest.ApaLive(g, theirs); c.CurrentPower() != 3 || effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Errorf("their Ogre was changed: %d/%d", c.CurrentPower(), c.CurrentToughness())
	}
	if got := effectstest.LoyaltyOf(g, walker); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestS82OldB007SarkhanMinusTwoStealsUntapsAndHastesUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	walker := effectstest.S82Walker(g, me.ID, "Sarkhan Vol", oldB007Sarkhan, 4)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ogre {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(ogre)})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ApaLive(g, ogre)
	if c.Controller != me.ID || c.Tapped {
		t.Errorf("Ogre: controller %v tapped %v, want mine and untapped", c.Controller, c.Tapped)
	}
	if !effectstest.HasEffectiveKeyword(t, g, ogre, "haste") {
		t.Error("no haste")
	}
	if got := effectstest.LoyaltyOf(g, walker); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if c := effectstest.ApaLive(g, ogre); c.Controller != opp.ID {
		t.Error("the Ogre did not go back at end of turn")
	}
}

func TestS82OldB007SarkhanMinusSixMakesFiveFlyingDragons(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.S82Walker(g, me.ID, "Sarkhan Vol", oldB007Sarkhan, 6)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Dragon" {
			n++
			if c.Power != 4 || c.Toughness != 4 || !game.HasKeyword(&c, "flying") {
				t.Errorf("Dragon %d/%d keywords %v", c.Power, c.Toughness, c.Keywords)
			}
		}
	}
	if n != 5 {
		t.Errorf("%d Dragons, want 5", n)
	}
	if _, ok := effectstest.BattlefieldCard(g, walker); ok {
		t.Error("Sarkhan at 0 loyalty should be gone")
	}
}

// --- Ajani Goldmane ------------------------------------------------------

func TestS82OldB007AjaniPlusOneGainsTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.S82Walker(g, me.ID, "Ajani Goldmane", oldB007Ajani, 4)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.S82Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestS82OldB007AjaniMinusOneCountersAndVigilanceOnYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	walker := effectstest.S82Walker(g, me.ID, "Ajani Goldmane", oldB007Ajani, 4)
	a := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if effectstest.CounterCount(g, id, "+1/+1") != 1 || !effectstest.HasEffectiveKeyword(t, g, id, "vigilance") {
			t.Errorf("creature %v: counters %d, vigilance %v", id, effectstest.CounterCount(g, id, "+1/+1"),
				effectstest.HasEffectiveKeyword(t, g, id, "vigilance"))
		}
	}
	if effectstest.CounterCount(g, theirs, "+1/+1") != 0 || effectstest.HasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("the opponent's creature was changed")
	}
}

func TestS82OldB007AjaniMinusSixMakesAnAvatarAsBigAsYourLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.S82Walker(g, me.ID, "Ajani Goldmane", oldB007Ajani, 6)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Avatar" {
			id = c.InstanceID
		}
	}
	if id == uuid.Nil {
		t.Fatal("no Avatar token")
	}
	if c := effectstest.EffectiveCardForTest(t, g, id); c.CurrentPower() != me.Life || c.CurrentToughness() != me.Life {
		t.Errorf("Avatar %d/%d, life %d", c.CurrentPower(), c.CurrentToughness(), me.Life)
	}
	g.WithWriteLock(func() { me.Life = 7 })
	g.RunStateChecksForTest()
	if c := effectstest.EffectiveCardForTest(t, g, id); c.CurrentPower() != 7 || c.CurrentToughness() != 7 {
		t.Errorf("after life change Avatar %d/%d, want 7/7", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Huatli, the Sun's Heart ---------------------------------------------

func TestS82OldB007HuatliMakesYourCreaturesAssignDamageByToughnessOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.S82Walker(g, me.ID, "Huatli, the Sun's Heart", oldB007Huatli, 7)
	mine := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	if !effectstest.GaLayered(t, g, mine).Effective().DamageByToughness {
		t.Error("my creature does not assign damage by toughness")
	}
	if effectstest.GaLayered(t, g, theirs).Effective().DamageByToughness {
		t.Error("an opponent's creature assigns damage by toughness")
	}
}

func TestS82OldB007HuatliMinusThreeGainsTheGreatestToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	walker := effectstest.S82Walker(g, me.ID, "Huatli, the Sun's Heart", oldB007Huatli, 7)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 5)
	effectstest.B12Creature(g, opp.ID, "Titan", "Creature — Giant", 9, 9) // not mine
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.S82Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d, want %d", me.Life, life+5)
	}
	if got := effectstest.LoyaltyOf(g, walker); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
}

func TestS82OldB007HuatliMinusThreeWithNoCreaturesGainsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.S82Walker(g, me.ID, "Huatli, the Sun's Heart", oldB007Huatli, 7)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.S82Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}

// --- Tezzeret, Master of the Bridge --------------------------------------

func TestS82OldB007TezzeretGivesCreatureAndPlaneswalkerSpellsAffinityForArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	effectstest.S82Walker(g, me.ID, "Tezzeret, Master of the Bridge", oldB007Tezzeret, 5)
	effectstest.B12Permanent(g, me.ID, "Artifact A", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Artifact B", "Artifact")
	creature := effectstest.HandCardFull(me, "Test Creature", "Creature — Bear", "{3}{U}", "", []string{"U"})
	effectstest.B06AddMana(me, "C", "U")
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("two artifacts should make {3}{U} cost {1}{U}: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want 0", me.ManaPool.Total())
	}
	effectstest.PassPriorityAroundTable(t, g)
	// An instant gets no reduction.
	instant := effectstest.HandCardFull(me, "Test Instant", "Instant", "{3}{U}", "", []string{"U"})
	effectstest.B06AddMana(me, "C", "U")
	if err := g.CastSpell(me.ID, instant, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("an instant was cheaper with two artifacts out")
	}
}

func TestS82OldB007TezzeretDoesNotReduceAnOpponentsSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	effectstest.S82Walker(g, opp.ID, "Tezzeret, Master of the Bridge", oldB007Tezzeret, 5)
	effectstest.B12Permanent(g, opp.ID, "Artifact A", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Artifact B", "Artifact")
	creature := effectstest.HandCardFull(me, "Test Creature", "Creature — Bear", "{3}{U}", "", []string{"U"})
	effectstest.B06AddMana(me, "C", "U")
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the opponent's Tezzeret reduced my spell")
	}
}

func TestS82OldB007TezzeretPlusTwoDealsXToEachOpponentAndGainsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.S82Walker(g, me.ID, "Tezzeret, Master of the Bridge", oldB007Tezzeret, 5)
	effectstest.B12Permanent(g, me.ID, "Artifact A", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Artifact B", "Artifact")
	effectstest.B12Permanent(g, g.Seats[1].ID, "Their Artifact", "Artifact")
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	var opps []int
	for i, p := range g.Seats {
		if p != me {
			opps = append(opps, i)
		}
	}
	before := map[int]int{}
	for _, i := range opps {
		before[i] = g.Seats[i].Life
	}
	effectstest.S82Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, i := range opps {
		if g.Seats[i].Life != before[i]-2 {
			t.Errorf("seat %d life %d, want %d", i, g.Seats[i].Life, before[i]-2)
		}
	}
	if me.Life != life+2 {
		t.Errorf("my life %d, want %d", me.Life, life+2)
	}
	if got := effectstest.LoyaltyOf(g, walker); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestS82OldB007TezzeretMinusThreeReturnsAnArtifactCardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.S82Walker(g, me.ID, "Tezzeret, Master of the Bridge", oldB007Tezzeret, 5)
	art := effectstest.PushGraveyardPermanent(me, "Sol Ring", "Artifact", "{1}")
	bear := effectstest.PushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceToMain(t, g)
	// A non-artifact card is not a legal target.
	err := g.ActivateCatalogAbility(me.ID, walker, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(bear)})
	if err == nil {
		t.Fatal("a creature card was a legal target")
	}
	effectstest.S82Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(art)})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(art) || me.Hand.Contains(bear) {
		t.Errorf("hand has artifact=%v creature=%v", me.Hand.Contains(art), me.Hand.Contains(bear))
	}
}

func TestS82OldB007TezzeretMinusEightPutsTheArtifactsAmongTheTopTenOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.S82Walker(g, me.ID, "Tezzeret, Master of the Bridge", oldB007Tezzeret, 8)
	// Each push lands on top: the eleventh card first, then eight fillers, then
	// two artifacts (one also a creature) make the top ten.
	effectstest.PutOnTopForTest(me, "Below The Ten", "Artifact")
	for i := 0; i < 8; i++ {
		effectstest.PutOnTopForTest(me, "Filler", "Sorcery")
	}
	effectstest.PutOnTopForTest(me, "Sol Ring", "Artifact")
	effectstest.PutOnTopForTest(me, "Golem", "Artifact Creature — Golem")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldByName(g, "Sol Ring") != 1 || effectstest.CountBattlefieldByName(g, "Golem") != 1 {
		t.Errorf("Sol Ring %d, Golem %d on the battlefield, want 1 each",
			effectstest.CountBattlefieldByName(g, "Sol Ring"), effectstest.CountBattlefieldByName(g, "Golem"))
	}
	if effectstest.CountBattlefieldByName(g, "Filler") != 0 {
		t.Error("a nonartifact card entered")
	}
	if effectstest.CountBattlefieldByName(g, "Below The Ten") != 0 {
		t.Error("the eleventh card entered")
	}
}
