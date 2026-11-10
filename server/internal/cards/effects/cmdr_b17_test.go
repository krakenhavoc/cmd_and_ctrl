package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b17_test.go — Commander deck cards, batch 17 (#5368).

const (
	cmdrB17ShalaiOracle     = "a0c47ab6-dfb4-46ee-a3f7-9e1521b4bb4b"
	cmdrB17DroverOracle     = "b0d60856-80f9-4c64-9768-28a0a690ecb1"
	cmdrB17GargoyleOracle   = "11284767-7095-4e0e-bc60-c69d58098348"
	cmdrB17ChaplainOracle   = "99e24041-910e-4ae2-a55b-b938d752eeff"
	cmdrB17MutineerOracle   = "8910f79a-66c9-41b3-b55a-e3e651ae6f26"
	cmdrB17BaralsOracle     = "a1cdcb8c-1e12-4d39-aa4a-db678a80df9c"
	cmdrB17LoreOracle       = "440881ad-48cf-4b82-8bb5-f506d1a51755"
	cmdrB17AlchemyOracle    = "66d65cda-e3d1-4a00-8ff3-c6d5ef0ab0c1"
	cmdrB17DjinnOracle      = "44036bb5-f300-40bd-9180-6b246315d886"
	cmdrB17BlightPileOracle = "1c0fc9b4-177e-4529-827f-a1e17ff7a752"
)

func TestCmdrB17ShalaiGivesHexproofToYouWalkersAndOthers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shalai := pushCatalogPermanent(g, me.ID, "Shalai, Voice of Plenty", "Legendary Creature — Angel", cmdrB17ShalaiOracle, false)
	mine := pushVanillaCreature(g, me.ID, "My Guy", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Guy", 2, 2)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: me.ID, Controller: me.ID,
	})
	if !hasPlayerAbility(playerAbilities(g, me), KeywordHexproof) {
		t.Error("you have hexproof")
	}
	if hasPlayerAbility(playerAbilities(g, opp), KeywordHexproof) {
		t.Error("opponents don't")
	}
	for _, c := range []struct {
		id   uuid.UUID
		name string
		want bool
	}{{mine, "other creature", true}, {walker, "planeswalker", true}, {theirs, "their creature", false}, {shalai, "Shalai herself", false}} {
		if got := eotHasAbility(effectiveAbilities(t, g, c.id), "hexproof"); got != c.want {
			t.Errorf("%s hexproof = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestCmdrB17ShalaiPutsACounterOnEachCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shalai := pushCatalogPermanent(g, me.ID, "Shalai, Voice of Plenty", "Legendary Creature — Angel", cmdrB17ShalaiOracle, false)
	mine := pushVanillaCreature(g, me.ID, "My Guy", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Guy", 2, 2)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "G"},
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, shalai, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{shalai: 1, mine: 1, theirs: 0} {
		if got := counterCount(g, id, "+1/+1"); got != want {
			t.Errorf("%s: %d counters, want %d", id, got, want)
		}
	}
}

func TestCmdrB17TwilightDroverGrowsWhenAnyCreatureTokenLeavesAndMakesSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drover := pushCatalogPermanent(g, me.ID, "Twilight Drover", "Creature — Spirit", cmdrB17DroverOracle, false)
	tok := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	nontoken := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	if err := g.DestroyPermanentForEffect(nontoken); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, drover, "+1/+1"); got != 0 {
		t.Fatalf("a nontoken creature leaving gave %d counters", got)
	}
	if err := g.DestroyPermanentForEffect(tok); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, drover, "+1/+1"); got != 1 {
		t.Fatalf("a creature token leaving gave %d counters, want 1", got)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, drover, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Spirit"); got != 2 {
		t.Fatalf("want two Spirit tokens, got %d", got)
	}
	// The counter was the cost, and the tokens leaving later feed it again.
	if got := counterCount(g, drover, "+1/+1"); got != 0 {
		t.Errorf("counter should be spent, have %d", got)
	}
}

