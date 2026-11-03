package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b16_cards_test.go — Secrets of Strixhaven Commander, sixteenth batch
// (#4834). Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

const (
	socB16AnalystOracle     = "977f9390-e01d-4e9d-8e9c-e543b1266972"
	socB16ProdigyOracle     = "ee2ef1e0-6803-4eb5-8664-605438c67505"
	socB16HypothesisOracle  = "ae27fa09-ab73-44b4-8e86-44d626ec221d"
	socB16NauseamOracle     = "981b0e21-e5e6-4a1e-bfde-679d56623f7f"
	socB16BringOracle       = "67183256-23cc-4af5-991e-d7148e942c63"
	socB16BrotherhoodOracle = "fe7bf8ca-88ae-4c6d-8d71-ef170a87c0d0"
	socB16DeduceOracle      = "eeae4161-9f35-45df-8b3c-c55e37f49cd1"
	socB16DutyOracle        = "67e6ce99-4883-4713-9822-cb2334765b5b"
	socB16KnockoutOracle    = "7e4931f4-f518-4200-aad4-4ac8acf202e5"
	socB16LivingEndOracle   = "3939fbe0-3be4-41ef-828d-c93f2cb33b4f"
)

func TestSocB16EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB16AnalystOracle:     "Zimone, Infinite Analyst",
		socB16ProdigyOracle:     "Zimone, Quandrix Prodigy",
		socB16HypothesisOracle:  "Zimone's Hypothesis",
		socB16NauseamOracle:     "Ad Nauseam",
		socB16BringOracle:       "Bring to Light",
		socB16BrotherhoodOracle: "Brotherhood's End",
		socB16DeduceOracle:      "Deduce",
		socB16DutyOracle:        "Duty Beyond Death",
		socB16KnockoutOracle:    "Knockout Maneuver",
		socB16LivingEndOracle:   "Living End",
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

// socB16TryCast puts a card in the active seat's hand and casts it, handing
// back the announce error and the card's id.
func socB16TryCast(t *testing.T, g *game.Game, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	me.Hand.PushTop(c)
	advanceToMain(t, g)
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, params)
}

// socB16Named lists the permanents named name on the battlefield.
func socB16Named(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// --- Deduce ---------------------------------------------------------------

func TestSocB16DeduceDrawsACardAndMakesAClue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Deduce", "Instant", socB16DeduceOracle, nil)
	passPriorityAroundTable(t, g)
	// The cast is a wash: the spell was pushed to hand and then cast.
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	clues := socB16Named(g, "Clue")
	if len(clues) != 1 || clues[0].Controller != me.ID || !clues[0].IsToken() || !clues[0].IsArtifact() {
		t.Errorf("Clues %+v, want one Clue artifact token you control", clues)
	}
}

// --- Brotherhood's End ----------------------------------------------------

func TestSocB16BrotherhoodsEndDamagesEachCreatureAndPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 5, 5)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	walker := b12Permanent(g, opp.ID, "Walker", "Legendary Planeswalker — Test")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == walker {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterLoyalty: 5}
			}
		}
	})
	castModal(t, g, "Brotherhood's End", "Sorcery", socB16BrotherhoodOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("your own 3/3 should die too")
	}
	if !g.Battlefield.Contains(big) || b12Card(t, g, big).DamageMarked != 3 {
		t.Error("the 5/5 should survive with 3 damage")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the damage mode does not touch artifacts")
	}
	if got := counterCount(g, walker, game.CounterLoyalty); got != 2 {
		t.Errorf("the planeswalker has %d loyalty, want 2", got)
	}
}

func TestSocB16BrotherhoodsEndDestroysArtifactsWithManaValueThreeOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := pushCostedPermanentForTest(g, me.ID, "Cheap Rock", "Artifact", "{3}")
	free := b12Permanent(g, opp.ID, "Free Rock", "Artifact")
	golem := pushCostedPermanentForTest(g, opp.ID, "Little Golem", "Artifact Creature — Golem", "{2}")
	dear := pushCostedPermanentForTest(g, opp.ID, "Dear Rock", "Artifact", "{4}")
	plain := pushCostedPermanentForTest(g, opp.ID, "Plain Bear", "Creature — Bear", "{1}")
	castModal(t, g, "Brotherhood's End", "Sorcery", socB16BrotherhoodOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"mana value 3": cheap, "no mana cost": free, "artifact creature": golem} {
		if g.Battlefield.Contains(id) {
			t.Errorf("the %s artifact survived", name)
		}
	}
	if !g.Battlefield.Contains(dear) {
		t.Error("the mana value 4 artifact should survive")
	}
	if !g.Battlefield.Contains(plain) || b12Card(t, g, plain).DamageMarked != 0 {
		t.Error("the destroy mode leaves a non-artifact creature alone")
	}
}

