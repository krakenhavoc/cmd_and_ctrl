package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Star Trek batch 6 (#4973): each card's printed text, driven through the
// engine.

const (
	trekB6PrimeDirectiveOracle = "9f1d416c-60de-4796-a513-097d1d9a2305"
	trekB6DataOracle           = "010630c6-5d7b-43e0-90a9-3e3cef9a4e6b"
	trekB6ScanOracle           = "5dabb7c1-8937-42b0-87c6-1cc7af46aaae"
	trekB6SaboteurOracle       = "af03149b-9307-443c-aefa-69bda52607ba"
	trekB6HoloCrewOracle       = "0b74d7de-8b9e-4a75-870e-230536fa8d79"
	trekB6EnergizeOracle       = "275257ea-3c72-4ca9-b2fc-e50efa071bdb"
	trekB6ChekovOracle         = "7d7e232f-e4f9-41db-a2d9-cfe88721b7a9"
	trekB6EnterpriseOracle     = "947729a8-f8b0-4e08-9a87-6602af0e3c0a"
	trekB6ProbeOracle          = "6fc7401f-66fe-4d17-be2f-afcf892be740"
	trekB6GuidanceOracle       = "6c18f853-684b-42f6-8bc6-e86596131ab1"
	trekB6Bear                 = "Creature — Bear"
)

func TestTrekB6EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		trekB6PrimeDirectiveOracle: "Consider the Prime Directive",
		trekB6DataOracle:           "Data, New Life",
		trekB6ScanOracle:           "Deep Space Scan",
		trekB6SaboteurOracle:       "Dominion Saboteur",
		trekB6HoloCrewOracle:       "Emergency Holographic Crew",
		trekB6EnergizeOracle:       "Energize",
		trekB6ChekovOracle:         "Ensign Pavel Chekov",
		trekB6EnterpriseOracle:     "Enterprise NX-01",
		trekB6ProbeOracle:          "Federation Probe",
		trekB6GuidanceOracle:       "Guidance Failure",
	}
	for oracle, name := range want {
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

// trekB6ModeThenTarget answers the open mode pick with `mode` and, when
// target is set, the pick_target prompt after it, then settles the stack.
func trekB6ModeThenTarget(t *testing.T, g *game.Game, me uuid.UUID, mode int, target uuid.UUID) {
	t.Helper()
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick(%d): %v", mode, err)
	}
	if target != uuid.Nil {
		pickCard(t, g, me, target)
	}
	passPriorityAroundTable(t, g)
}

// trekB6SurveilAll puts every card on the open surveil prompt into the
// graveyard and reports how many there were.
func trekB6SurveilAll(t *testing.T, g *game.Game, me uuid.UUID) int {
	t.Helper()
	c := surveilChoiceFor(g, me)
	if c == nil {
		t.Fatalf("no surveil prompt: %+v", g.PendingChoices)
	}
	n := len(c.ScryCards)
	if err := g.ResolveSurveil(c.ID, me, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	return n
}

// --- Consider the Prime Directive ---

func TestTrekB6PrimeDirectiveInterfereCountersUnlessPaid(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	sorcery := handCardFull(me, "Some Sorcery", "Sorcery", "", "", nil)
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast sorcery: %v", err)
	}
	if err := rfCastModal(t, g, "Consider the Prime Directive", trekB6PrimeDirectiveOracle, []int{0}, rfCardTarget(sorcery)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passUntilTaxed(t, g, me.ID)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(sorcery) {
		t.Error("declining the {3} should counter the spell")
	}
}

func TestTrekB6PrimeDirectiveInterfereRefusesAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[1].ID, "Bear", trekB6Bear, 2, 2)
	if err := rfCastModal(t, g, "Consider the Prime Directive", trekB6PrimeDirectiveOracle, []int{0}, rfCardTarget(bear)); err == nil {
		t.Error("a creature on the battlefield was accepted as the target spell")
	}
}

