package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 006, group A: Koth of the Hammer, Huatli Radiant Champion,
// Kaya Ghost Assassin, Nahiri Storm of Stone (#5462).

const (
	s82OldB6Koth   = "589186e0-3ebe-46a3-a9e7-798b1712fa93"
	s82OldB6Huatli = "5ab5b82a-83e1-44d7-b8c3-9cd7accc1ca4"
	s82OldB6Kaya   = "5e4c81b9-a31e-45e6-983f-30f11b8a05c0"
	s82OldB6Nahiri = "6332d6d3-1d77-486c-b040-02fc7225f245"
)

func TestS82OldB006ACardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		s82OldB6Koth: "Koth of the Hammer", s82OldB6Huatli: "Huatli, Radiant Champion",
		s82OldB6Kaya: "Kaya, Ghost Assassin", s82OldB6Nahiri: "Nahiri, Storm of Stone",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v %v", name, spec, ok)
		}
	}
}

func s82OldB6Mountain(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, owner, "Mountain", "Basic Land — Mountain", "", false)
}

// --- Koth of the Hammer -----------------------------------------------------

func TestS82OldB006KothPlusOneUntapsAndAnimatesAMountain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	koth := effectstest.S82Walker(g, me.ID, "Koth of the Hammer", s82OldB6Koth, 3)
	mtn := s82OldB6Mountain(g, me.ID)
	forest := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, mtn).Tapped = true })
	effectstest.AdvanceToMain(t, g)
	// A Forest is not a legal target.
	if err := g.ActivateCatalogAbility(me.ID, koth, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(forest)}); err == nil {
		t.Fatal("+1 targeted a Forest")
	}
	effectstest.S82Activate(t, g, me.ID, koth, 0, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(mtn)})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ApaLive(g, mtn)
	if c.Tapped {
		t.Error("Mountain is still tapped")
	}
	if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !c.HasSubtypeWord("Elemental") {
		t.Errorf("Mountain is %d/%d creature=%v land=%v", c.CurrentPower(), c.CurrentToughness(), c.IsCreature(), c.IsLand())
	}
	if got := effectstest.LoyaltyOf(g, koth); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
}

func TestS82OldB006KothMinusTwoAddsRedPerMountain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	koth := effectstest.S82Walker(g, me.ID, "Koth of the Hammer", s82OldB6Koth, 3)
	for i := 0; i < 3; i++ {
		s82OldB6Mountain(g, me.ID)
	}
	s82OldB6Mountain(g, opp.ID) // not mine
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, koth, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HazelPoolTotal(me); got != 3 {
		t.Errorf("pool has %d mana, want 3 (three Mountains I control)", got)
	}
	if got := effectstest.LoyaltyOf(g, koth); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

func TestS82OldB006KothEmblemGivesMountainsAPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	koth := effectstest.S82Walker(g, me.ID, "Koth of the Hammer", s82OldB6Koth, 5)
	mtn := s82OldB6Mountain(g, me.ID)
	theirs := s82OldB6Mountain(g, opp.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, koth, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.EmblemsOf(g, me.ID)) != 1 {
		t.Fatal("no emblem")
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(opp.ID, theirs, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Error("an opponent's Mountain got the ping")
	}
	effectstest.S82Activate(t, g, me.ID, mtn, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
	if c, _ := effectstest.BattlefieldCard(g, mtn); !c.Tapped {
		t.Error("the Mountain did not tap for the ping")
	}
}

// --- Huatli, Radiant Champion ------------------------------------------------

func TestS82OldB006HuatliPlusOnePutsALoyaltyCounterPerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	huatli := effectstest.S82Walker(g, me.ID, "Huatli, Radiant Champion", s82OldB6Huatli, 3)
	effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 1, 1) // not mine
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, huatli, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, huatli); got != 3+1+2 {
		t.Errorf("loyalty %d, want 6 (the +1 and two creatures)", got)
	}
}

func TestS82OldB006HuatliMinusOneGivesPlusXPlusX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	huatli := effectstest.S82Walker(g, me.ID, "Huatli, Radiant Champion", s82OldB6Huatli, 3)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, huatli, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(a)})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ApaLive(g, a); c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("target is %d/%d, want 5/5 (+3/+3 for three creatures)", c.CurrentPower(), c.CurrentToughness())
	}
	if got := effectstest.LoyaltyOf(g, huatli); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestS82OldB006HuatliEmblemOffersADrawForEachCreatureThatEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	huatli := effectstest.S82Walker(g, me.ID, "Huatli, Radiant Champion", s82OldB6Huatli, 8)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, huatli, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.EmblemsOf(g, me.ID)) != 1 {
		t.Fatal("no emblem")
	}
	hand := me.Hand.Size()
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	hand = me.Hand.Size() - 1
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, bear); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- Kaya, Ghost Assassin ----------------------------------------------------