// --- Knockout Maneuver ----------------------------------------------------

func socB16Knockout(t *testing.T, g *game.Game, mine, theirs uuid.UUID, strict bool) (uuid.UUID, error) {
	t.Helper()
	return socB16TryCast(t, g, game.Card{Name: "Knockout Maneuver", TypeLine: "Sorcery", OracleID: socB16KnockoutOracle},
		game.CastSpellParams{Strict: strict, Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
		}})
}

func TestSocB16KnockoutManeuverCountersThenBitesForTheNewPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 6, 6)
	if _, err := socB16Knockout(t, g, mine, small, true); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
	if g.Battlefield.Contains(small) {
		t.Error("the counter makes it 3 power, so 3 damage kills the 3/3")
	}
	if _, err := socB16Knockout(t, g, mine, big, true); err != nil {
		t.Fatalf("second cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b12Card(t, g, mine).CurrentPower(); got != 4 {
		t.Errorf("power %d after two counters, want 4", got)
	}
	if got := b12Card(t, g, big).DamageMarked; got != 4 {
		t.Errorf("the 6/6 took %d, want 4 (power with the second counter)", got)
	}
}

func TestSocB16KnockoutManeuverRefusesTheWrongTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if _, err := socB16Knockout(t, g, mine, other, true); err == nil {
		t.Error("the second target is a creature an opponent controls: your own creature was accepted")
	}
	if _, err := socB16Knockout(t, g, theirs, theirs, true); err == nil {
		t.Error("the first target is a creature you control: an opponent's was accepted")
	}
}

func TestSocB16KnockoutManeuverGivesTheCounterEvenIfTheVictimIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if _, err := socB16Knockout(t, g, mine, theirs, true); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, theirs); err != nil {
		t.Fatalf("remove the victim: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("counters %d, want 1 (the spell still resolves for its legal target)", got)
	}
}

// --- Duty Beyond Death ----------------------------------------------------

func TestSocB16DutyBeyondDeathSacrificesThenProtectsAndGrowsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	keeper := b12Creature(g, me.ID, "Keeper", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	floatForTest(g, me, "WC")
	_, err := socB16TryCast(t, g, game.Card{Name: "Duty Beyond Death", TypeLine: "Instant", ManaCost: "{1}{W}", OracleID: socB16DutyOracle},
		game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the creature was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	k := b12Card(t, g, keeper)
	their := b12Card(t, g, theirs)
	if got := counterCount(g, keeper, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters on the other creature %d, want 1", got)
	}
	if !game.HasKeyword(&k, "indestructible") {
		t.Error("the other creature did not gain indestructible")
	}
	if game.HasKeyword(&their, "indestructible") || counterCount(g, theirs, game.CounterPlusOne) != 0 {
		t.Error("an opponent's creature got the effect")
	}
	// Lethal damage does not destroy it.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == keeper {
				g.Battlefield.Cards[i].DamageMarked = 9
			}
		}
	})
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(keeper) {
		t.Error("indestructible did not save the creature from lethal damage")
	}
}

func TestSocB16DutyBeyondDeathNeedsACreatureToSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	floatForTest(g, me, "WC")
	_, err := socB16TryCast(t, g, game.Card{Name: "Duty Beyond Death", TypeLine: "Instant", ManaCost: "{1}{W}", OracleID: socB16DutyOracle},
		game.CastSpellParams{Strict: true})
	if err == nil {
		t.Error("cast with no creature to sacrifice")
	}
}

// --- Ad Nauseam -----------------------------------------------------------

func socB16CastNauseam(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Ad Nauseam", "Instant", socB16NauseamOracle, nil)
	passPriorityAroundTable(t, g)
}

func socB16Repeat(t *testing.T, g *game.Game, who uuid.UUID, yes bool) {
	t.Helper()
	c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, who)
	if c == nil {
		t.Fatal("Ad Nauseam did not ask whether to repeat")
	}
	if err := g.ResolveConfirm(c.ID, who, yes); err != nil {
		t.Fatalf("ResolveConfirm(%v): %v", yes, err)
	}
	passPriorityAroundTable(t, g)
}