func TestTrekB6PrimeDirectiveObserveTapsStunsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", trekB6Bear, 2, 2)
	seedLibrary(me, "Drawn")
	if err := rfCastModal(t, g, "Consider the Prime Directive", trekB6PrimeDirectiveOracle, []int{1}, rfCardTarget(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	c := mscM2Live(t, g, bear)
	if !c.Tapped || c.Counters[game.CounterStun] != 1 {
		t.Errorf("bear: tapped %v, stun %d; want tapped with one stun counter", c.Tapped, c.Counters[game.CounterStun])
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d after the draw", me.Hand.Size(), hand+1)
	}
}

func TestTrekB6PrimeDirectiveObserveWithNoTargetStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Drawn")
	if err := rfCastModal(t, g, "Consider the Prime Directive", trekB6PrimeDirectiveOracle, []int{1}, nil); err != nil {
		t.Fatalf("cast with no target: %v", err)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d after the draw", me.Hand.Size(), hand+1)
	}
}

// --- Data, New Life ---

func TestTrekB6DataGrowsOnlyOnItsControllersDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	data := b12Push(g, me.ID, "Data, New Life", "Legendary Artifact Creature — Android Scientist", trekB6DataOracle, 0, 1)
	seedLibrary(me, "A", "B")
	seedLibrary(opp, "C")
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterOn(g, data, game.CounterPlusOne); got != 2 {
		t.Fatalf("+1/+1 counters after two draws = %d, want 2", got)
	}
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("DrawCard opp: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterOn(g, data, game.CounterPlusOne); got != 2 {
		t.Errorf("an opponent's draw added a counter: %d", got)
	}
}

// --- Deep Space Scan ---

func TestTrekB6DeepSpaceScanDrawsPerLandNameAndLiftsTheMaximum(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	b12Permanent(g, g.Seats[1].ID, "Swamp", "Basic Land — Swamp")
	seedLibrary(me, "A", "B", "C", "D")
	castCatalogSpell(t, g, "Deep Space Scan", "Sorcery", trekB6ScanOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d: two land names (an opponent's land does not count)", me.Hand.Size(), hand+2)
	}
	if got := mhMax(g, me); got != -1 {
		t.Errorf("maximum hand size %d, want none (-1)", got)
	}
	if got := mhMax(g, g.Seats[1]); got != 7 {
		t.Errorf("an opponent's maximum hand size changed to %d", got)
	}
}

// --- Dominion Saboteur ---

func TestTrekB6SaboteurCopiesNonlegendaryWithTheCopiedCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hero := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Hero", TypeLine: "Legendary Creature — Human Soldier",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterPlusOne: 2, game.CounterStun: 1},
	})
	castCatalogSpell(t, g, "Dominion Saboteur", "Creature — Shapeshifter", trekB6SaboteurOracle, nil)
	resolveWithCopyChoice(t, g, hero)
	var copy game.Card
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Their Hero" {
			copy, found = layeredCard(t, g, c.InstanceID), true
		}
	}
	if !found {
		t.Fatal("no copy of the hero entered under the caster's control")
	}
	if copy.IsLegendary() {
		t.Error("the copy is legendary")
	}
	if copy.Counters[game.CounterPlusOne] != 2 || copy.Counters[game.CounterStun] != 1 {
		t.Errorf("copy's counters = %v, want two +1/+1 and one stun", copy.Counters)
	}
	if !g.Battlefield.Contains(hero) {
		t.Error("the legend rule took the original: the copy should not be legendary")
	}
}

func TestTrekB6SaboteurDeclinedIsAPlainThreeThree(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: trekB6Bear,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterPlusOne: 3},
	})
	id := castCatalogSpell(t, g, "Dominion Saboteur", "Creature — Shapeshifter", trekB6SaboteurOracle, nil)
	resolveWithCopyChoice(t, g, uuid.Nil)
	c := layeredCard(t, g, id)
	if c.Name != "Dominion Saboteur" || c.Counters[game.CounterPlusOne] != 0 || c.CurrentPower() != 3 {
		t.Errorf("declined: name %q counters %v, want a plain Dominion Saboteur", c.Name, c.Counters)
	}
}

// --- Emergency Holographic Crew ---

func TestTrekB6HoloCrewReturnsASpacecraftCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ship := pushGraveyardCardTyped(me, "Old Ship", "Artifact — Spacecraft")
	bear := pushGraveyardCardTyped(me, "Old Bear", trekB6Bear)
	castAndResolveCreature(t, g, "Emergency Holographic Crew", "Artifact Creature — Illusion Officer", trekB6HoloCrewOracle)
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no pick_target prompt for the Spacecraft card")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a non-Spacecraft card was a legal target")
	}
	pickCard(t, g, me.ID, ship)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ship) {
		t.Error("the Spacecraft card did not return to hand")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the other card left the graveyard")
	}
}

