package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tdm_omen_cards_test.go — Tarkir: Dragonstorm's omen cards (#4463,
// #5175). Each test plays the real catalog entries and asserts what the
// printed text says. Every Omen half that resolves is checked for
// CR 720.3d: the card is in its owner's library, not in exile and not in
// a graveyard. The engine-level lifecycle is internal/game/omen_test.go.

// tdmOmenCard builds an omen card the way deck.toGameCard does: layout
// `adventure`, the Dragon as face 0 and the Omen as face 1, then
// SetFace(0).
func tdmOmenCard(oracle string, dragon, omen game.Face) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		Layout:     game.LayoutAdventure,
		Faces:      []game.Face{dragon, omen},
	}
	c.SetFace(0)
	return c
}

func tdmDragon(name, cost string, power, toughness int) game.Face {
	return game.Face{Name: name, TypeLine: "Creature — Dragon", ManaCost: cost,
		Colors: game.ColorsInManaCost(cost), Power: power, Toughness: toughness}
}

func tdmOmen(name, typeLine, cost string) game.Face {
	return game.Face{Name: name, TypeLine: typeLine + " — Omen", ManaCost: cost,
		Colors: game.ColorsInManaCost(cost)}
}

var tdmOmenCards = map[string]game.Card{
	"Bloomvine Regent": tdmOmenCard(bloomvineRegentOracleID,
		tdmDragon("Bloomvine Regent", "{3}{G}{G}", 4, 5), tdmOmen("Claim Territory", "Sorcery", "{2}{G}")),
	"Dirgur Island Dragon": tdmOmenCard(dirgurIslandDragonOracleID,
		tdmDragon("Dirgur Island Dragon", "{5}{U}", 4, 4), tdmOmen("Skimming Strike", "Instant", "{1}{U}")),
	"Disruptive Stormbrood": tdmOmenCard(disruptiveStormbroodOracleID,
		tdmDragon("Disruptive Stormbrood", "{4}{G}", 3, 3), tdmOmen("Petty Revenge", "Sorcery", "{1}{B}")),
	"Feral Deathgorger": tdmOmenCard(feralDeathgorgerOracleID,
		tdmDragon("Feral Deathgorger", "{5}{B}", 3, 5), tdmOmen("Dusk Sight", "Sorcery", "{1}{B}")),
	"Marang River Regent": tdmOmenCard(marangRiverRegentOracleID,
		tdmDragon("Marang River Regent", "{4}{U}{U}", 6, 7), tdmOmen("Coil and Catch", "Instant", "{3}{U}")),
	"Purging Stormbrood": tdmOmenCard(purgingStormbroodOracleID,
		tdmDragon("Purging Stormbrood", "{4}{B}", 4, 4), tdmOmen("Absorb Essence", "Instant", "{1}{W}")),
	"Riling Dawnbreaker": tdmOmenCard(rilingDawnbreakerOracleID,
		tdmDragon("Riling Dawnbreaker", "{4}{W}", 3, 4), tdmOmen("Signaling Roar", "Sorcery", "{1}{W}")),
	"Runescale Stormbrood": tdmOmenCard(runescaleStormbroodOracleID,
		tdmDragon("Runescale Stormbrood", "{3}{R}", 2, 4), tdmOmen("Chilling Screech", "Instant", "{1}{U}")),
	"Sagu Wildling": tdmOmenCard(saguWildlingOracleID,
		tdmDragon("Sagu Wildling", "{4}{G}", 3, 3), tdmOmen("Roost Seek", "Sorcery", "{G}")),
	"Scavenger Regent": tdmOmenCard(scavengerRegentOracleID,
		tdmDragon("Scavenger Regent", "{3}{B}", 4, 4), tdmOmen("Exude Toxin", "Sorcery", "{X}{B}{B}")),
	"Stormshriek Feral": tdmOmenCard(stormshriekFeralOracleID,
		tdmDragon("Stormshriek Feral", "{4}{R}", 3, 3), tdmOmen("Flush Out", "Sorcery", "{1}{R}")),
	"Twinmaw Stormbrood": tdmOmenCard(twinmawStormbroodOracleID,
		tdmDragon("Twinmaw Stormbrood", "{5}{W}", 5, 4), tdmOmen("Charring Bite", "Sorcery", "{1}{R}")),
	"Whirlwing Stormbrood": tdmOmenCard(whirlwingStormbroodOracleID,
		tdmDragon("Whirlwing Stormbrood", "{4}{U}", 4, 3), tdmOmen("Dynamic Soar", "Sorcery", "{2}{G}")),
}

