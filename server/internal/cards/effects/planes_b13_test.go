package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b13_test.go — Planechase plane cards, batches 13–15 (ADR 0216
// slice 4, #4870): the gold-bordered Unknown Event planes. Every test plays
// the real catalog entry face up on a Planechase table (planes_harness_test.go).

var planeTarnation = planeFixture{"08081904-4115-4afd-ac82-f00421e54038", "Tarnation", "Plane — Thunder Junction"}

func TestTarnationLetsTheCriminalDrawAndChaosPings(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTarnation)
	g, plane := newPlaneGame(t, planeTarnation)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	foe := opponentOf(g)
	before := me.Hand.Size()
	// Targeting an opponent is a crime (CR 700.13).
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: foe.ID}})
	if n := itemsFromPlane(g, plane.InstanceID); n != 1 {
		t.Fatalf("the crime put %d plane items on the way, want 1", n)
	}
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	// The bolt left the hand, one card was drawn.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d after casting a bolt (a card put in it) and drawing one, want %d", got, before+1)
	}

	life := foe.Life
	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatalf("a blank triggered the plane")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPlayerTarget(t, g, foe.ID)
	// Targeting an opponent is itself a crime: Tarnation asks the planar
	// controller again, who draws on the way.
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if foe.Life != life-1 {
		t.Errorf("foe life %d, want %d after 1 damage", foe.Life, life-1)
	}

	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	rollChaos(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Error("a face-down plane's chaos ability triggered")
	}
}

func answerPlayerTarget(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: who}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

var planeSkyDeck = planeFixture{"8baa42cf-32bc-4d1f-b4bd-02a4fce376c4", "Sky Deck", "Plane — Chicago"}

func TestSkyDeckPutsGlassCountersAndDestroysThemOnTheWayOut(t *testing.T) {
	requirePlaneRegisteredFull(t, planeSkyDeck)
	g, plane := newPlaneGame(t, planeSkyDeck)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	foe := opponentOf(g)
	mine := pushBattlefieldCreature(g, me.ID)
	theirs := pushBattlefieldCreature(g, foe.ID)
	bystander := pushBattlefieldCreature(g, foe.ID)
	g.BumpLayerVersionForTest()

	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatalf("a blank triggered the plane")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, theirs); !ok || c.Counters[planesB13GlassCounter] != 1 {
		t.Fatalf("the chosen permanent has %v glass counters, want 1", c.Counters)
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, mine)
	passPriorityAroundTable(t, g)

	// Walking away destroys every permanent with a glass counter.
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("the opponent's glass creature survived the planeswalk away")
	}
	if _, ok := battlefieldCard(g, mine); ok {
		t.Error("the planar controller's glass creature survived")
	}
	if _, ok := battlefieldCard(g, bystander); !ok {
		t.Error("a creature with no glass counter was destroyed")
	}
}

var planeTheBean = planeFixture{"b977ff66-6c52-4f33-8f25-e310d835be8b", "The Bean", "Plane — Chicago"}

// reflectionsOf lists the battlefield cards controlled by `seat` that are
// Reflections.
func reflectionsOf(g *game.Game, seat uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == seat && c.HasSubtype("Reflection") {
			out = append(out, c)
		}
	}
	return out
}

func TestTheBeanCopiesEachPlayersCreatureAsAShadowReflection(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTheBean)
	g, _ := newPlaneGame(t, planeTheBean)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats {
		mine[p.ID] = pushBattlefieldCreature(g, p.ID)
		pushBattlefieldCreature(g, p.ID) // a second choice each
	}
	g.BumpLayerVersionForTest()
	planeswalkToPlane(t, g, planeTheBean.Oracle)
	passPriorityAroundTable(t, g)
	for i := range g.Seats {
		seat := g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)]
		answerChooseCardsOrPick(t, g, seat.ID, mine[seat.ID])
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		refl := reflectionsOf(g, p.ID)
		if len(refl) != 1 || !refl[0].IsToken() {
			t.Fatalf("%s has %d Reflection tokens, want 1", p.Name, len(refl))
		}
		if !refl[0].IsCreature() || !refl[0].HasSubtype("Test") {
			t.Errorf("the copy lost its other types: %q", refl[0].TypeLine)
		}
		if !p1g2HasKeyword(g, refl[0].InstanceID, "shadow") {
			t.Errorf("%s's Reflection has no shadow", p.Name)
		}
	}
	// A blank does nothing; chaos copies one of the planar controller's.
	before := len(reflectionsOf(g, me.ID))
	rollBlank(t, g, me.ID)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerChooseCardsOrPick(t, g, me.ID, mine[me.ID])
	passPriorityAroundTable(t, g)
	if got := len(reflectionsOf(g, me.ID)); got != before+1 {
		t.Errorf("planar controller has %d Reflections after chaos, want %d", got, before+1)
	}
	for _, p := range g.Seats[1:] {
		if len(reflectionsOf(g, p.ID)) != 1 {
			t.Errorf("chaos gave %s a Reflection", p.Name)
		}
	}
}

