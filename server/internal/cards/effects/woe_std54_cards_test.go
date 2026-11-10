package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_std54_cards_test.go — Wilds of Eldraine, standard batch 54 (#4471).
// Each test plays the real catalog entry through the engine.

const (
	std54RangerOracle    = "1189b5d5-ca8f-4b4e-a5f1-91122a485cf3"
	std54FoodFightOracle = "00571e7f-4e40-4a07-a529-34580e942338"
	std54FireboltOracle  = "f1882cc8-c81a-4131-b639-1bd6e083f8ba"
	std54CrescendoOracle = "0332f7b5-dad3-4fd0-b86c-78bd8301f59d"
	std54HearthOracle    = hearthElementalOracleID
	std54ImodaneOracle   = "d70d4686-2c35-4149-aed1-7e3b9e022702"
	std54RotisserieOracl = "e33c1744-843f-438f-a8c3-2d9f60120776"
	std54ChampionOracle  = "debdfab7-454f-4fec-9ee8-aa5b570ea7fb"
	std54BraveOracle     = "49d189dd-9ce7-434f-931b-7110cd5255e7"
	std54ArchivistOracle = "d7035db0-4bde-4ba3-9028-dd14191c8126"
)

func std54Seats(g *game.Game) (me, opp *game.Player) { return woeB1Seats(g) }

func TestStd54EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		std54RangerOracle:        "Boundary Lands Ranger",
		std54FoodFightOracle:     "Food Fight",
		std54FireboltOracle:      "Frantic Firebolt",
		std54CrescendoOracle:     "Gnawing Crescendo",
		std54HearthOracle:        "Hearth Elemental",
		std54HearthOracle + "#1": "Stoke Genius",
		std54ImodaneOracle:       "Imodane, the Pyrohammer",
		std54RotisserieOracl:     "Rotisserie Elemental",
		std54ChampionOracle:      "Agatha's Champion",
		std54BraveOracle:         "Brave the Wilds",
		std54ArchivistOracle:     "Elvish Archivist",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// std54Token puts an artifact token under owner's control: a permanent
// bargain can sacrifice.
func std54Token(g *game.Game, owner uuid.UUID) uuid.UUID { return woeILToken(g, owner) }

// std54AnswerTarget answers an open target pick with the given cards.
func std54AnswerTarget(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	if latestPickTarget(g, chooser) == nil {
		return
	}
	answerTriggerTargets(t, g, chooser, ids...)
}

// --- Boundary Lands Ranger ---------------------------------------------------------

func TestStd54BoundaryLandsRangerRummagesWithAPowerFourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := std54Seats(g)
	pushCatalogPermanent(g, me.ID, "Boundary Lands Ranger", "Creature — Human Ranger", std54RangerOracle, false)
	b12Creature(g, me.ID, "Big Bear", woeMPBearType, 4, 4)
	advanceTo(t, g, game.StepBeginCombat)
	monSettle(t, g)
	hand, grave := me.Hand.Size(), me.Graveyard.Size()
	answerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	monSettle(t, g)
	if me.Hand.Size() != hand || me.Graveyard.Size() != grave+1 {
		t.Errorf("hand %d graveyard %d, want %d and %d", me.Hand.Size(), me.Graveyard.Size(), hand, grave+1)
	}
}

func TestStd54BoundaryLandsRangerDoesNothingWithoutAPowerFourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := std54Seats(g)
	pushCatalogPermanent(g, me.ID, "Boundary Lands Ranger", "Creature — Human Ranger", std54RangerOracle, false)
	b12Creature(g, me.ID, "Small Bear", woeMPBearType, 3, 3)
	advanceTo(t, g, game.StepBeginCombat)
	monSettle(t, g)
	if discardChoiceFor(g, me.ID) != nil {
		t.Error("a discard prompt with no creature of power 4 or greater")
	}
}

// --- Food Fight -----------------------------------------------------------------