func TestTdmOmenEveryHalfIsRegisteredFull(t *testing.T) {
	for name, c := range tdmOmenCards {
		for face, key := range []string{c.OracleID, c.OracleID + "#1"} {
			spec, ok := Lookup(key)
			if !ok {
				t.Errorf("%s face %d (%s) is not registered", name, face, key)
				continue
			}
			if spec.Name != c.Faces[face].Name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
				t.Errorf("%s face %d: name %q, completeness %s, caveats %v",
					name, face, spec.Name, spec.Completeness, spec.Caveats)
			}
		}
	}
}

// tdmCastOmen puts the named omen card in the active seat's hand at a
// main phase and casts its Omen half with `params`.
func tdmCastOmen(t *testing.T, g *game.Game, name string, params game.CastSpellParams) (*game.Player, uuid.UUID) {
	t.Helper()
	params.Face = 1
	id := castFromHand(t, g, tdmOmenCards[name], params)
	return g.Seats[g.Turn.ActiveSeat], id
}

// tdmCastDragon casts the named omen card's Dragon half and resolves it
// and anything it triggers.
func tdmCastDragon(t *testing.T, g *game.Game, name string) (*game.Player, uuid.UUID) {
	t.Helper()
	id := castFromHand(t, g, tdmOmenCards[name], game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatalf("%s did not enter the battlefield", name)
	}
	return g.Seats[g.Turn.ActiveSeat], id
}

// tdmWantShuffledIn is CR 720.3d: a resolved Omen is in its owner's
// library and nowhere else.
func tdmWantShuffledIn(t *testing.T, g *game.Game, owner *game.Player, id uuid.UUID) {
	t.Helper()
	if !owner.Library.Contains(id) {
		t.Errorf("the resolved Omen is not in its owner's library (stack %v, graveyard %v, exile %v, hand %v)",
			g.Stack.Contains(id), owner.Graveyard.Contains(id), g.Exile.Contains(id), owner.Hand.Contains(id))
	}
	if owner.Graveyard.Contains(id) || g.Exile.Contains(id) {
		t.Error("CR 720.3d: the Omen went to a graveyard or to exile as well")
	}
}

// tdmAnswerTriggerTarget answers the active seat's open pick_target
// prompt for an enters trigger with `card`.
func tdmAnswerTriggerTarget(t *testing.T, g *game.Game, card uuid.UUID) {
	t.Helper()
	pickTriggerTarget(t, g, g.Seats[g.Turn.ActiveSeat].ID, card)
}

func tdmBasic(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Basic Land — " + name}
}

// --- Bloomvine Regent // Claim Territory -------------------------------------------

func TestTdmBloomvineRegentGainsThreeForItselfAndAnotherDragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Life
	tdmCastDragon(t, g, "Bloomvine Regent")
	if got := me.Life - before; got != 3 {
		t.Fatalf("life +%d as the Regent entered, want +3", got)
	}
	castFromHand(t, g, game.Card{Name: "Test Drake", TypeLine: "Creature — Dragon", ManaCost: "{1}", Power: 2, Toughness: 2},
		game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := me.Life - before; got != 6 {
		t.Errorf("life +%d after another Dragon entered, want +6", got)
	}
	castFromHand(t, g, game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 2, Toughness: 2},
		game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := me.Life - before; got != 6 {
		t.Errorf("life +%d after a non-Dragon entered, want still +6", got)
	}
}