var planeOteclan = planeFixture{"24b2ba5b-c1cb-4c73-89bf-94d47f468622", "Oteclán", "Plane — Ixalan"}

func TestOteclanMakesChaosEnsueOnEntryAndAtUpkeepAndDiscovers(t *testing.T) {
	requirePlaneRegisteredFull(t, planeOteclan)
	g, plane := newPlaneGame(t, planeOteclan)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	hit := discoverLibraryCard(me.ID, "Opt", "Instant", "{U}")
	pushLibraryCardForTest(me, hit)

	planeswalkToPlane(t, g, planeOteclan.Oracle)
	// Chaos ensues (no die was rolled): the chaos ability follows.
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.CastPermissionOnCardByIDForEffect(hit.InstanceID) == nil {
		t.Fatalf("chaos ensuing on entry did not discover 3 the top card")
	}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventRollDie {
			t.Fatalf("a die was rolled")
		}
	}
	answerMayCastPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	// A blank does nothing.
	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatalf("a blank triggered the plane")
	}

	// The next player's upkeep: chaos ensues again, and they discover.
	next := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	next.Library.Cards = nil
	top := discoverLibraryCard(next.ID, "Shock", "Instant", "{R}")
	pushLibraryCardForTest(next, top)
	toNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.CastPermissionOnCardByIDForEffect(top.InstanceID) == nil {
		t.Fatalf("chaos ensuing at the planar controller's upkeep did not discover for them")
	}
}

var planeRaiders = planeFixture{"33ed1120-82b1-4936-b210-2f0323a9401c", "Raiders' Allegiance", "Plane — Ixalan"}

func TestRaidersAllegianceRaidsTreasuresAndProliferates(t *testing.T) {
	requirePlaneRegisteredFull(t, planeRaiders)
	g, _ := newPlaneGame(t, planeRaiders)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, seat := range []int{(g.Turn.ActiveSeat + 1) % 4, (g.Turn.ActiveSeat + 2) % 4, (g.Turn.ActiveSeat + 2) % 4} {
		bear := pushBattlefieldCreature(g, me.ID)
		if err := g.DeclareAttacker(bear, g.Seats[seat].ID); err != nil {
			t.Fatal(err)
		}
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	pirates := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.HasSubtype("Pirate") {
			pirates++
		}
	}
	if pirates != 2 || me.Counters[planesB13PointCounter] != 2 {
		t.Fatalf("%d Pirates and %d point counters after two players were dealt combat damage, want 2 and 2", pirates, me.Counters[planesB13PointCounter])
	}

	// Chaos proliferates once per Pirate: the first ask is open.
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) == 0 {
		t.Error("chaos with two Pirates asked no proliferate")
	}
	for len(g.PendingChoices) > 0 {
		c := g.PendingChoices[0]
		if err := g.ResolveProliferate(c.ID, c.Chooser, nil); err != nil {
			t.Fatalf("ResolveProliferate: %v", err)
		}
		passPriorityAroundTable(t, g)
	}

	// Walking away: the planar controller has the most point counters.
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Treasure"); n != 3 {
		t.Errorf("%d Treasures, want 3 for the one player with the most point counters", n)
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d, want %d", got, hand+3)
	}
	for _, p := range g.Seats {
		if p.Counters[planesB13PointCounter] != 0 {
			t.Errorf("%s kept %d point counters", p.Name, p.Counters[planesB13PointCounter])
		}
	}
}

var planeWindyCity = planeFixture{"04491781-0e6a-4a30-a69f-db226702adfc", "The Windy City", "Plane — Chicago"}

func TestTheWindyCityGivesFlyingExilesFlyersAndCleansUp(t *testing.T) {
	requirePlaneRegisteredFull(t, planeWindyCity)
	g, plane := newPlaneGame(t, planeWindyCity)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	foe := opponentOf(g)
	mine := pushBattlefieldCreature(g, me.ID)
	theirs := pushBattlefieldCreature(g, foe.ID)
	g.BumpLayerVersionForTest()

	// Walk away first (its "away" ability has nothing to remove), then back.
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	passPriorityAroundTable(t, g)
	planeswalkToPlane(t, g, planeWindyCity.Oracle)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, mine)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, mine); c.Counters[game.CounterFlying] != 1 || !p1g2HasKeyword(g, mine, "flying") {
		t.Fatalf("the creature has %v counters and flying %v, want a flying counter and flying", c.Counters, p1g2HasKeyword(g, mine, "flying"))
	}

	// Chaos exiles a creature with flying; an opponent's, without it, is no target.
	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatalf("a blank triggered the plane")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("a creature without flying was accepted as the target")
	}
	answerPickTarget(t, g, mine)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, mine); ok {
		t.Fatal("the flier was not exiled")
	}
	if g.CastPermissionOnCardByIDForEffect(mine) == nil {
		t.Error("its owner was given no permission to cast it")
	} else if p := g.CastPermissionOnCardByIDForEffect(mine); p.Cost != "{2}" || p.Player != me.ID {
		t.Errorf("permission %+v, want the owner paying {2}", p)
	}

	// Walking away removes every flying counter.
	other := pushBattlefieldCreature(g, foe.ID)
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, other, game.CounterFlying, 2) })
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, other); c.Counters[game.CounterFlying] != 0 {
		t.Errorf("%d flying counters remain after the planeswalk away", c.Counters[game.CounterFlying])
	}
}