func TestSocB16AdNauseamRepeatsAsLongAsYouSayYes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibraryTop(me, "Zero", "Land", "")
	seedLibraryTop(me, "Three", "Sorcery", "{2}{R}")
	two := seedLibraryTop(me, "Two", "Creature — Bear", "{1}{G}")
	hand, life, lib := me.Hand.Size(), me.Life, me.Library.Size()
	socB16CastNauseam(t, g)
	// The cast is a wash for the hand: the spell was pushed there and cast.
	if !me.Hand.Contains(two) || me.Life != life-2 {
		t.Fatalf("after one pass: Two in hand %v, life %d (was %d), want it in hand and -2", me.Hand.Contains(two), me.Life, life)
	}
	socB16Repeat(t, g, me.ID, true)
	if me.Life != life-5 {
		t.Fatalf("after the second pass life %d, want %d", me.Life, life-5)
	}
	socB16Repeat(t, g, me.ID, true)
	if me.Life != life-5 {
		t.Errorf("a land costs no life: %d, want %d", me.Life, life-5)
	}
	socB16Repeat(t, g, me.ID, false)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("hand grew by %d, want 3", got)
	}
	if got := lib - me.Library.Size(); got != 3 {
		t.Errorf("library shrank by %d, want 3", got)
	}
	if c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID); c != nil {
		t.Error("a prompt is still open after saying no")
	}
}

func TestSocB16AdNauseamStopsAfterOnePassWhenDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibraryTop(me, "Next", "Creature — Bear", "{5}")
	seedLibraryTop(me, "Top", "Creature — Bear", "{3}")
	life := me.Life
	socB16CastNauseam(t, g)
	socB16Repeat(t, g, me.ID, false)
	if me.Life != life-3 {
		t.Errorf("life %d, want %d", me.Life, life-3)
	}
}

func TestSocB16AdNauseamRevealsRatherThanDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibraryTop(me, "Top", "Creature — Bear", "{1}")
	draws := func() int {
		n := 0
		for _, ev := range g.Events {
			if ev.Kind == game.EventDrawCard {
				n++
			}
		}
		return n
	}
	before := draws()
	socB16CastNauseam(t, g)
	socB16Repeat(t, g, me.ID, false)
	if got := draws() - before; got != 0 {
		t.Fatalf("Ad Nauseam drew %d cards; it reveals and puts them into hand", got)
	}
}

func TestSocB16AdNauseamOnAnEmptyLibraryAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	life := me.Life
	socB16CastNauseam(t, g)
	if c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID); c != nil {
		t.Error("asked to repeat with nothing left to reveal")
	}
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}

// --- Bring to Light -------------------------------------------------------

func socB16BringLibrary(me *game.Player) map[string]uuid.UUID {
	ids := seedSearchLibrary(me,
		game.Card{Name: "Basic Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Four Drop", TypeLine: "Creature — Giant", ManaCost: "{3}{G}", Power: 4, Toughness: 4},
		game.Card{Name: "Three Sorcery", TypeLine: "Sorcery", ManaCost: "{2}{R}"},
		game.Card{Name: "Two Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "One Instant", TypeLine: "Instant", ManaCost: "{U}"},
		game.Card{Name: "Three Artifact", TypeLine: "Artifact", ManaCost: "{3}"},
	)
	return map[string]uuid.UUID{"land": ids[0], "four": ids[1], "sorcery": ids[2], "bear": ids[3], "instant": ids[4], "artifact": ids[5]}
}

func socB16CastBring(t *testing.T, g *game.Game, pool string) {
	t.Helper()
	me := g.Seats[0]
	advanceToMain(t, g)
	floatForTest(g, me, pool)
	if _, err := socB16TryCast(t, g, game.Card{Name: "Bring to Light", TypeLine: "Sorcery", ManaCost: "{3}{G}{U}", OracleID: socB16BringOracle},
		game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestSocB16BringToLightFindsAtMostTheColorsSpentAndCastsItFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := socB16BringLibrary(me)
	socB16CastBring(t, g, "GUCCC") // two colors of mana
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.SearchCards {
		offered[id] = true
	}
	if !offered[ids["bear"]] || !offered[ids["instant"]] {
		t.Errorf("the mana value 2 creature and mana value 1 instant should be offered: %v", offered)
	}
	for _, name := range []string{"land", "four", "sorcery", "artifact"} {
		if offered[ids[name]] {
			t.Errorf("%s was offered: wrong type or mana value above two colors", name)
		}
	}
	answerSearchByID(t, g, me.ID, ids["bear"])
	if !g.Exile.Contains(ids["bear"]) {
		t.Fatal("the found card is not exiled")
	}
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, ids["bear"])
	if err := castThroughResolutionCast(t, g, me.ID, ids["bear"], game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids["bear"]) {
		t.Error("the free creature did not resolve onto the battlefield")
	}
}

func TestSocB16BringToLightWithFiveColorsFindsASorceryAndMayDeclineTheCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := socB16BringLibrary(me)
	socB16CastBring(t, g, "GUWBR")
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.SearchCards {
		offered[id] = true
	}
	for _, name := range []string{"four", "sorcery", "bear", "instant"} {
		if !offered[ids[name]] {
			t.Errorf("%s should be offered at five colors", name)
		}
	}
	if offered[ids["land"]] || offered[ids["artifact"]] {
		t.Error("a land or an artifact was offered")
	}
	answerSearchByID(t, g, me.ID, ids["sorcery"])
	declineResolutionCast(t, g, me.ID)
	if !g.Exile.Contains(ids["sorcery"]) {
		t.Error("a declined card should stay in exile")
	}
}