func TestCmdrB17WakestoneGargoyleLetsDefendersAttackThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gargoyle := pushTestPermanent(g, me.ID, game.Card{Name: "Wakestone Gargoyle", OracleID: cmdrB17GargoyleOracle,
		TypeLine: "Creature — Gargoyle", Power: 3, Toughness: 4})
	wall := pushTestPermanent(g, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	theirs := pushTestPermanent(g, opp.ID, game.Card{Name: "Their Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	advanceToMainOf(t, g, 0)
	advanceTo(t, g.Clone(), game.StepDeclareAttackers)
	before := g.Clone()
	advanceTo(t, before, game.StepDeclareAttackers)
	wantDefenderRefused(t, before, gargoyle, opp.ID)
	wantDefenderRefused(t, before, wall, opp.ID)

	me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, gargoyle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	// A defender that arrives after the ability resolved is covered too.
	late := pushTestPermanent(g, me.ID, game.Card{Name: "Late Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{gargoyle, wall, late} {
		if err := declareResult(g, id, opp.ID); err != nil {
			t.Errorf("attack with %s: %v", id, err)
		}
	}
	if err := declareResult(g, theirs, me.ID); err == nil {
		t.Error("an opponent's defender gets no waiver")
	}
	if !hasEffectiveKeyword(t, g, wall, "defender") {
		t.Error("the keyword stays")
	}
	// Until end of turn only.
	advanceToDeclareAttackersOf(t, g, 1)
	advanceToDeclareAttackersOf(t, g, 0)
	wantDefenderRefused(t, g, wall, opp.ID)
}

func TestCmdrB17WingmantleChaplainCountsDefendersAndTriggersOnOthers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTestPermanent(g, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	castCatalogSpell(t, g, "Wingmantle Chaplain", "Creature — Human Cleric", cmdrB17ChaplainOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Bird"); got != 2 {
		t.Fatalf("Birds after the enters trigger = %d, want 2 (the Wall and the Chaplain)", got)
	}
	cmdrB17Enter(t, g, me, game.Card{Name: "Late Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Bird"); got != 3 {
		t.Fatalf("Birds after another defender entered = %d, want 3", got)
	}
	cmdrB17Enter(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Bird"); got != 3 {
		t.Fatalf("a creature without defender made a Bird: %d", got)
	}
}

func TestCmdrB17BlightPileDrainsByDefenders(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pile := pushTestPermanent(g, me.ID, game.Card{Name: "Blight Pile", OracleID: cmdrB17BlightPileOracle,
		TypeLine: "Creature — Phyrexian", Power: 3, Toughness: 3})
	pushTestPermanent(g, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	pushTestPermanent(g, opp.ID, game.Card{Name: "Their Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	before := opp.Life
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, pile, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("opponent lost %d, want 2 (Blight Pile has defender if the keyword is printed)", got)
	}
}

// cmdrB17Enter puts a card into a player's hand and moves it onto the
// battlefield, so its enters triggers fire as they would in play.
func cmdrB17Enter(t *testing.T, g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	c.InstanceID, c.Owner, c.Controller = uuid.New(), p.ID, p.ID
	p.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("enter %s: %v", c.Name, err)
	}
	return c.InstanceID
}

func TestCmdrB17AmphinMutineerExilesANonSalamanderAndGivesItsControllerAToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Victim", 5, 5)
	salamander := pushTestPermanent(g, opp.ID, game.Card{Name: "Lizard", TypeLine: "Creature — Salamander", Power: 2, Toughness: 2})
	castCatalogSpell(t, g, "Amphin Mutineer", "Creature — Salamander Pirate", cmdrB17MutineerOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target was asked for")
	}
	if err := g.ResolvePickTarget(pick.ID, pick.Chooser, game.TargetRef{Kind: game.TargetCard, ID: salamander}); err == nil {
		t.Fatal("a Salamander is not a legal target")
	}
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(victim) {
		t.Fatal("the target was not exiled")
	}
	if got := countBattlefieldNamed(g, opp.ID, "Salamander Warrior"); got != 1 {
		t.Fatalf("the exiled creature's controller has %d Salamander Warriors, want 1", got)
	}
	if got := countBattlefieldNamed(g, me.ID, "Salamander Warrior"); got != 0 {
		t.Errorf("the caster got %d tokens", got)
	}
	tok := findBattlefieldByName(g, "Salamander Warrior")
	if p, tough := effectivePower(t, g, tok), effectiveToughness(t, g, tok); p != 4 || tough != 3 {
		t.Errorf("token is %d/%d, want 4/3", p, tough)
	}
}

func TestCmdrB17AmphinMutineerOnlySalamandersMeansNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	salamander := pushTestPermanent(g, opp.ID, game.Card{Name: "Lizard", TypeLine: "Creature — Salamander", Power: 2, Toughness: 2})
	castCatalogSpell(t, g, "Amphin Mutineer", "Creature — Salamander Pirate", cmdrB17MutineerOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was offered with only Salamanders around")
	}
	if g.Exile.Contains(salamander) || countBattlefieldNamed(g, opp.ID, "Salamander Warrior") != 0 {
		t.Error("nothing should have happened")
	}
}

func TestCmdrB17AmphinMutineerEncoreMakesATokenCopyPerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	id, _ := activateFromGraveyard(t, g, "Amphin Mutineer", "Creature — Salamander Pirate",
		cmdrB17MutineerOracle, 3, 3, "{C}{C}{C}{C}{U}{U}", game.ActivateAbilityParams{})
	if !g.Exile.Contains(id) {
		t.Fatal("encore's cost exiles the card")
	}
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Amphin Mutineer" && c.IsToken() {
			n++
		}
	}
	if n != 3 {
		t.Fatalf("%d token copies, want one per opponent (3)", n)
	}
}