func std54ActivateGranted(t *testing.T, g *game.Game, me *game.Player, artifact uuid.UUID, target game.TargetRef) error {
	t.Helper()
	idx, ref := grantedActivatedIndex(t, g, artifact)
	if idx < 0 {
		t.Fatalf("%s has no granted activated ability", artifact)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	return g.ActivateCatalogAbility(me.ID, artifact, idx, game.ActivateAbilityParams{
		Ref: ref, Targets: []game.TargetRef{target},
	})
}

func TestStd54FoodFightGivesArtifactsASacrificeForDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := std54Seats(g)
	gaPut(g, me.ID, "Food Fight", "Enchantment", std54FoodFightOracle)
	rock := gaPut(g, me.ID, "Mind Stone", "Artifact", "")
	bear := gaPut(g, me.ID, "Bear", "Creature — Bear", "")
	theirs := gaPut(g, opp.ID, "Their Rock", "Artifact", "")
	if i, _ := grantedActivatedIndex(t, g, bear); i >= 0 {
		t.Error("a non-artifact has the granted ability")
	}
	if i, _ := grantedActivatedIndex(t, g, theirs); i >= 0 {
		t.Error("an opponent's artifact has the granted ability")
	}
	advanceToMain(t, g)
	life := lifeOf(g, opp.ID)
	if err := std54ActivateGranted(t, g, me, rock, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-2 {
		t.Errorf("opponent at %d, want %d (1 plus one Food Fight)", got, life-2)
	}
}

func TestStd54FoodFightCountsEveryFoodFightYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := std54Seats(g)
	gaPut(g, me.ID, "Food Fight", "Enchantment", std54FoodFightOracle)
	gaPut(g, me.ID, "Food Fight", "Enchantment", std54FoodFightOracle)
	gaPut(g, opp.ID, "Food Fight", "Enchantment", std54FoodFightOracle)
	rock := gaPut(g, me.ID, "Mind Stone", "Artifact", "")
	foe := b12Creature(g, opp.ID, "Foe", woeMPBearType, 5, 5)
	advanceToMain(t, g)
	if err := std54ActivateGranted(t, g, me, rock, game.TargetRef{Kind: game.TargetCard, ID: foe}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, foe); got != 3 {
		t.Errorf("foe has %d damage, want 3 (1 plus your two Food Fights, not theirs)", got)
	}
}

// --- Frantic Firebolt -------------------------------------------------------------

func TestStd54FranticFireboltCountsInstantsSorceriesAndAdventures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := std54Seats(g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	dftABInGraveyard(me, "Opt", "Instant")
	dftABInGraveyard(me, "Divination", "Sorcery")
	dftABInGraveyard(me, "Grizzly Bears", "Creature — Bear")
	dftABInGraveyard(me, "Forest", "Basic Land — Forest")
	me.Graveyard.PushTop(adventureCard(me.ID, adventureCase{oracle: std54HearthOracle, faces: [2]game.Face{
		{Name: "Hearth Elemental", TypeLine: "Creature — Elemental", ManaCost: "{5}{R}", Power: 4, Toughness: 5},
		{Name: "Stoke Genius", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{R}"}}}))
	foe := b12Creature(g, opp.ID, "Foe", woeMPBearType, 9, 9)
	castCatalogSpell(t, g, "Frantic Firebolt", "Instant", std54FireboltOracle, cardRefs(foe))
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, foe); got != 5 {
		t.Errorf("damage %d, want 5 (2 plus an instant, a sorcery and an Adventure card; not itself)", got)
	}
}

func TestStd54FranticFireboltDealsTwoWithAnEmptyGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := std54Seats(g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	foe := b12Creature(g, opp.ID, "Foe", woeMPBearType, 9, 9)
	castCatalogSpell(t, g, "Frantic Firebolt", "Instant", std54FireboltOracle, cardRefs(foe))
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, foe); got != 2 {
		t.Errorf("damage %d, want 2", got)
	}
}

// --- Gnawing Crescendo ------------------------------------------------------------

func TestStd54GnawingCrescendoPumpsAndMakesRatsForNontokenDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := std54Seats(g)
	bear := b12Creature(g, me.ID, "Bear", woeMPBearType, 2, 2)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Gnawing Crescendo", "Instant", std54CrescendoOracle, nil)
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, bear); p != 4 {
		t.Errorf("the Bear has power %d, want 4", p)
	}
	late := b12Creature(g, me.ID, "Latecomer", woeMPBearType, 2, 2)
	if p, _ := effectivePTOf(t, g, late); p != 2 {
		t.Errorf("a creature that entered later has power %d, want 2", p)
	}
	woeB1Destroy(t, g, token)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 0 {
		t.Fatalf("%d Rats after a token died, want 0", got)
	}
	woeB1Destroy(t, g, bear)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Fatalf("%d Rats after a nontoken creature died, want 1", got)
	}
	assertRestrictions(t, g, findBattlefieldByName(g, "Rat"), game.CantBlock)
	woeB1Destroy(t, g, late)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 2 {
		t.Errorf("%d Rats after a second death, want 2 (it triggers every time this turn)", got)
	}
}