// --- Living End -----------------------------------------------------------

func TestSocB16LivingEndIsSuspendedAndSwapsTheCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	id := handCardForTest(me, "Living End", "Sorcery", socB16LivingEndOracle)
	stampLayout(g, me, id, "normal")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("Living End has no mana cost and was hard cast")
	}
	sa := suspendOf(socB16LivingEndOracle)
	if sa == nil || sa.Counters != 3 || sa.Cost != "{2}{B}{B}" {
		t.Fatalf("suspend declaration %+v, want 3 / {2}{B}{B}", sa)
	}
	floatForTest(g, me, "BBCC")
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend: %v", err)
	}
	if got := b12Card(t, g, id).Counters[game.CounterTime]; !g.Exile.Contains(id) || got != 3 {
		t.Fatalf("in exile %v with %d time counters, want exiled with 3", g.Exile.Contains(id), got)
	}

	b12Creature(g, me.ID, "Mine On Board", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Theirs On Board", "Creature — Bear", 2, 2)
	gyCard(me, "Mine In Yard", "Creature — Elf", "{1}", "")
	gyCard(opp, "Theirs In Yard", "Creature — Elf", "{1}", "")
	spell := gyCard(opp, "Dead Spell", "Sorcery", "{1}", "")
	crRemove(t, g, id, game.CounterTime, 2)

	advanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKind(g, game.PendingChoiceResolutionCast) == nil {
		t.Fatal("the last time counter coming off offered no cast")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, name := range []string{"Mine On Board", "Theirs On Board"} {
		if len(socB16Named(g, name)) != 0 {
			t.Errorf("%s survived the sacrifice", name)
		}
	}
	if len(socB16Named(g, "Mine In Yard")) != 1 || socB16Named(g, "Mine In Yard")[0].Controller != me.ID {
		t.Error("your graveyard creature did not return under your control")
	}
	if len(socB16Named(g, "Theirs In Yard")) != 1 || socB16Named(g, "Theirs In Yard")[0].Controller != opp.ID {
		t.Error("their graveyard creature did not return under their control")
	}
	// The creatures sacrificed went to the graveyard, not back.
	inYard := map[string]bool{}
	for _, p := range []*game.Player{me, opp} {
		for _, c := range p.Graveyard.Cards {
			inYard[c.Name] = true
		}
	}
	if !inYard["Mine On Board"] || !inYard["Theirs On Board"] || !inYard["Dead Spell"] || !opp.Graveyard.Contains(spell) {
		t.Errorf("graveyards %v: the sacrificed creatures and the sorcery should be there", inYard)
	}
}

// --- Zimone's Hypothesis --------------------------------------------------

func socB16Hypothesis(t *testing.T, g *game.Game, counterOn uuid.UUID, option int) {
	t.Helper()
	me := g.Seats[0]
	castCatalogSpell(t, g, "Zimone's Hypothesis", "Instant", socB16HypothesisOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no prompt for the optional counter")
	}
	var picks []uuid.UUID
	if counterOn != uuid.Nil {
		picks = []uuid.UUID{counterOn}
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, picks); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	o := latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if o == nil {
		t.Fatal("did not ask odd or even")
	}
	if err := g.ResolveOptionPick(o.ID, me.ID, option); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func socB16HypothesisTable(g *game.Game) map[string]uuid.UUID {
	me, opp := g.Seats[0], g.Seats[1]
	return map[string]uuid.UUID{
		"zero":  b12Creature(g, me.ID, "Zero", "Creature — Wall", 0, 4),
		"one":   b12Creature(g, opp.ID, "One", "Creature — Bear", 1, 1),
		"two":   b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2),
		"three": b12Creature(g, opp.ID, "Three", "Creature — Bear", 3, 3),
	}
}

func TestSocB16HypothesisOddReturnsOddPowerCreaturesOfEveryPlayer(t *testing.T) {
	g := newCatalogGame(t)
	ids := socB16HypothesisTable(g)
	socB16Hypothesis(t, g, uuid.Nil, 0)
	for name, wantOn := range map[string]bool{"zero": true, "one": false, "two": true, "three": false} {
		if g.Battlefield.Contains(ids[name]) != wantOn {
			t.Errorf("odd: %s on battlefield = %v, want %v", name, !wantOn, wantOn)
		}
	}
	if !g.Seats[1].Hand.Contains(ids["one"]) {
		t.Error("a returned creature goes to its owner's hand")
	}
}

func TestSocB16HypothesisEvenReturnsEvenPowerCreaturesIncludingZero(t *testing.T) {
	g := newCatalogGame(t)
	ids := socB16HypothesisTable(g)
	socB16Hypothesis(t, g, uuid.Nil, 1)
	for name, wantOn := range map[string]bool{"zero": false, "one": true, "two": false, "three": true} {
		if g.Battlefield.Contains(ids[name]) != wantOn {
			t.Errorf("even: %s on battlefield = %v, want %v", name, !wantOn, wantOn)
		}
	}
}

func TestSocB16HypothesisJudgesThePowerWithTheNewCounter(t *testing.T) {
	g := newCatalogGame(t)
	ids := socB16HypothesisTable(g)
	// The 2/2 becomes 3/3, so it is odd and the 3/3 and it both go.
	socB16Hypothesis(t, g, ids["two"], 0)
	if counterCount(g, ids["two"], game.CounterPlusOne) != 0 && g.Battlefield.Contains(ids["two"]) {
		t.Error("the countered creature should be odd power and returned")
	}
	for name, wantOn := range map[string]bool{"zero": true, "one": false, "two": false, "three": false} {
		if g.Battlefield.Contains(ids[name]) != wantOn {
			t.Errorf("%s on battlefield = %v, want %v", name, !wantOn, wantOn)
		}
	}
}

func TestSocB16HypothesisMayCounterAnOpponentsCreatureOrDecline(t *testing.T) {
	g := newCatalogGame(t)
	ids := socB16HypothesisTable(g)
	// Counter on their 1/1 makes it 2 power: even, so even returns it.
	socB16Hypothesis(t, g, ids["one"], 1)
	if g.Battlefield.Contains(ids["one"]) {
		t.Error("their 1/1 with a counter has power 2, which is even and goes back")
	}
}

// --- Zimone, Infinite Analyst ---------------------------------------------

func TestSocB16AnalystCutsTheFirstXSpellByItsCountersAndGrowsOnIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	zim := pushCatalogPermanent(g, me.ID, "Zimone, Infinite Analyst", "Legendary Creature — Human Wizard", socB16AnalystOracle, false)
	addCountersForTest(g, zim, game.CounterPlusOne, 2)
	advanceToMain(t, g)

	// A spell without {X} is not discounted.
	floatForTest(g, me, "U")
	if _, err := socB16TryCast(t, g, game.Card{Name: "Plain", TypeLine: "Sorcery", ManaCost: "{1}{U}"},
		game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a spell without {X} cost {1}{U} for {U}")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })

	// {X}{U} with X=1 is {1}{U}; two counters take {1} off twice, down to {U}.
	floatForTest(g, me, "U")
	if _, err := socB16TryCast(t, g, game.Card{Name: "First X", TypeLine: "Sorcery", ManaCost: "{X}{U}"},
		game.CastSpellParams{Strict: true, XValue: 1}); err != nil {
		t.Fatalf("the first X spell with two counters on Zimone: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, zim, game.CounterPlusOne); got != 4 {
		t.Fatalf("Zimone has %d counters, want 4 (two more for the first X spell)", got)
	}

	// The second one is neither discounted nor a trigger.
	floatForTest(g, me, "U")
	if _, err := socB16TryCast(t, g, game.Card{Name: "Second X", TypeLine: "Sorcery", ManaCost: "{X}{U}"},
		game.CastSpellParams{Strict: true, XValue: 1}); err == nil {
		t.Fatal("the second X spell this turn was discounted")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "CU")
	if _, err := socB16TryCast(t, g, game.Card{Name: "Second X", TypeLine: "Sorcery", ManaCost: "{X}{U}"},
		game.CastSpellParams{Strict: true, XValue: 1}); err != nil {
		t.Fatalf("the second X spell at full price: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, zim, game.CounterPlusOne); got != 4 {
		t.Errorf("Zimone has %d counters after the second X spell, want 4", got)
	}
}

func TestSocB16AnalystWithNoCountersGrowsWithoutAnyDiscount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	zim := pushCatalogPermanent(g, me.ID, "Zimone, Infinite Analyst", "Legendary Creature — Human Wizard", socB16AnalystOracle, false)
	advanceToMain(t, g)
	floatForTest(g, me, "U")
	if _, err := socB16TryCast(t, g, game.Card{Name: "First X", TypeLine: "Sorcery", ManaCost: "{X}{U}"},
		game.CastSpellParams{Strict: true, XValue: 1}); err == nil {
		t.Fatal("no counters, yet the X spell was discounted")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "CU")
	if _, err := socB16TryCast(t, g, game.Card{Name: "First X", TypeLine: "Sorcery", ManaCost: "{X}{U}"},
		game.CastSpellParams{Strict: true, XValue: 1}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, zim, game.CounterPlusOne); got != 2 {
		t.Errorf("Zimone has %d counters, want 2", got)
	}
}

func TestSocB16AnalystIgnoresAnOpponentsXSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	zim := pushCatalogPermanent(g, me.ID, "Zimone, Infinite Analyst", "Legendary Creature — Human Wizard", socB16AnalystOracle, false)
	other := g.Seats[1]
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	advanceToMain(t, g)
	id := uuid.New()
	other.Hand.PushTop(game.Card{InstanceID: id, Name: "Their X", TypeLine: "Sorcery", ManaCost: "{X}{U}", Owner: other.ID, Controller: other.ID})
	floatForTest(g, other, "CU")
	if err := g.CastSpell(other.ID, id, game.CastSpellParams{Strict: true, XValue: 1}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, zim, game.CounterPlusOne); got != 0 {
		t.Errorf("Zimone has %d counters after an opponent's X spell, want 0", got)
	}
}

// --- Zimone, Quandrix Prodigy ---------------------------------------------

func socB16Prodigy(t *testing.T, g *game.Game) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	zim := pushCatalogPermanent(g, me.ID, "Zimone, Quandrix Prodigy", "Legendary Creature — Human Wizard", socB16ProdigyOracle, false)
	advanceToMain(t, g)
	return me, zim
}

