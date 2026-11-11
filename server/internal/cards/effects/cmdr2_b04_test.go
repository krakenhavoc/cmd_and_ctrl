package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b04_test.go — Commander-deck wave 2, batch 4 (#5368): choose a
// creature type.

const (
	cmdr2HauntingVoyageOracle = "334c284e-7bc3-4c18-9f8d-4d846f80008e"
	cmdr2KindredSummonsOracle = "b488e75f-af9c-43e2-8aca-f269fa2c720e"
	cmdr2PactSerpentOracle    = "f275039e-3dba-49b0-92aa-92721b1c7f95"
	cmdr2EtchingsOracle       = "0114f97f-b196-4af1-90aa-eba164cb8f22"
	cmdr2FuryOracle           = "97079964-7a2e-42c5-ad4a-9015c18a1e97"
	cmdr2RubbleOracle         = "76d07c99-8c6b-4c2c-a0f6-47541999015b"
	cmdr2BoonOracle           = "f04441b0-755c-4e7b-a9e4-4e03b8175ab6"
	cmdr2IconOracle           = "2b4973cc-83b0-4986-a19e-b4252706527e"
	cmdr2HeirloomBladeOracle  = "7c3a8766-e440-48c5-a561-197490efdba5"
)

func TestCmdr2B04CardsAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{
		cmdr2HauntingVoyageOracle, cmdr2KindredSummonsOracle, cmdr2PactSerpentOracle, cmdr2EtchingsOracle,
		cmdr2FuryOracle, cmdr2RubbleOracle, cmdr2BoonOracle, cmdr2IconOracle, cmdr2HeirloomBladeOracle,
	} {
		spec, ok := Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v, want Full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// cmdr2Goblin and cmdr2Bear seed battlefield creatures of a known type.
func cmdr2Creature(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return b12Creature(g, owner, name, typeLine, 2, 2)
}

// --- Icon of Ancestry ---------------------------------------------------------

func TestCmdr2B04IconOfAncestryPumpsTheChosenTypeAndLooksForOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gob := cmdr2Creature(g, me.ID, "Goblin", "Creature — Goblin")
	bear := cmdr2Creature(g, me.ID, "Bear", "Creature — Bear")
	icon := pushNamedTribePermanent(t, g, me.ID, "Icon of Ancestry", "Artifact", cmdr2IconOracle, "Goblin")
	if got := effectivePower(t, g, gob); got != 3 {
		t.Errorf("Goblin power %d, want 3", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("Bear power %d, want 2", got)
	}

	// Library top three: a Bear, a Goblin, a land. A fourth Goblin is too deep.
	deep := seedLibraryTop(me, "Deep Goblin", "Creature — Goblin", "{R}")
	land := seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	hit := seedLibraryTop(me, "Hit Goblin", "Creature — Goblin", "{R}")
	miss := seedLibraryTop(me, "Miss Bear", "Creature — Bear", "{G}")
	hand := me.Hand.Size()
	fillPoolColored(me, "C", 3)
	if err := g.ActivateCatalogAbility(me.ID, icon, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if p == nil {
		t.Fatal("no choose_cards prompt")
	}
	if len(p.ChooseCards) != 1 || p.ChooseCards[0] != hit {
		t.Fatalf("offered %v, want only the Goblin among the top three", p.ChooseCards)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{hit}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(hit) || me.Hand.Size() != hand+1 {
		t.Error("the Goblin did not reach the hand")
	}
	if len(me.Library.Cards) < 2 {
		t.Fatal("the library is too short")
	}
	bottom := map[uuid.UUID]bool{me.Library.Cards[0].InstanceID: true, me.Library.Cards[1].InstanceID: true}
	if !bottom[miss] || !bottom[land] {
		t.Error("the two cards that were not taken should be the bottom two of the library")
	}
	if !me.Library.Contains(deep) {
		t.Error("the fourth card left the library")
	}
	if n := len(me.Library.Cards); me.Library.Cards[n-1].InstanceID != deep {
		t.Errorf("the top card is %v, want the card that was fourth", me.Library.Cards[n-1].Name)
	}
}

// --- Etchings of the Chosen ---------------------------------------------------

func TestCmdr2B04EtchingsPumpsAndSacrificesForIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gob := cmdr2Creature(g, me.ID, "Goblin", "Creature — Goblin")
	bear := cmdr2Creature(g, me.ID, "Bear", "Creature — Bear")
	et := pushNamedTribePermanent(t, g, me.ID, "Etchings of the Chosen", "Enchantment", cmdr2EtchingsOracle, "Goblin")
	if got := effectivePower(t, g, gob); got != 3 {
		t.Errorf("Goblin power %d, want 3", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("Bear power %d, want 2", got)
	}
	// The Bear is not of the chosen type: it can't be sacrificed.
	fillPoolColored(me, "C", 2)
	if err := g.ActivateCatalogAbility(me.ID, et, 0, game.ActivateAbilityParams{Strict: true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("sacrificed a creature of another type")
	}
	if err := g.ActivateCatalogAbility(me.ID, et, 0, game.ActivateAbilityParams{Strict: true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		SacrificeIDs: []uuid.UUID{gob}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(gob) {
		t.Error("the Goblin was not sacrificed")
	}
	if !hasKeywordOnBattlefield(t, g, bear, "indestructible") {
		t.Error("the Bear did not gain indestructible")
	}
}

// --- Kindred Boon -------------------------------------------------------------

func TestCmdr2B04KindredBoonCountersMakeCreaturesIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gob := cmdr2Creature(g, me.ID, "Goblin", "Creature — Goblin")
	bear := cmdr2Creature(g, me.ID, "Bear", "Creature — Bear")
	boon := pushNamedTribePermanent(t, g, me.ID, "Kindred Boon", "Enchantment", cmdr2BoonOracle, "Goblin")
	fillPoolColored(me, "W", 2)
	if err := g.ActivateCatalogAbility(me.ID, boon, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("put a divinity counter on a creature of another type")
	}
	if err := g.ActivateCatalogAbility(me.ID, boon, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: gob}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, gob); c.Counters["divinity"] != 1 {
		t.Fatalf("divinity counters = %d, want 1", c.Counters["divinity"])
	}
	if !hasKeywordOnBattlefield(t, g, gob, "indestructible") {
		t.Error("the Goblin with a divinity counter is not indestructible")
	}
	if hasKeywordOnBattlefield(t, g, bear, "indestructible") {
		t.Error("the Bear is indestructible without a counter")
	}
	destroyIt(t, g, gob)
	if !g.Battlefield.Contains(gob) {
		t.Error("an indestructible Goblin was destroyed")
	}
}

// --- From the Rubble ----------------------------------------------------------

func TestCmdr2B04FromTheRubbleReturnsAChosenTypeCreatureWithAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushNamedTribePermanent(t, g, me.ID, "From the Rubble", "Enchantment", cmdr2RubbleOracle, "Goblin")
	gob := pushGraveyardCardTyped(me, "Dead Goblin", "Creature — Goblin")
	bear := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	advanceToEndStep(t, g)
	for i := 0; i < 6; i++ {
		if p := latestPickTarget(g, me.ID); p != nil {
			if hasID(p.PickTargetCards, bear) || !hasID(p.PickTargetCards, gob) {
				t.Fatalf("offered %v, want only the Goblin", p.PickTargetCards)
			}
			pickCard(t, g, me.ID, gob)
		}
		if stackFullyEmpty(g) {
			break
		}
		passPriorityAroundTable(t, g)
	}
	c, ok := battlefieldCard(g, gob)
	if !ok {
		t.Fatal("the Goblin did not return")
	}
	if c.Counters[game.CounterFinality] != 1 {
		t.Errorf("finality counters = %d, want 1", c.Counters[game.CounterFinality])
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the Bear left the graveyard")
	}
}

// --- Pact of the Serpent -----------------------------------------------------

func TestCmdr2B04PactOfTheSerpentDrawsAndLosesPerCreatureOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr2Creature(g, opp.ID, "Goblin A", "Creature — Goblin")
	cmdr2Creature(g, opp.ID, "Goblin B", "Creature — Goblin")
	cmdr2Creature(g, opp.ID, "Bear", "Creature — Bear")
	cmdr2Creature(g, me.ID, "My Goblin", "Creature — Goblin")
	castCatalogSpell(t, g, "Pact of the Serpent", "Sorcery", cmdr2PactSerpentOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	hand, life := opp.Hand.Size(), opp.Life
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Goblin")
	if got := opp.Hand.Size() - hand; got != 2 {
		t.Errorf("the target drew %d, want 2 (their two Goblins)", got)
	}
	if got := life - opp.Life; got != 2 {
		t.Errorf("the target lost %d life, want 2", got)
	}
}

// --- Haunting Voyage -----------------------------------------------------------

func cmdr2VoyageTable(t *testing.T) (*game.Game, *game.Player, []uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	var goblins []uuid.UUID
	for _, n := range []string{"Goblin A", "Goblin B", "Goblin C"} {
		goblins = append(goblins, pushGraveyardCardTyped(me, n, "Creature — Goblin"))
	}
	bear := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	return g, me, goblins, bear
}

func TestCmdr2B04HauntingVoyageReturnsUpToTwoOfTheChosenType(t *testing.T) {
	g, me, goblins, bear := cmdr2VoyageTable(t)
	castCatalogSpell(t, g, "Haunting Voyage", "Sorcery", cmdr2HauntingVoyageOracle, nil)
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Goblin")
	p := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if p == nil {
		t.Fatal("no choose_cards prompt")
	}
	if p.ChooseMin != 0 || p.ChooseMax != 2 || len(p.ChooseCards) != 3 || hasID(p.ChooseCards, bear) {
		t.Fatalf("offered %v in [%d,%d], want the three Goblins, up to two", p.ChooseCards, p.ChooseMin, p.ChooseMax)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, goblins[:3]); err == nil {
		t.Fatal("three cards were accepted for 'up to two'")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, goblins[:2]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(goblins[0]) || !g.Battlefield.Contains(goblins[1]) || g.Battlefield.Contains(goblins[2]) {
		t.Error("exactly the two chosen Goblins should have returned")
	}
}

func TestCmdr2B04HauntingVoyageForetoldReturnsThemAll(t *testing.T) {
	g, me, goblins, bear := cmdr2VoyageTable(t)
	advanceTo(t, g, game.StepPrecombatMain)
	card := handCardForTest(me, "Haunting Voyage", "Sorcery", cmdr2HauntingVoyageOracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, card, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell: %v", err)
	}
	g.WithWriteLock(func() { g.Turn.Seq++ })
	fillPoolColored(me, "B", 2)
	fillPoolColored(me, "C", 5)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyForetell,
	}); err != nil {
		t.Fatalf("cast the foretold spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Goblin")
	for _, id := range goblins {
		if !g.Battlefield.Contains(id) {
			t.Errorf("a Goblin stayed in the graveyard")
		}
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear returned")
	}
}

// --- Kindred Summons -----------------------------------------------------------

func TestCmdr2B04KindredSummonsRevealsXCreatureCardsOfTheType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmdr2Creature(g, me.ID, "My Goblin A", "Creature — Goblin")
	cmdr2Creature(g, me.ID, "My Goblin B", "Creature — Goblin")
	cmdr2Creature(g, me.ID, "My Bear", "Creature — Bear")
	// Top to bottom: Bear card, Goblin 1, Forest, Goblin 2, Goblin 3.
	g3 := seedLibraryTop(me, "Library Goblin 3", "Creature — Goblin", "{R}")
	g2 := seedLibraryTop(me, "Library Goblin 2", "Creature — Goblin", "{R}")
	forest := seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	g1 := seedLibraryTop(me, "Library Goblin 1", "Creature — Goblin", "{R}")
	bearCard := seedLibraryTop(me, "Library Bear", "Creature — Bear", "{G}")
	castCatalogSpell(t, g, "Kindred Summons", "Instant", cmdr2KindredSummonsOracle, nil)
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Goblin")
	if !g.Battlefield.Contains(g1) || !g.Battlefield.Contains(g2) {
		t.Error("the first two Goblin cards should have entered (X = 2)")
	}
	if g.Battlefield.Contains(g3) {
		t.Error("a third Goblin entered; the reveal stops at X = 2")
	}
	if g.Battlefield.Contains(bearCard) || g.Battlefield.Contains(forest) {
		t.Error("a non-Goblin entered the battlefield")
	}
	for _, id := range []uuid.UUID{g3, forest, bearCard} {
		if !me.Library.Contains(id) {
			t.Errorf("a revealed card that stayed was not shuffled back into the library")
		}
	}
}

// --- Descendants' Fury ----------------------------------------------------------

func TestCmdr2B04DescendantsFurySacrificesOneForACreatureThatSharesAType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Descendants' Fury", "Enchantment", cmdr2FuryOracle, false)
	elf := cmdr2Creature(g, me.ID, "Elf Warrior", "Creature — Elf Warrior")
	bear := cmdr2Creature(g, me.ID, "Bear", "Creature — Bear")
	bearCard := seedLibraryTop(me, "Library Bear", "Creature — Bear", "{G}")
	wrong := seedLibraryTop(me, "Library Goblin", "Creature — Goblin", "{R}")
	hit := seedLibraryTop(me, "Library Warrior", "Creature — Human Warrior", "{R}")
	_ = bearCard
	dealCombatDamageToPlayer(g, elf, opp.ID, 2)
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	p := cmdr2WaitForChooseCards(t, g, me.ID)
	if len(p.ChooseCards) != 2 || p.ChooseMin != 0 || p.ChooseMax != 1 {
		t.Fatalf("offered %v in [%d,%d], want the two creatures that connected, up to one", p.ChooseCards, p.ChooseMin, p.ChooseMax)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{elf}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(elf) {
		t.Error("the Elf Warrior was not sacrificed")
	}
	if !g.Battlefield.Contains(hit) {
		t.Error("the Warrior card should have entered (shares Warrior with the sacrificed Elf Warrior)")
	}
	if g.Battlefield.Contains(wrong) {
		t.Error("a Goblin entered")
	}
	if !me.Library.Contains(wrong) {
		t.Error("the revealed Goblin should be on the bottom of the library")
	}
}

func TestCmdr2B04DescendantsFuryDecliningSacrificesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Descendants' Fury", "Enchantment", cmdr2FuryOracle, false)
	elf := cmdr2Creature(g, me.ID, "Elf Warrior", "Creature — Elf Warrior")
	seedLibraryTop(me, "Library Warrior", "Creature — Human Warrior", "{R}")
	dealCombatDamageToPlayer(g, elf, opp.ID, 2)
	p := cmdr2WaitForChooseCards(t, g, me.ID)
	if err := g.ResolveChooseCards(p.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(elf) {
		t.Error("the Elf was sacrificed anyway")
	}
}

// cmdr2WaitForChooseCards passes priority until a choose_cards prompt opens.
func cmdr2WaitForChooseCards(t *testing.T, g *game.Game, who uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 8; i++ {
		if p := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, who); p != nil {
			return p
		}
		passPriorityAroundTable(t, g)
	}
	t.Fatal("no choose_cards prompt opened")
	return nil
}

// --- Heirloom Blade ---------------------------------------------------------------

func TestCmdr2B04HeirloomBladePumpsAndFindsACreatureThatSharesATypeWhenItsCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	elf := cmdr2Creature(g, me.ID, "Elf Warrior", "Creature — Elf Warrior")
	blade := pushCatalogPermanent(g, me.ID, "Heirloom Blade", "Artifact — Equipment", cmdr2HeirloomBladeOracle, false)
	equipTo(t, g, me.ID, blade, elf)
	if got := effectivePower(t, g, elf); got != 5 {
		t.Fatalf("equipped power %d, want 5", got)
	}
	wrong := seedLibraryTop(me, "Library Bear", "Creature — Bear", "{G}")
	hit := seedLibraryTop(me, "Library Warrior", "Creature — Human Warrior", "{R}")
	hand := me.Hand.Size()
	destroyIt(t, g, elf)
	assistCardSettle(t, g, me.ID, nil, true)
	if !me.Hand.Contains(hit) || me.Hand.Size() != hand+1 {
		t.Error("the Warrior card did not reach the hand")
	}
	if !me.Library.Contains(wrong) {
		t.Error("the revealed Bear should be on the bottom of the library")
	}
}

func TestCmdr2B04HeirloomBladeRevealIsOptional(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	elf := cmdr2Creature(g, me.ID, "Elf Warrior", "Creature — Elf Warrior")
	blade := pushCatalogPermanent(g, me.ID, "Heirloom Blade", "Artifact — Equipment", cmdr2HeirloomBladeOracle, false)
	equipTo(t, g, me.ID, blade, elf)
	hit := seedLibraryTop(me, "Library Warrior", "Creature — Human Warrior", "{R}")
	destroyIt(t, g, elf)
	assistCardSettle(t, g, me.ID, nil, false)
	if me.Hand.Contains(hit) {
		t.Error("declining still took the card")
	}
}