func TestStd54GnawingCrescendoIgnoresAnOpponentsCreaturesDying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := std54Seats(g)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	castCatalogSpell(t, g, "Gnawing Crescendo", "Instant", std54CrescendoOracle, nil)
	passPriorityAroundTable(t, g)
	woeB1Destroy(t, g, theirs)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 0 {
		t.Errorf("%d Rats after an opponent's creature died, want 0", got)
	}
}

// --- Hearth Elemental // Stoke Genius ---------------------------------------------

func TestStd54HearthElementalCostsLessPerInstantSorceryOrAdventureCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := std54Seats(g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	const typeLine = "Creature — Elemental"
	if got := selfPricedMV(t, g, me, std54HearthOracle, typeLine, "{5}{R}"); got != 6 {
		t.Errorf("with an empty graveyard it costs mana value %d, want 6", got)
	}
	dftABInGraveyard(me, "Opt", "Instant")
	dftABInGraveyard(me, "Divination", "Sorcery")
	dftABInGraveyard(me, "Grizzly Bears", "Creature — Bear")
	me.Graveyard.PushTop(adventureCard(me.ID, adventureCase{oracle: "adv-oracle", faces: [2]game.Face{
		{Name: "Some Knight", TypeLine: "Creature — Human Knight", ManaCost: "{2}{W}", Power: 2, Toughness: 2},
		{Name: "Some Adventure", TypeLine: "Instant — Adventure", ManaCost: "{W}"}}}))
	if got := selfPricedMV(t, g, me, std54HearthOracle, typeLine, "{5}{R}"); got != 3 {
		t.Errorf("with three counting cards it costs mana value %d, want 3 ({2}{R})", got)
	}
	dftABInGraveyard(opp, "Their Opt", "Instant")
	if got := selfPricedMV(t, g, me, std54HearthOracle, typeLine, "{5}{R}"); got != 3 {
		t.Errorf("an opponent's graveyard changed the cost to %d, want 3", got)
	}
	for i := 0; i < 5; i++ {
		dftABInGraveyard(me, "Another Opt", "Instant")
	}
	if got := selfPricedMV(t, g, me, std54HearthOracle, typeLine, "{5}{R}"); got != 1 {
		t.Errorf("with eight cards it costs mana value %d, want 1 ({R} is never reduced)", got)
	}
}

func TestStd54StokeGeniusDiscardsYourHandAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := std54Seats(g)
	seedLibrary(me, "A", "B", "C", "D")
	card := woeB1Adventure(me, std54HearthOracle,
		game.Face{Name: "Hearth Elemental", TypeLine: "Creature — Elemental", ManaCost: "{5}{R}", Power: 4, Toughness: 5},
		game.Face{Name: "Stoke Genius", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{R}", Colors: []string{"R"}})
	hand := me.Hand.Size() - 1
	if hand < 1 {
		t.Fatal("setup: the hand is empty")
	}
	grave := me.Graveyard.Size()
	woeB1CastAdventure(t, g, me, card, nil)
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand %d, want 2 (the old hand discarded, two drawn)", got)
	}
	if got := me.Graveyard.Size(); got != grave+hand {
		t.Errorf("graveyard %d, want %d", got, grave+hand)
	}
}

// --- Imodane, the Pyrohammer ------------------------------------------------------

func TestStd54ImodaneMirrorsASpellsDamageToEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := std54Seats(g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	dftABInGraveyard(me, "Opt", "Instant")
	pushCatalogPermanent(g, me.ID, "Imodane, the Pyrohammer", "Legendary Creature — Human Knight", std54ImodaneOracle, false)
	foe := b12Creature(g, opp.ID, "Foe", woeMPBearType, 9, 9)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = lifeOf(g, p.ID)
	}
	castCatalogSpell(t, g, "Frantic Firebolt", "Instant", std54FireboltOracle, cardRefs(foe))
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, foe); got != 3 {
		t.Fatalf("damage %d, want 3", got)
	}
	for _, p := range g.Seats {
		want := before[p.ID]
		if p.ID != me.ID {
			want -= 3
		}
		if got := lifeOf(g, p.ID); got != want {
			t.Errorf("%s is at %d, want %d", p.Name, got, want)
		}
	}
}