func TestCmdrB17BaralsExpertiseBouncesThreeAndOffersAFreeCastIncludingTheBounced(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushTestPermanent(g, me.ID, game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2})
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	rock := pushTestPermanent(g, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"})
	land := pushTestPermanent(g, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	big := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: big, Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{4}{R}", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Baral's Expertise", "Sorcery", cmdrB17BaralsOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetCard, ID: rock}})
	if err := castCatalogSpellErr(t, g, "Baral's Expertise", "Sorcery", cmdrB17BaralsOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: land}}); err == nil {
		t.Error("a land is not an artifact or creature")
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) || !opp.Hand.Contains(rock) {
		t.Fatal("the three targets did not return to their owners' hands")
	}
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, mine)
	if slices.Contains(offer.DiscardOptions, big) {
		t.Error("a spell with mana value 5 was offered")
	}
	if err := castThroughResolutionCast(t, g, me.ID, mine, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast of the returned creature: %v", err)
	}
}

func TestCmdrB17DraconicLoreCostsTwoLessWithADragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if got := cmdrB17Price(t, g, me.ID, cmdrB17LoreOracle, "Draconic Lore", "Instant", "{5}{U}"); got != 6 {
		t.Fatalf("without a Dragon the price is %d, want 6", got)
	}
	pushTestPermanent(g, me.ID, game.Card{Name: "Whelp", TypeLine: "Creature — Dragon", Power: 2, Toughness: 2})
	if got := cmdrB17Price(t, g, me.ID, cmdrB17LoreOracle, "Draconic Lore", "Instant", "{5}{U}"); got != 4 {
		t.Fatalf("with a Dragon the price is %d, want 4", got)
	}
	handBefore := me.Hand.Size()
	castCatalogSpell(t, g, "Draconic Lore", "Instant", cmdrB17LoreOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+3 {
		t.Errorf("hand %d, want %d (drew three)", got, handBefore+3)
	}
}