func TestTrekB6HoloCrewSurveilsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	castAndResolveCreature(t, g, "Emergency Holographic Crew", "Artifact Creature — Illusion Officer", trekB6HoloCrewOracle)
	trekB6ModeThenTarget(t, g, me.ID, 1, uuid.Nil)
	if n := trekB6SurveilAll(t, g, me.ID); n != 2 {
		t.Errorf("surveil looked at %d cards, want 2", n)
	}
}

// --- Energize ---

func TestTrekB6EnergizeBouncesAnOpponentsCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", trekB6Bear, 2, 2)
	if err := castCatalogSpellErr(t, g, "Energize", "Sorcery", trekB6EnergizeOracle, cardRefs(mine)); err == nil {
		t.Fatal("a creature the caster controls was a legal target")
	}
	theirs := b12Creature(g, opp.ID, "Theirs", trekB6Bear, 2, 2)
	castCatalogSpell(t, g, "Energize", "Sorcery", trekB6EnergizeOracle, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Error("the opponent's creature did not return to its owner's hand")
	}
}

func TestTrekB6EnergizeBeamMeUpFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mine := b12Creature(g, me.ID, "Mine", trekB6Bear, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", trekB6Bear, 2, 2)
	card := graveyardCardOf(g, me, "Energize", "Sorcery", "{1}{U}", trekB6EnergizeOracle)

	// Without a creature to return, the cost cannot be paid.
	floatMana(t, g, me, "{C}{C}{C}{U}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: trekB6BeamMeUpKey, Targets: cardRefs(theirs),
	}); err == nil {
		t.Fatal("Beam me up was cast without returning a creature")
	}
	// An opponent's creature is not one you control.
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: trekB6BeamMeUpKey, Targets: cardRefs(theirs),
		AltCostIDs: []uuid.UUID{theirs},
	}); err == nil {
		t.Fatal("Beam me up returned an opponent's creature as its cost")
	}
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: trekB6BeamMeUpKey, Targets: cardRefs(theirs),
		AltCostIDs: []uuid.UUID{mine},
	}); err != nil {
		t.Fatalf("Beam me up: %v", err)
	}
	if !me.Hand.Contains(mine) {
		t.Fatal("the returned creature is not in its owner's hand")
	}
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Error("the target did not return to its owner's hand")
	}
	if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("Energize after a Beam me up cast is in %v, want exile", z)
	}
}

// --- Ensign Pavel Chekov ---

func TestTrekB6ChekovExilesXFromEachOpponentAndLetsYouPlayAChosenLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	type top struct{ land, other uuid.UUID }
	tops := map[uuid.UUID]top{}
	for _, opp := range g.Seats[1:] {
		other := seedLibraryTop(opp, "Spell", "Sorcery", "{1}")
		land := seedLibraryTop(opp, "Wastes", "Basic Land — Wastes", "")
		tops[opp.ID] = top{land: land, other: other}
	}
	castAndResolveCreature(t, g, "Ensign Pavel Chekov", "Legendary Creature — Human Pilot", trekB6ChekovOracle)
	passPriorityAroundTable(t, g)
	for _, opp := range g.Seats[1:] {
		for _, id := range []uuid.UUID{tops[opp.ID].land, tops[opp.ID].other} {
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
				t.Fatalf("seat %s's top two were not both exiled (X = 2)", opp.ID)
			}
		}
	}
	var chosen []uuid.UUID
	for i := 0; i < 3; i++ {
		pick := chooseCardsChoiceFor(g, me.ID)
		if pick == nil {
			t.Fatalf("prompt %d of 3 is missing: %+v", i+1, g.PendingChoices)
		}
		if len(pick.ChooseCards) != 1 {
			t.Fatalf("prompt %d offers %d cards, want only that opponent's one land", i+1, len(pick.ChooseCards))
		}
		chosen = append(chosen, pick.ChooseCards[0])
		if err := g.ResolveChooseCards(pick.ID, me.ID, pick.ChooseCards); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	if pick := chooseCardsChoiceFor(g, me.ID); pick != nil {
		t.Fatal("a fourth prompt opened")
	}
	// The chosen land is playable from exile; the nonland card is not castable.
	advanceToMain(t, g)
	other := tops[g.Seats[1].ID].other
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("a card that was not chosen was castable from exile")
	}
	land := chosen[0]
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("play the chosen land from exile: %v", err)
	}
	if c, ok := battlefieldCard(g, land); !ok || c.Controller != me.ID {
		t.Error("the chosen land is not on the battlefield under Chekov's controller")
	}
}