func TestTdmClaimTerritoryFindsTwoForestsThenShufflesItselfIn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	one := pushLibraryCardForTest(me, tdmBasic("Forest"))
	two := pushLibraryCardForTest(me, tdmBasic("Forest"))
	pushLibraryCardForTest(me, tdmBasic("Mountain"))
	_, id := tdmCastOmen(t, g, "Bloomvine Regent", game.CastSpellParams{})
	passPriorityAroundTable(t, g)

	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt for the first Forest")
	}
	if searchOptionNamed(g, c, "Mountain") != uuid.Nil {
		t.Error("Claim Territory offered a basic Mountain")
	}
	answerSearchByID(t, g, me.ID, one)
	answerSearchByID(t, g, me.ID, two)
	passPriorityAroundTable(t, g)

	land, ok := lookupBattlefield(g, one)
	if !ok || !land.Tapped {
		t.Errorf("the first Forest is not on the battlefield tapped (on battlefield %v)", ok)
	}
	if !me.Hand.Contains(two) {
		t.Error("the second Forest is not in hand")
	}
	tdmWantShuffledIn(t, g, me, id)
}

// --- Dirgur Island Dragon // Skimming Strike ----------------------------------------

func TestTdmSkimmingStrikeTapsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	_, id := tdmCastOmen(t, g, "Dirgur Island Dragon", game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	passPriorityAroundTable(t, g)
	if c, _ := lookupBattlefield(g, bear); !c.Tapped {
		t.Error("the target creature is not tapped")
	}
	// One card drawn: the Omen joined the hand and left it after the count.
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want +1 (one card drawn)", got)
	}
	tdmWantShuffledIn(t, g, me, id)
}

// --- Disruptive Stormbrood // Petty Revenge -----------------------------------------

func TestTdmDisruptiveStormbroodDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	relic := b12Permanent(g, opp.ID, "Test Relic", "Artifact")
	castFromHand(t, g, tdmOmenCards["Disruptive Stormbrood"], game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	tdmAnswerTriggerTarget(t, g, relic)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact survived the enters trigger")
	}
}

func TestTdmPettyRevengeDestroysOnlyPowerThreeOrLess(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	card := tdmOmenCards["Disruptive Stormbrood"]
	card.InstanceID, card.Owner, card.Controller = uuid.New(), me.ID, me.ID
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}}); err == nil {
		t.Error("Petty Revenge targeted a creature with power 4")
	}
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}}}); err != nil {
		t.Fatalf("cast Petty Revenge: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the power-3 creature survived")
	}
	tdmWantShuffledIn(t, g, me, card.InstanceID)
}

// --- Feral Deathgorger // Dusk Sight ------------------------------------------------

func TestTdmDuskSightPutsACounterAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	_, id := tdmCastOmen(t, g, "Feral Deathgorger", game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	passPriorityAroundTable(t, g)
	if c, _ := lookupBattlefield(g, bear); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the bear has %d +1/+1 counters, want 1", c.Counters[game.CounterPlusOne])
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want +1 (one card drawn)", got)
	}
	tdmWantShuffledIn(t, g, me, id)
}

func TestTdmFeralDeathgorgerKeywords(t *testing.T) {
	spec, _ := Lookup(feralDeathgorgerOracleID)
	if len(spec.PrintedKeywords) != 2 || len(spec.Triggered) != 1 {
		t.Errorf("Feral Deathgorger: keywords %v, %d triggers; want flying, deathtouch and the exile trigger",
			spec.PrintedKeywords, len(spec.Triggered))
	}
}

// --- Marang River Regent // Coil and Catch ------------------------------------------

func TestTdmCoilAndCatchDrawsThreeThenDiscards(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	_, id := tdmCastOmen(t, g, "Marang River Regent", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := discardOwed(g, me.ID); got != 1 {
		t.Fatalf("the discard prompt asks for %d cards, want 1", got)
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	// +3, -1: the Omen joined the hand and left it after the count.
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand changed by %d, want +2", got)
	}
	tdmWantShuffledIn(t, g, me, id)
}

// --- Purging Stormbrood // Absorb Essence -------------------------------------------

func TestTdmAbsorbEssencePumpsAndGrants(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	_, id := tdmCastOmen(t, g, "Purging Stormbrood", game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	passPriorityAroundTable(t, g)
	c, _ := lookupBattlefield(g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("bear is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if !game.HasKeyword(&c, "lifelink") || !game.HasKeyword(&c, "hexproof") {
		t.Error("the bear did not gain lifelink and hexproof")
	}
	tdmWantShuffledIn(t, g, me, id)
}

func TestTdmPurgingStormbroodRemovesEveryCounter(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 2, "stun": 1}
			}
		}
	})
	castFromHand(t, g, tdmOmenCards["Purging Stormbrood"], game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	tdmAnswerTriggerTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	c, _ := lookupBattlefield(g, bear)
	for kind, n := range c.Counters {
		if n != 0 {
			t.Errorf("the bear still has %d %s counters", n, kind)
		}
	}
}