func TestCmdrB17ForbiddenAlchemyTakesOneOfFourAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedLibrary(me, "One", "Two", "Three", "Four", "Fifth")
	alch := castCatalogSpell(t, g, "Forbidden Alchemy", "Instant", cmdrB17AlchemyOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 4 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("want exactly one of the top four, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the chosen card is not in hand")
	}
	for _, id := range []uuid.UUID{ids[0], ids[2], ids[3]} {
		if !me.Graveyard.Contains(id) {
			t.Error("a card not chosen is not in the graveyard")
		}
	}
	if !me.Library.Contains(ids[4]) {
		t.Error("the fifth card left the library")
	}
	if !me.Graveyard.Contains(alch) {
		t.Fatal("Forbidden Alchemy is not in the graveyard")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{C}{C}{C}{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, alch, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	passPriorityAroundTable(t, g)
	again := latestChooseCardsFor(g, me.ID)
	if again == nil {
		t.Fatal("the flashback should look at four more cards")
	}
	if err := g.ResolveChooseCards(again.ID, me.ID, []uuid.UUID{again.ChooseCards[0]}); err != nil {
		t.Fatal(err)
	}
	if !g.Exile.Contains(alch) {
		t.Error("the flashed-back spell is exiled")
	}
}

func TestCmdrB17HaughtyDjinnPowerCountsInstantsAndSorceriesAndDiscounts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	djinn := pushTestPermanent(g, me.ID, game.Card{Name: "Haughty Djinn", OracleID: cmdrB17DjinnOracle,
		TypeLine: "Creature — Djinn", Power: 0, Toughness: 4})
	if p := effectivePower(t, g, djinn); p != 0 {
		t.Fatalf("empty graveyard: power %d, want 0", p)
	}
	cmdrB17ToGraveyard(t, g, me, game.Card{Name: "Bolt", TypeLine: "Instant"})
	cmdrB17ToGraveyard(t, g, me, game.Card{Name: "Wrath", TypeLine: "Sorcery"})
	cmdrB17ToGraveyard(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	cmdrB17ToGraveyard(t, g, opp, game.Card{Name: "Opp Bolt", TypeLine: "Instant"})
	if p := effectivePower(t, g, djinn); p != 2 {
		t.Errorf("power %d, want 2 (only your instants and sorceries)", p)
	}
	if got := b40PricedFor(t, g, me.ID, "Divination", "Sorcery", "{2}{U}"); got != 2 {
		t.Errorf("sorcery costs %d, want 2", got)
	}
	if got := b40PricedFor(t, g, me.ID, "Bear", "Creature — Bear", "{2}{U}"); got != 3 {
		t.Errorf("creature costs %d, want 3", got)
	}
}

// cmdrB17ToGraveyard moves a card from a hand to its owner's graveyard
// through the engine, so the layer cache hears about it.
func cmdrB17ToGraveyard(t *testing.T, g *game.Game, p *game.Player, c game.Card) {
	t.Helper()
	c.InstanceID, c.Owner, c.Controller = uuid.New(), p.ID, p.ID
	p.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: p.ID}, c.InstanceID); err != nil {
		t.Fatalf("to graveyard %s: %v", c.Name, err)
	}
}

// cmdrB17Price is the mana value a card's controller owes through the
// cost pipeline, with the card's oracle id so self modifiers apply.
func cmdrB17Price(t *testing.T, g *game.Game, caster uuid.UUID, oracle, name, typeLine, manaCost string) int {
	t.Helper()
	base, err := game.ParseCost(manaCost)
	if err != nil {
		t.Fatal(err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{
		Card: game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost,
			OracleID: oracle, Owner: caster, Controller: caster},
		Controller: caster, FromZone: game.ZoneHand,
	})
	if err != nil {
		t.Fatal(err)
	}
	return out.ManaValue()
}