func TestSocB16ProdigyPutsALandFromHandOntoTheBattlefieldTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, zim := socB16Prodigy(t, g)
	forest := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	bear := handCardForTest(me, "Bear", "Creature — Bear", "")
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, zim, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt for the land")
	}
	if !hasID(c.ChooseCards, forest) || hasID(c.ChooseCards, bear) {
		t.Errorf("candidates %v: want the land and not the creature", c.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, forest)
	if !ok || !card.Tapped {
		t.Errorf("Forest on battlefield %v, tapped %v; want it there tapped", ok, card.Tapped)
	}
	if z, _ := battlefieldCard(g, zim); !z.Tapped {
		t.Error("Zimone should have tapped for the cost")
	}
}

func TestSocB16ProdigyLandAbilityNeedsTheManaAndMayBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me, zim := socB16Prodigy(t, g)
	forest := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	if err := g.ActivateCatalogAbility(me.ID, zim, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, zim, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) || !me.Hand.Contains(forest) {
		t.Error("declining should leave the land in hand")
	}
}

func TestSocB16ProdigyDrawsOneOrTwoByLandCount(t *testing.T) {
	for _, tc := range []struct{ lands, want int }{{7, 1}, {8, 2}} {
		g := newCatalogGame(t)
		me, zim := socB16Prodigy(t, g)
		for i := 0; i < tc.lands; i++ {
			pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
		}
		hand := me.Hand.Size()
		floatForTest(g, me, "CCC")
		if err := g.ActivateCatalogAbility(me.ID, zim, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
			t.Fatalf("%d lands: activated with three mana", tc.lands)
		}
		floatForTest(g, me, "C")
		if err := g.ActivateCatalogAbility(me.ID, zim, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("%d lands: activate: %v", tc.lands, err)
		}
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size() - hand; got != tc.want {
			t.Errorf("%d lands: drew %d, want %d", tc.lands, got, tc.want)
		}
	}
}