func TestStd54ImodaneIgnoresDamageFromAnAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := std54Seats(g)
	pushCatalogPermanent(g, me.ID, "Imodane, the Pyrohammer", "Legendary Creature — Human Knight", std54ImodaneOracle, false)
	gaPut(g, me.ID, "Food Fight", "Enchantment", std54FoodFightOracle)
	rock := gaPut(g, me.ID, "Mind Stone", "Artifact", "")
	foe := b12Creature(g, opp.ID, "Foe", woeMPBearType, 9, 9)
	advanceToMain(t, g)
	life := lifeOf(g, opp.ID)
	if err := std54ActivateGranted(t, g, me, rock, game.TargetRef{Kind: game.TargetCard, ID: foe}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, foe); got != 2 {
		t.Fatalf("damage %d, want 2", got)
	}
	if got := lifeOf(g, opp.ID); got != life {
		t.Errorf("opponent at %d after an ability's damage, want %d", got, life)
	}
}

// --- Rotisserie Elemental ---------------------------------------------------------

func TestStd54RotisserieElementalSacrificesToExileAndPlayCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := std54Seats(g)
	elemental := pushCatalogPermanent(g, me.ID, "Rotisserie Elemental", "Creature — Elemental", std54RotisserieOracl, false)
	ids := samiLibraryTopFirst(me, game.Card{Name: "One"}, game.Card{Name: "Two"}, game.Card{Name: "Three"})
	advanceToMain(t, g)

	// First hit: decline, and it keeps its counter.
	dealCombatDamageToPlayer(g, elemental, opp.ID, 1)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, elemental)
	if !ok {
		t.Fatal("the Elemental left after the player declined")
	}
	if got := c.Counters["skewer"]; got != 1 {
		t.Fatalf("%d skewer counters, want 1", got)
	}
	if g.Exile.Contains(ids[0]) {
		t.Fatal("cards were exiled though the player declined")
	}

	// Second hit: sacrifice, and X is the two counters it had.
	dealCombatDamageToPlayer(g, elemental, opp.ID, 1)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(elemental) {
		t.Fatal("the Elemental was not sacrificed")
	}
	for i, id := range ids {
		if want := i < 2; g.Exile.Contains(id) != want {
			t.Errorf("library card %d exiled = %v, want %v (X = 2)", i, !want, want)
		}
	}
}

// --- Agatha's Champion ------------------------------------------------------------

func std54CastCreature(t *testing.T, g *game.Game, card game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card.InstanceID = uuid.New()
	card.Owner, card.Controller = me.ID, me.ID
	me.Hand.PushTop(card)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, params); err != nil {
		t.Fatalf("CastSpell %s: %v", card.Name, err)
	}
	return card.InstanceID
}

func std54Champion() game.Card {
	return game.Card{Name: "Agatha's Champion", TypeLine: "Creature — Human Knight", OracleID: std54ChampionOracle,
		ManaCost: "{4}{G}", Power: 4, Toughness: 3}
}