func TestS82OldB006KayaZeroWithNoTargetExilesHerAndReturnsAtNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kaya := effectstest.S82Walker(g, me.ID, "Kaya, Ghost Assassin", s82OldB6Kaya, 5)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.S82Activate(t, g, me.ID, kaya, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if _, on := effectstest.BattlefieldCard(g, kaya); on {
		t.Fatal("Kaya is still on the battlefield")
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	if effectstest.CountBattlefieldByName(g, "Kaya, Ghost Assassin") != 0 {
		t.Fatal("Kaya came back before the upkeep")
	}
	s82OldB6ToNextUpkeepOf(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldByName(g, "Kaya, Ghost Assassin") != 1 {
		t.Fatal("Kaya did not return at the beginning of my next upkeep")
	}
}

func TestS82OldB006KayaZeroCanExileATargetCreatureInstead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	kaya := effectstest.S82Walker(g, me.ID, "Kaya, Ghost Assassin", s82OldB6Kaya, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, kaya, 0, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if _, on := effectstest.BattlefieldCard(g, bear); on {
		t.Fatal("the Bear is still on the battlefield")
	}
	if _, on := effectstest.BattlefieldCard(g, kaya); !on {
		t.Fatal("Kaya left too")
	}
	s82OldB6ToNextUpkeepOf(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldByName(g, "Bear") != 1 {
		t.Error("the Bear did not return under its owner's control")
	}
	if got := effectstest.CountBattlefieldByName(g, "Kaya, Ghost Assassin"); got != 1 {
		t.Errorf("%d Kayas, want 1", got)
	}
}

func TestS82OldB006KayaZeroCanChooseToExileHerselfEvenWithATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	kaya := effectstest.S82Walker(g, me.ID, "Kaya, Ghost Assassin", s82OldB6Kaya, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, kaya, 0, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if _, on := effectstest.BattlefieldCard(g, kaya); on {
		t.Error("Kaya stayed")
	}
	if _, on := effectstest.BattlefieldCard(g, bear); !on {
		t.Error("the Bear was exiled")
	}
}

func TestS82OldB006KayaMinusOneDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kaya := effectstest.S82Walker(g, me.ID, "Kaya, Ghost Assassin", s82OldB6Kaya, 5)
	effectstest.AdvanceToMain(t, g)
	mine, theirs := me.Life, []int{}
	for _, s := range g.Seats {
		theirs = append(theirs, s.Life)
	}
	effectstest.S82Activate(t, g, me.ID, kaya, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine+2 {
		t.Errorf("my life %d, want %d", me.Life, mine+2)
	}
	for i, s := range g.Seats {
		if s.ID != me.ID && s.Life != theirs[i]-2 {
			t.Errorf("seat %d life %d, want %d", i, s.Life, theirs[i]-2)
		}
	}
}

func TestS82OldB006KayaMinusTwoDiscardsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	kaya := effectstest.S82Walker(g, me.ID, "Kaya, Ghost Assassin", s82OldB6Kaya, 5)
	card := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: card, Name: "Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.S82Activate(t, g, me.ID, kaya, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, card)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(card) {
		t.Error("the opponent kept the card")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func s82OldB6ToNextUpkeepOf(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	for i := 0; i < 80; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep && g.Seats[g.Turn.ActiveSeat].ID == me.ID {
			return
		}
	}
	t.Fatal("never reached my next upkeep")
}

// --- Nahiri, Storm of Stone --------------------------------------------------

func TestS82OldB006NahiriGivesFirstStrikeOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.S82Walker(g, me.ID, "Nahiri, Storm of Stone", s82OldB6Nahiri, 6)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Wolf", "Creature — Wolf", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("no first strike during my turn")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "first strike") {
		t.Error("an opponent's creature has first strike")
	}
	for g.Seats[g.Turn.ActiveSeat].ID == me.ID {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("first strike on an opponent's turn")
	}
}

func TestS82OldB006NahiriMinusXDamagesATappedCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	nahiri := effectstest.S82Walker(g, me.ID, "Nahiri, Storm of Stone", s82OldB6Nahiri, 6)
	tapped := effectstest.B12Creature(g, opp.ID, "Tapped Bear", "Creature — Bear", 2, 5)
	untapped := effectstest.B12Creature(g, opp.ID, "Untapped Bear", "Creature — Bear", 2, 5)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, tapped).Tapped = true })
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, nahiri, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: effectstest.S82TargetCard(untapped)}); err == nil {
		t.Fatal("−X targeted an untapped creature")
	}
	effectstest.S82Activate(t, g, me.ID, nahiri, 0, game.ActivateAbilityParams{XValue: 4, Targets: effectstest.S82TargetCard(tapped)})
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, tapped); c.DamageMarked != 4 {
		t.Errorf("damage %d, want 4", c.DamageMarked)
	}
	if got := effectstest.LoyaltyOf(g, nahiri); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestS82OldB006NahiriMakesEquipCheaperOnlyOnYourTurn(t *testing.T) {
	spec, ok := effects.Lookup(s82OldB6Nahiri)
	if !ok || len(spec.CostModifiers) != 1 {
		t.Fatalf("Nahiri declares %d cost modifiers, want the one equip discount", len(spec.CostModifiers))
	}
	m := spec.CostModifiers[0]
	if !m.Activations || m.AppliesTo == nil {
		t.Fatalf("the discount is not an activation-scoped modifier: %+v", m)
	}
}