// --- Enterprise NX-01 ---

func TestTrekB6EnterpriseStunsOnEntryAndFliesAtSix(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", trekB6Bear, 6, 6)
	theirs := b12Creature(g, opp.ID, "Theirs", trekB6Bear, 2, 2)
	ship := castHoldCreature(t, g, "Enterprise NX-01", "Legendary Artifact — Spacecraft", trekB6EnterpriseOracle, 4, 3)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("a creature the caster controls was a legal target")
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if c := mscM2Live(t, g, theirs); !c.Tapped || c.Counters[game.CounterStun] != 1 {
		t.Errorf("their creature: tapped %v, stun %d", c.Tapped, c.Counters[game.CounterStun])
	}
	if layeredCard(t, g, ship).IsCreature() {
		t.Fatal("unstationed, the Spacecraft is a creature")
	}
	station(t, g, me, ship, mine)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, ship)
	if !c.IsCreature() || c.CurrentPower() != 4 || c.CurrentToughness() != 3 ||
		!game.HasKeyword(&c, "flying") || !game.HasKeyword(&c, "vigilance") {
		t.Errorf("at 6+: creature %v %d/%d flying %v vigilance %v; want a 4/3 flier with vigilance",
			c.IsCreature(), c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "flying"), game.HasKeyword(&c, "vigilance"))
	}
}

// --- Federation Probe ---

func TestTrekB6FederationProbePumpsATarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", trekB6Bear, 2, 2)
	castAndResolveCreature(t, g, "Federation Probe", "Artifact Creature — Construct", trekB6ProbeOracle)
	trekB6ModeThenTarget(t, g, me.ID, 0, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("bear is %d/%d, want 3/3", p, tg)
	}
}

func TestTrekB6FederationProbeSurveilsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B")
	probe := castAndResolveCreature(t, g, "Federation Probe", "Artifact Creature — Construct", trekB6ProbeOracle)
	trekB6ModeThenTarget(t, g, me.ID, 1, uuid.Nil)
	if n := trekB6SurveilAll(t, g, me.ID); n != 1 {
		t.Errorf("surveil looked at %d cards, want 1", n)
	}
	if c := layeredCard(t, g, probe); !game.HasKeyword(&c, "flying") {
		t.Error("Federation Probe has no flying")
	}
}

// --- Guidance Failure ---

func TestTrekB6GuidanceFailureTucksASpacecraftByItsOwnersChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ship := b12Permanent(g, opp.ID, "Their Ship", "Artifact — Spacecraft")
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Guidance Failure", "Instant", trekB6GuidanceOracle, cardRefs(rock)); err == nil {
		t.Fatal("a noncreature, non-Spacecraft artifact was a legal target")
	}
	castCatalogSpell(t, g, "Guidance Failure", "Instant", trekB6GuidanceOracle, cardRefs(ship))
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
		t.Fatal("the caster was asked; the owner chooses")
	}
	c := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if c == nil {
		t.Fatalf("the owner was not asked: %+v", g.PendingChoices)
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, []uuid.UUID{ship}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if cards := opp.Library.Cards; len(cards) == 0 || cards[0].InstanceID != ship {
		t.Error("the Spacecraft is not on the bottom of its owner's library")
	}
}

func TestTrekB6GuidanceFailureCostsOneLessAtAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Attacker", trekB6Bear, 2, 2)
	idle := b12Creature(g, opp.ID, "Idle", trekB6Bear, 2, 2)
	advanceToMain(t, g)
	card := game.Card{Name: "Guidance Failure", TypeLine: "Instant", ManaCost: "{3}{U}", OracleID: trekB6GuidanceOracle}
	if err := socB19CastWithPool(g, me, card, "U", 3, game.CastSpellParams{Targets: cardRefs(idle)}); err == nil {
		t.Fatal("cast for {2}{U} at a creature that is not attacking")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	socB34AttackWith(t, g, opp.ID, attacker)
	card.InstanceID = uuid.New()
	if err := socB19CastWithPool(g, me, card, "U", 3, game.CastSpellParams{Targets: cardRefs(attacker)}); err != nil {
		t.Fatalf("cast for {2}{U} at an attacking creature: %v", err)
	}
}