// --- Riling Dawnbreaker // Signaling Roar -------------------------------------------

func TestTdmSignalingRoarCreatesAWhiteSoldier(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	before := len(g.Battlefield.Cards)
	me, id := tdmCastOmen(t, g, "Riling Dawnbreaker", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if len(g.Battlefield.Cards) != before+1 {
		t.Fatalf("battlefield %d cards, want %d", len(g.Battlefield.Cards), before+1)
	}
	tok := g.Battlefield.Cards[len(g.Battlefield.Cards)-1]
	if !tok.IsToken() || tok.Power != 2 || tok.Toughness != 2 || len(tok.Colors) != 1 || tok.Colors[0] != "W" ||
		!tok.HasSubtype("Soldier") || tok.Controller != me.ID {
		t.Errorf("token = %q %d/%d %v, want a 2/2 white Soldier", tok.TypeLine, tok.Power, tok.Toughness, tok.Colors)
	}
	tdmWantShuffledIn(t, g, me, id)
}

// --- Runescale Stormbrood // Chilling Screech ---------------------------------------

func TestTdmRunescaleStormbroodGrowsOnANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	_, id := tdmCastDragon(t, g, "Runescale Stormbrood")
	castFromHand(t, g, game.Card{Name: "Test Ritual", TypeLine: "Sorcery", ManaCost: "{0}"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if c, _ := lookupBattlefield(g, id); c.CurrentPower() != 4 {
		t.Errorf("Stormbrood power %d after a noncreature spell, want 4", c.CurrentPower())
	}
	castFromHand(t, g, game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 2, Toughness: 2},
		game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if c, _ := lookupBattlefield(g, id); c.CurrentPower() != 4 {
		t.Errorf("Stormbrood power %d after a non-Dragon creature spell, want still 4", c.CurrentPower())
	}
}

// Chilling Screech counters an Omen half (Charring Bite, mana value 2),
// and the countered Omen goes to its owner's graveyard: CR 720.3d is
// about a spell that resolves.
func TestTdmChillingScreechCountersAnOmenIntoTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	me, bite := tdmCastOmen(t, g, "Twinmaw Stormbrood", game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})

	screech := tdmOmenCards["Runescale Stormbrood"]
	screech.InstanceID, screech.Owner, screech.Controller = uuid.New(), me.ID, me.ID
	me.Hand.PushTop(screech)
	if err := g.CastSpell(me.ID, screech.InstanceID, game.CastSpellParams{Face: 1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bite}}}); err != nil {
		t.Fatalf("cast Chilling Screech: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bite) || me.Library.Contains(bite) {
		t.Errorf("the countered Omen: graveyard %v, library %v; want the graveyard",
			me.Graveyard.Contains(bite), me.Library.Contains(bite))
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the countered Charring Bite still dealt its damage")
	}
	tdmWantShuffledIn(t, g, me, screech.InstanceID)
}

// --- Sagu Wildling // Roost Seek ----------------------------------------------------

func TestTdmSaguWildlingGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Life
	tdmCastDragon(t, g, "Sagu Wildling")
	if got := me.Life - before; got != 3 {
		t.Errorf("life +%d, want +3", got)
	}
}

func TestTdmRoostSeekFindsABasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	island := pushLibraryCardForTest(me, tdmBasic("Island"))
	libBefore := me.Library.Size()
	_, id := tdmCastOmen(t, g, "Sagu Wildling", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, island)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(island) {
		t.Error("the basic land is not in hand")
	}
	// -1 (the land), +1 (the Omen).
	if got := me.Library.Size(); got != libBefore {
		t.Errorf("library %d cards, want %d", got, libBefore)
	}
	tdmWantShuffledIn(t, g, me, id)
}

// --- Scavenger Regent // Exude Toxin ------------------------------------------------