var planePrestonsStage = planeFixture{"804c913e-7dcd-4e83-a8d2-47b26e796c2e", "Preston's Stage", "Plane — Las Vegas"}

func TestPrestonsStageHatsForMagicTricksAndFlashForCreatures(t *testing.T) {
	requirePlaneRegisteredFull(t, planePrestonsStage)
	g, plane := newPlaneGame(t, planePrestonsStage)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	foe := opponentOf(g)

	// A sorcery is no trick.
	castCatalogSpell(t, g, "Test Ritual", "Sorcery", "", nil)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatalf("a sorcery triggered the plane")
	}
	passPriorityAroundTable(t, g)
	// An instant is.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: foe.ID}})
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Hat"); n != 1 {
		t.Fatalf("%d Hats after the planar controller cast an instant, want 1", n)
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Hat" && c.Controller != me.ID {
			t.Errorf("the Hat is controlled by %v, want the player who performed the trick", c.Controller)
		}
	}
	// Turning a face-down creature face up by another player is one too.
	bear := pushBattlefieldCreature(g, foe.ID)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventTurnedFaceUp, Actor: foe.ID, CardID: bear})
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	got := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Hat" && c.Controller == foe.ID {
			got++
		}
	}
	if got != 1 {
		t.Errorf("the opponent has %d Hats after turning a creature face up, want 1", got)
	}

	// Chaos: a flash statement for creature cards in hand, until my next turn.
	rollBlank(t, g, me.ID)
	if len(me.Statics) != 0 {
		t.Fatalf("a blank granted a statement")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	found := false
	for _, s := range me.Statics {
		if s.Timing.Timing == game.TimingFlash && s.Timing.Filter.CreatureOnly && s.Timing.FromZone == game.ZoneHand {
			found = true
		}
	}
	if !found {
		t.Errorf("no flash statement for creature cards in hand: %+v", me.Statics)
	}
}

var planeCommandZone = planeFixture{"eaa5979c-86d3-4647-9a45-fd80ba4bf192", "The Command Zone", "Plane — MagicCon"}

func TestTheCommandZoneDrawsPutsDoublesAndResetsTax(t *testing.T) {
	requirePlaneRegisteredFull(t, planeCommandZone)
	g, plane := newPlaneGame(t, planeCommandZone)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	steady := g.Seats[(g.Turn.ActiveSeat+1)%4] // controls their commander
	bare := g.Seats[(g.Turn.ActiveSeat+2)%4]   // no commander at all

	// me: Elvish Visionary (draw a card on entering) waits in the command zone.
	commander := uuid.New()
	me.Command.PushTop(game.Card{
		InstanceID: commander, Name: "Elvish Visionary", TypeLine: "Creature — Elf Shaman",
		OracleID: "c6a3a882-a127-4590-93d7-679ef4313efe", ManaCost: "{1}{G}", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	// steady: a commander on the battlefield.
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Steady Commander", TypeLine: "Legendary Creature — Test", Power: 2, Toughness: 2,
		Owner: steady.ID, Controller: steady.ID, IsCommander: true,
	})
	g.BumpLayerVersionForTest()
	myHand, steadyHand, bareHand := me.Hand.Size(), steady.Hand.Size(), bare.Hand.Size()

	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	passPriorityAroundTable(t, g)
	planeswalkToPlane(t, g, planeCommandZone.Oracle)
	passPriorityAroundTable(t, g)
	if got := steady.Hand.Size(); got != steadyHand+1 {
		t.Errorf("the player who controls their commander has %d cards, want %d", got, steadyHand+1)
	}
	if got := bare.Hand.Size(); got != bareHand {
		t.Errorf("a player with no commander drew (%d → %d)", bareHand, got)
	}
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, commander); !ok {
		// A put commander is the same card, a new object (CR 400.7).
		found := false
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Name == "Elvish Visionary" && c.Controller == me.ID {
				found = true
			}
		}
		if !found {
			t.Fatal("the commander was not put onto the battlefield")
		}
	}
	// Its enter trigger, an ability of my commander during my turn, happens twice.
	if got := me.Hand.Size(); got != myHand+2 {
		t.Errorf("hand %d → %d, want two draws from the doubled enter trigger", myHand, got)
	}

	// Chaos resets every commander's tax.
	for _, p := range g.Seats {
		p.CommanderCasts[uuid.New()] = 2
	}
	rollBlank(t, g, me.ID)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		for id, n := range p.CommanderCasts {
			if n != 0 {
				t.Errorf("%s: commander %s still cast %d times", p.Name, id, n)
			}
		}
	}
	_ = plane
}