func TestStd54AgathasChampionFightsWhenBargained(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := std54Seats(g)
	token := std54Token(g, me.ID)
	foe := b12Creature(g, opp.ID, "Foe", woeMPBearType, 2, 2)
	champion := std54CastCreature(t, g, std54Champion(), game.CastSpellParams{
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	std54AnswerTarget(t, g, me.ID, foe)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(champion) {
		t.Fatal("the Champion did not resolve")
	}
	if g.Battlefield.Contains(foe) {
		t.Error("the foe survived the fight")
	}
	if got := damageOn(t, g, champion); got != 2 {
		t.Errorf("the Champion has %d damage, want 2", got)
	}
}

func TestStd54AgathasChampionDoesNothingWhenNotBargained(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := std54Seats(g)
	foe := b12Creature(g, opp.ID, "Foe", woeMPBearType, 2, 2)
	champion := std54CastCreature(t, g, std54Champion(), game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, g.Seats[g.Turn.ActiveSeat].ID) != nil {
		t.Fatal("an unbargained Champion asked for a target")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(champion) || !g.Battlefield.Contains(foe) {
		t.Fatal("something left the battlefield")
	}
	if got := damageOn(t, g, foe); got != 0 {
		t.Errorf("the foe has %d damage, want 0", got)
	}
}

// --- Brave the Wilds --------------------------------------------------------------

func std54BraveSetup(t *testing.T, g *game.Game, me *game.Player) (land, forest uuid.UUID) {
	t.Helper()
	land = gaPut(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	f := game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(f)
	return land, f.InstanceID
}

func TestStd54BraveTheWildsAnimatesALandWhenBargained(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := std54Seats(g)
	land, forest := std54BraveSetup(t, g, me)
	token := std54Token(g, me.ID)
	p1g2Cast(t, g, "Brave the Wilds", "Sorcery", std54BraveOracle, "{G}", game.CastSpellParams{
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token}, Targets: cardRefs(land),
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(forest) {
		t.Error("the basic land was not put into hand")
	}
	c := gaLayered(t, g, land)
	if !c.IsLand() || !c.IsCreature() || !c.HasSubtype("Elemental") {
		t.Errorf("the land is %q, want a land Elemental creature", c.TypeLine)
	}
	if p, tough := effectivePTOf(t, g, land); p != 3 || tough != 3 {
		t.Errorf("the land is %d/%d, want 3/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, land, "haste") {
		t.Error("the land has no haste")
	}
	// It lasts: past the end of the turn it is still a 3/3 creature.
	for i := 0; i < 12 && g.Turn.Step != game.StepCleanup; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if p, _ := effectivePTOf(t, g, land); p != 3 {
		t.Errorf("the animation ended with the turn: power %d", p)
	}
}

func TestStd54BraveTheWildsOnlySearchesWhenNotBargained(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := std54Seats(g)
	land, forest := std54BraveSetup(t, g, me)
	p1g2Cast(t, g, "Brave the Wilds", "Sorcery", std54BraveOracle, "{G}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(forest) {
		t.Error("the basic land was not put into hand")
	}
	if c := gaLayered(t, g, land); c.IsCreature() {
		t.Error("an unbargained cast animated a land")
	}
}

func TestStd54BraveTheWildsCantBeBargainedWithoutALand(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := std54Seats(g)
	token := std54Token(g, me.ID)
	err := g.CastSpell(me.ID, p1g2Hand(g, me, "Brave the Wilds", "Sorcery", std54BraveOracle, "{G}"), game.CastSpellParams{
		Strict: true, OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if err == nil {
		t.Error("a bargained Brave the Wilds was cast with no target land")
	}
}

// p1g2Hand puts a card in the active player's hand in a main phase and
// returns its ID.
func p1g2Hand(g *game.Game, p *game.Player, name, typeLine, oracle, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			return id
		}
	}
	return id
}

// --- Elvish Archivist -------------------------------------------------------------

func std54Enters(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine string) {
	t.Helper()
	id := b12Permanent(g, owner, name, typeLine)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: owner})
	})
	passPriorityAroundTable(t, g)
}

func TestStd54ElvishArchivistGrowsOncePerTurnForArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := std54Seats(g)
	archivist := pushCatalogPermanent(g, me.ID, "Elvish Archivist", "Creature — Elf Artificer", std54ArchivistOracle, false)
	std54Enters(t, g, opp.ID, "Their Rock", "Artifact")
	if c, _ := battlefieldCard(g, archivist); c.Counters[game.CounterPlusOne] != 0 {
		t.Fatal("an opponent's artifact grew the Archivist")
	}
	std54Enters(t, g, me.ID, "My Rock", "Artifact")
	if c, _ := battlefieldCard(g, archivist); c.Counters[game.CounterPlusOne] != 2 {
		t.Fatalf("%d counters after the first artifact, want 2", c.Counters[game.CounterPlusOne])
	}
	std54Enters(t, g, me.ID, "My Other Rock", "Artifact")
	if c, _ := battlefieldCard(g, archivist); c.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("%d counters after a second artifact this turn, want still 2", c.Counters[game.CounterPlusOne])
	}
}

func TestStd54ElvishArchivistDrawsOncePerTurnForEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := std54Seats(g)
	seedLibrary(me, "A", "B", "C", "D")
	pushCatalogPermanent(g, me.ID, "Elvish Archivist", "Creature — Elf Artificer", std54ArchivistOracle, false)
	hand := handSize(me)
	std54Enters(t, g, opp.ID, "Their Anthem", "Enchantment")
	if got := handSize(me); got != hand {
		t.Fatalf("hand %d after an opponent's enchantment, want %d", got, hand)
	}
	std54Enters(t, g, me.ID, "My Anthem", "Enchantment")
	if got := handSize(me); got != hand+1 {
		t.Fatalf("hand %d after your enchantment, want %d", got, hand+1)
	}
	std54Enters(t, g, me.ID, "My Other Anthem", "Enchantment")
	if got := handSize(me); got != hand+1 {
		t.Errorf("hand %d after a second enchantment this turn, want %d", got, hand+1)
	}
	std54Enters(t, g, me.ID, "My Rock", "Artifact")
	if got := handSize(me); got != hand+1 {
		t.Errorf("an artifact drew a card: hand %d", got)
	}
}