func TestTdmExudeToxinSparesDragons(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	drake := b12Creature(g, opp.ID, "Drake", "Creature — Dragon", 3, 3)
	advanceToMain(t, g)
	me, id := tdmCastOmen(t, g, "Scavenger Regent", game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)
	if c, _ := lookupBattlefield(g, bear); c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("bear is %d/%d, want 1/1", c.CurrentPower(), c.CurrentToughness())
	}
	if c, _ := lookupBattlefield(g, drake); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("the Dragon is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	tdmWantShuffledIn(t, g, me, id)
}

// --- Stormshriek Feral // Flush Out -------------------------------------------------

func TestTdmStormshriekFeralPumps(t *testing.T) {
	g := newCatalogGame(t)
	me, id := tdmCastDragon(t, g, "Stormshriek Feral")
	fillPoolColored(me, "R", 2)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := lookupBattlefield(g, id); c.CurrentPower() != 4 {
		t.Errorf("power %d, want 4", c.CurrentPower())
	}
}

func TestTdmFlushOutDiscardsThenDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	_, id := tdmCastOmen(t, g, "Stormshriek Feral", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	// -1, +2: the Omen joined the hand and left it after the count.
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want +1", got)
	}
	tdmWantShuffledIn(t, g, me, id)
}

// --- Twinmaw Stormbrood // Charring Bite --------------------------------------------

func TestTdmTwinmawStormbroodGainsFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Life
	tdmCastDragon(t, g, "Twinmaw Stormbrood")
	if got := me.Life - before; got != 5 {
		t.Errorf("life +%d, want +5", got)
	}
}

func TestTdmCharringBiteHitsOnlyACreatureWithoutFlying(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	giant := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 5, 5)
	bird := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	card := tdmOmenCards["Twinmaw Stormbrood"]
	card.InstanceID, card.Owner, card.Controller = uuid.New(), me.ID, me.ID
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bird}}}); err == nil {
		t.Error("Charring Bite targeted a creature with flying")
	}
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: giant}}}); err != nil {
		t.Fatalf("cast Charring Bite: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(giant) {
		t.Error("the 5/5 survived 5 damage")
	}
	tdmWantShuffledIn(t, g, me, card.InstanceID)
	if c, ok := cardInZone(me.Library, card.InstanceID); !ok || c.ManaValue() != 6 || c.Name != "Twinmaw Stormbrood" {
		t.Errorf("in the library the card should be the {5}{W} Dragon (CR 720.4): %q, mana value %d", c.Name, c.ManaValue())
	}
}

// --- Whirlwing Stormbrood // Dynamic Soar -------------------------------------------

func TestTdmWhirlwingStormbroodLetsSorceriesAndDragonsBeCastAtFlashSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Whirlwing Stormbrood",
		TypeLine: "Creature — Dragon", OracleID: whirlwingStormbroodOracleID, Power: 4, Toughness: 3,
		Owner: me.ID, Controller: me.ID})
	// Still the draw step: no sorcery-speed window is open.
	sorcery := game.Card{InstanceID: uuid.New(), Name: "Test Sorcery", TypeLine: "Sorcery", ManaCost: "{0}", Owner: me.ID, Controller: me.ID}
	drake := game.Card{InstanceID: uuid.New(), Name: "Test Drake", TypeLine: "Creature — Dragon", ManaCost: "{0}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	bear := game.Card{InstanceID: uuid.New(), Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{0}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(sorcery)
	me.Hand.PushTop(drake)
	me.Hand.PushTop(bear)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a non-Dragon creature was cast at flash speed")
	}
	if err := g.CastSpell(me.ID, sorcery.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a sorcery at flash speed: %v", err)
	}
	if err := g.CastSpell(me.ID, drake.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a Dragon spell at flash speed: %v", err)
	}
}

func TestTdmDynamicSoarPutsThreeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	_, id := tdmCastOmen(t, g, "Whirlwing Stormbrood", game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	passPriorityAroundTable(t, g)
	if c, _ := lookupBattlefield(g, bear); c.Counters[game.CounterPlusOne] != 3 {
		t.Errorf("the bear has %d +1/+1 counters, want 3", c.Counters[game.CounterPlusOne])
	}
	tdmWantShuffledIn(t, g, me, id)
}
