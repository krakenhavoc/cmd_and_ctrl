package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b43_cards_test.go — Star Trek Commander, batch 43 (#4974). Each
// test plays the real catalog entry and asserts what its text says.

const (
	trekB43EnterpriseCOracle = "f2768bab-a6a4-490a-a676-fbfbfefdd8b8"
	trekB43FarragutOracle    = "3b2e3deb-1f21-4311-b8a9-4153089ac5ff"
	trekB43PrometheusOracle  = "d19ebe9f-c24d-405c-b8e8-a602274076c3"
	trekB43StargazerOracle   = "a5b72c5d-4c73-49cd-9fc5-855856807ff2"
	trekB43ShoreLeaveOracle  = "b5f50ad0-a9b3-4bfc-aaf4-55199e489265"
	trekB43EvasiveOracle     = "3b2f1364-56e7-461a-b03f-0e14d83b0f5e"
	trekB43HeadOfSecOracle   = "fb42bade-4bbc-4119-b6cc-71695fe43dd9"
	trekB43ShuttleCrewOracle = "25855177-a416-4083-a6e4-8c9e00361d41"
	trekB43SickbayOracle     = "94688e62-f39f-41fd-9309-f16b715ef724"
	trekB43StarfleetOracle   = "83b31215-ea9b-4ad9-a39e-3fe6cdb9b7de"
)

func TestTrekB43EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB43EnterpriseCOracle: "U.S.S. Enterprise-C, Time-Tossed",
		trekB43FarragutOracle:    "U.S.S. Farragut, Bold Explorer",
		trekB43PrometheusOracle:  "U.S.S. Prometheus, Prototype",
		trekB43StargazerOracle:   "U.S.S. Stargazer, Here and There",
		trekB43ShoreLeaveOracle:  "Shore Leave Planet",
		trekB43EvasiveOracle:     "Evasive Maneuvers",
		trekB43HeadOfSecOracle:   "Head of Security",
		trekB43ShuttleCrewOracle: "Shuttle Crew",
		trekB43SickbayOracle:     "Sickbay Orderly",
		trekB43StarfleetOracle:   "Starfleet Crew",
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

// --- fixtures -------------------------------------------------------------------

// trekB43Ship puts a station Spacecraft on the battlefield under p with n
// charge counters.
func trekB43Ship(t *testing.T, g *game.Game, p uuid.UUID, name, oracle string, power, toughness, n int) uuid.UUID {
	t.Helper()
	id := b12Push(g, p, name, "Legendary Artifact — Spacecraft", oracle, power, toughness)
	if n > 0 {
		g.WithWriteLock(func() {
			if err := g.AddCounterForEffect(id, game.CounterCharge, n); err != nil {
				t.Fatalf("charge counters: %v", err)
			}
		})
	}
	return id
}

// trekB43Hand replaces p's hand with fresh cards of the given type line.
func trekB43Hand(g *game.Game, p *game.Player, typeLine string, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(names))
	g.WithWriteLock(func() {
		p.Hand.Cards = nil
		for i, n := range names {
			ids[i] = uuid.New()
			p.Hand.PushTop(game.Card{InstanceID: ids[i], Name: n, TypeLine: typeLine, Owner: p.ID, Controller: p.ID, Power: 2, Toughness: 2})
		}
	})
	return ids
}

func trekB43InHand(p *game.Player, ids ...uuid.UUID) int {
	n := 0
	for _, id := range ids {
		if p.Hand.Contains(id) {
			n++
		}
	}
	return n
}

// --- U.S.S. Enterprise-C, Time-Tossed --------------------------------------------

func TestTrekB43EnterpriseCWheelsAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "One", "Two", "Three", "Four", "Five", "Six")
	ship := castFromHand(t, g, game.Card{Name: "U.S.S. Enterprise-C, Time-Tossed", TypeLine: "Legendary Artifact — Spacecraft",
		OracleID: trekB43EnterpriseCOracle, ManaCost: "{5}", Power: 6, Toughness: 6}, game.CastSpellParams{})
	old := trekB43Hand(g, me, "Instant", "Old A", "Old B", "Old C")
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if trekB43InHand(me, old...) != 0 || me.Graveyard.Size() < 3 {
		t.Errorf("the hand was not discarded: %d old cards in hand, graveyard %d", trekB43InHand(me, old...), me.Graveyard.Size())
	}
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("three discarded, three drawn: hand %d", got)
	}
	if layeredCard(t, g, ship).IsCreature() {
		t.Error("an unstationed Enterprise-C is a creature")
	}
}

func TestTrekB43EnterpriseCMayKeepTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "One", "Two", "Three")
	castFromHand(t, g, game.Card{Name: "U.S.S. Enterprise-C, Time-Tossed", TypeLine: "Legendary Artifact — Spacecraft",
		OracleID: trekB43EnterpriseCOracle, ManaCost: "{5}", Power: 6, Toughness: 6}, game.CastSpellParams{})
	old := trekB43Hand(g, me, "Instant", "Old A", "Old B")
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if trekB43InHand(me, old...) != 2 || me.Hand.Size() != 2 {
		t.Errorf("declined: hand %d, old cards kept %d", me.Hand.Size(), trekB43InHand(me, old...))
	}
}

func TestTrekB43EnterpriseCWheelsWhenItDiesAsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "One", "Two", "Three", "Four")
	ship := trekB43Ship(t, g, me.ID, "U.S.S. Enterprise-C, Time-Tossed", trekB43EnterpriseCOracle, 6, 6, 3)
	c := layeredCard(t, g, ship)
	if !c.IsCreature() || !game.HasKeyword(&c, "flying") || c.CurrentPower() != 6 {
		t.Fatalf("at 3+ the Enterprise-C is a 6/6 flying creature: creature=%v flying=%v power=%d",
			c.IsCreature(), game.HasKeyword(&c, "flying"), c.CurrentPower())
	}
	old := trekB43Hand(g, me, "Instant", "Old A", "Old B")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ship) })
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) == nil {
		t.Fatal("dying as a creature did not offer the discard")
	}
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if trekB43InHand(me, old...) != 0 || me.Hand.Size() != 2 {
		t.Errorf("two discarded, two drawn: hand %d, old cards kept %d", me.Hand.Size(), trekB43InHand(me, old...))
	}
}

// Below three charge counters the Spacecraft is no creature, so going to
// the graveyard is not dying (CR 700.4).
func TestTrekB43EnterpriseCDoesNotDieAsANoncreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ship := trekB43Ship(t, g, me.ID, "U.S.S. Enterprise-C, Time-Tossed", trekB43EnterpriseCOracle, 6, 6, 2)
	trekB43Hand(g, me, "Instant", "Old A")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ship) })
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("a noncreature Spacecraft put into the graveyard triggered as if it died")
	}
}

// --- U.S.S. Farragut, Bold Explorer --------------------------------------------

// trekB43FarragutPutOne plays a land with Farragut out and puts `pick`
// onto the battlefield with the landfall trigger.
func trekB43FarragutPutOne(t *testing.T, g *game.Game, me *game.Player, pick uuid.UUID) {
	t.Helper()
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	c := latestChooseCards(g, me.ID)
	if c == nil {
		t.Fatal("landfall did not offer a creature card from hand")
	}
	var picks []uuid.UUID
	if pick != uuid.Nil {
		picks = []uuid.UUID{pick}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, picks); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestTrekB43FarragutPutsAHastyCreatureThenBottomsItAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top", "Second")
	ship := trekB43Ship(t, g, me.ID, "U.S.S. Farragut, Bold Explorer", trekB43FarragutOracle, 4, 4, 0)
	hand := trekB43Hand(g, me, "Creature — Bear", "Hasty Bear")
	trekB43FarragutPutOne(t, g, me, hand[0])
	if !g.Battlefield.Contains(hand[0]) {
		t.Fatal("the creature card did not enter")
	}
	if !hasEffectiveKeyword(t, g, hand[0], "haste") {
		t.Error("the creature did not gain haste")
	}
	if layeredCard(t, g, ship).IsCreature() {
		t.Error("an unstationed Farragut is a creature")
	}
	handBefore := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hand[0]) {
		t.Fatal("the creature is still on the battlefield after the end step trigger")
	}
	if bottom := me.Library.Cards[0]; bottom.InstanceID != hand[0] {
		t.Errorf("the library's bottom card is %q, want the creature", bottom.Name)
	}
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("when you do, draw a card: hand %d -> %d", handBefore, got)
	}
}

// The creature left before the end step: nothing goes to the library, so
// "when you do" never happens and nothing is drawn (CR 603.7c, 603.12).
func TestTrekB43FarragutDrawsNothingWhenTheCreatureIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top", "Second")
	trekB43Ship(t, g, me.ID, "U.S.S. Farragut, Bold Explorer", trekB43FarragutOracle, 4, 4, 0)
	hand := trekB43Hand(g, me, "Creature — Bear", "Doomed Bear")
	trekB43FarragutPutOne(t, g, me, hand[0])
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hand[0]) })
	librarySize, handBefore := me.Library.Size(), me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != librarySize || me.Hand.Size() != handBefore {
		t.Errorf("a creature that died first was tucked or paid a draw: library %d -> %d, hand %d -> %d",
			librarySize, me.Library.Size(), handBefore, me.Hand.Size())
	}
}

func TestTrekB43FarragutMayPutNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	trekB43Ship(t, g, me.ID, "U.S.S. Farragut, Bold Explorer", trekB43FarragutOracle, 4, 4, 0)
	hand := trekB43Hand(g, me, "Creature — Bear", "Kept Bear")
	trekB43FarragutPutOne(t, g, me, uuid.Nil)
	if g.Battlefield.Contains(hand[0]) || !me.Hand.Contains(hand[0]) {
		t.Error("declining the put still moved the creature card")
	}
}

func TestTrekB43FarragutFliesAtSeven(t *testing.T) {
	g := newCatalogGame(t)
	ship := trekB43Ship(t, g, g.Seats[0].ID, "U.S.S. Farragut, Bold Explorer", trekB43FarragutOracle, 4, 4, 7)
	c := layeredCard(t, g, ship)
	if !c.IsCreature() || !game.HasKeyword(&c, "flying") || c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("at 7+ Farragut is a 4/4 flying creature: creature=%v flying=%v %d/%d",
			c.IsCreature(), game.HasKeyword(&c, "flying"), c.CurrentPower(), c.CurrentToughness())
	}
}

// --- U.S.S. Prometheus, Prototype ------------------------------------------------

func trekB43PrometheusAttack(t *testing.T, counters int) (hit, other, mine int) {
	t.Helper()
	g := newCatalogGame(t)
	me, def, oth := g.Seats[0], g.Seats[1], g.Seats[2]
	ship := trekB43Ship(t, g, me.ID, "U.S.S. Prometheus, Prototype", trekB43PrometheusOracle, 6, 6, counters)
	c := layeredCard(t, g, ship)
	if !c.IsCreature() || !game.HasKeyword(&c, "flying") || c.CurrentPower() != 6 {
		t.Fatalf("at %d counters the Prometheus is a 6/6 flier: creature=%v flying=%v power=%d",
			counters, c.IsCreature(), game.HasKeyword(&c, "flying"), c.CurrentPower())
	}
	hitBefore, otherBefore, mineBefore := def.Life, oth.Life, me.Life
	declareAttack(t, g, def.ID, ship)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	return hitBefore - def.Life, otherBefore - oth.Life, mineBefore - me.Life
}

func TestTrekB43PrometheusSpreadsCombatDamageAtEight(t *testing.T) {
	hit, other, mine := trekB43PrometheusAttack(t, 8)
	if hit != 6 || other != 6 || mine != 0 {
		t.Errorf("at 8+: defender lost %d, each other opponent %d, controller %d; want 6, 6, 0", hit, other, mine)
	}
}

func TestTrekB43PrometheusBelowEightOnlyHitsTheDefender(t *testing.T) {
	hit, other, _ := trekB43PrometheusAttack(t, 4)
	if hit != 6 || other != 0 {
		t.Errorf("at 4 counters: defender lost %d, other opponent %d; want 6, 0", hit, other)
	}
}

// --- U.S.S. Stargazer, Here and There --------------------------------------------

func TestTrekB43StargazerMakesAnAttackingIllusionForOneCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ship := trekB43Ship(t, g, me.ID, "U.S.S. Stargazer, Here and There", trekB43StargazerOracle, 4, 3, 6)
	if c := layeredCard(t, g, ship); !c.IsCreature() || !game.HasKeyword(&c, "flying") || c.CurrentPower() != 4 || c.CurrentToughness() != 3 {
		t.Fatalf("at 6+ the Stargazer is a 4/3 flier: %+v", c)
	}
	before := opp.Life
	declareAttack(t, g, opp.ID, ship)
	passAndSettleAttackSplit(t, g, me.ID)
	illusions := battlefieldIDsNamed(g, "Spacecraft Illusion")
	if len(illusions) != 1 {
		t.Fatalf("%d Spacecraft Illusions, want 1", len(illusions))
	}
	tok := cardByID(g, illusions[0])
	if !tok.Tapped || tok.AttackingTarget != opp.ID || effectivePower(t, g, illusions[0]) != 4 ||
		effectiveToughness(t, g, illusions[0]) != 3 || !hasEffectiveKeyword(t, g, illusions[0], "flying") ||
		!tok.IsArtifact() || len(tok.Colors) != 0 {
		t.Errorf("the Illusion is not a tapped, attacking, colorless 4/3 artifact flier: %+v", tok)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != before-8 {
		t.Errorf("defender at %d, want %d (the Stargazer and the Illusion)", opp.Life, before-8)
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Spacecraft Illusion")); n != 0 {
		t.Errorf("%d Illusions left after end of combat, want 0", n)
	}
	if !g.Battlefield.Contains(ship) {
		t.Error("the Stargazer itself was exiled")
	}
}

func TestTrekB43StargazerIsNoCreatureBelowSix(t *testing.T) {
	g := newCatalogGame(t)
	ship := trekB43Ship(t, g, g.Seats[0].ID, "U.S.S. Stargazer, Here and There", trekB43StargazerOracle, 4, 3, 5)
	if layeredCard(t, g, ship).IsCreature() {
		t.Error("the Stargazer is a creature at 5 charge counters")
	}
}

// --- Shore Leave Planet ---------------------------------------------------------

func TestTrekB43ShoreLeavePlanetEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	land := playLandFromHand(t, g, "Shore Leave Planet", trekB43ShoreLeaveOracle)
	if c, ok := battlefieldCard(g, land); !ok || !c.Tapped {
		t.Errorf("Shore Leave Planet on the battlefield %v, tapped %v; want both", ok, c.Tapped)
	}
}

func TestTrekB43ShoreLeavePlanetBecomesANonlegendaryRobotCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	planet := dcSeed(g, me.ID, "Shore Leave Planet", "Land — Planet", trekB43ShoreLeaveOracle, 0, 0)
	legend := b12Creature(g, me.ID, "Captain Legend", "Legendary Creature — Human Officer", 3, 4)
	dcMana(t, g, me.ID, "{C}{C}{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, planet, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: legend}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	got := dcCard(t, g, planet)
	if got.Name != "Captain Legend" || !got.IsCreature() || got.IsLand() || got.CurrentPower() != 3 {
		t.Fatalf("the Planet is %q (creature %v, land %v, power %d), want a copy of Captain Legend",
			got.Name, got.IsCreature(), got.IsLand(), got.CurrentPower())
	}
	if got.HasSupertype("Legendary") || !got.IsArtifact() || !got.HasSubtype("Robot") || !got.HasSubtype("Human") {
		t.Errorf("the copy is %q; want a nonlegendary Artifact Creature — Human Officer Robot", got.TypeLine)
	}
	if !g.Battlefield.Contains(legend) || legendPrompt(g, me.ID) != nil {
		t.Error("the nonlegendary copy put the original to the legend rule")
	}
	dcCleanup(g)
	if back := dcCard(t, g, planet); back.Name != "Shore Leave Planet" || !back.IsLand() || back.IsCreature() {
		t.Errorf("after cleanup the Planet is %q (land %v); want the land back", back.Name, back.IsLand())
	}
}

func TestTrekB43ShoreLeavePlanetTargetsOnlyYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	planet := dcSeed(g, me.ID, "Shore Leave Planet", "Land — Planet", trekB43ShoreLeaveOracle, 0, 0)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	dcMana(t, g, me.ID, "{C}{C}{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, planet, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Error("an opponent's creature was accepted as the copy's target")
	}
}

// --- Evasive Maneuvers -------------------------------------------------------------

func TestTrekB43EvasiveManeuversPumpsYourBlockerAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	crtBlock(g, blocker, attacker, me.ID)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Evasive Maneuvers", "Instant", trekB43EvasiveOracle, b16TargetCard(blocker))
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, blocker), effectiveToughness(t, g, blocker); p != 4 || tg != 4 {
		t.Errorf("blocker is %d/%d, want 4/4", p, tg)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn (the pushed spell left)", hand, got)
	}
}

func TestTrekB43EvasiveManeuversRefusesAnIdleOrOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	idle := b12Creature(g, g.Seats[0].ID, "Idle Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Evasive Maneuvers", "Instant", trekB43EvasiveOracle, b16TargetCard(idle)); err == nil {
		t.Error("a creature that is not blocking was accepted as the target")
	}
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	theirBlocker := b12Creature(g2, opp2.ID, "Their Blocker", "Creature — Bear", 2, 2)
	myAttacker := b12Creature(g2, me2.ID, "My Attacker", "Creature — Bear", 2, 2)
	crtBlock(g2, theirBlocker, myAttacker, opp2.ID)
	if err := castCatalogSpellErr(t, g2, "Evasive Maneuvers", "Instant", trekB43EvasiveOracle, b16TargetCard(theirBlocker)); err == nil {
		t.Error("an opponent's blocking creature was accepted as the target")
	}
}

// --- Shuttle Crew -------------------------------------------------------------------

func TestTrekB43ShuttleCrewDealsFourToATappedOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	untapped := b12Creature(g, opp.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Tapped Bear", "Creature — Bear", 2, 2)
	battlefieldCardPtr(t, g, mine).Tapped = true
	big := b12Creature(g, opp.ID, "Tapped Giant", "Creature — Giant", 6, 6)
	battlefieldCardPtr(t, g, big).Tapped = true

	crew := castAndResolveCreature(t, g, "Shuttle Crew", "Creature — Human Officer", trekB43ShuttleCrewOracle)
	assertKeywords(t, g, crew, "flying")
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, id := range []uuid.UUID{untapped, mine} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
			t.Errorf("an untapped or own creature %s was accepted as a target", id)
		}
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err != nil {
		t.Fatalf("the tapped opposing creature was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := battlefieldCardPtr(t, g, big).DamageMarked; got != 4 {
		t.Errorf("damage on the tapped creature = %d, want 4", got)
	}
}

// --- the keyword-only creatures ------------------------------------------------------

func TestTrekB43KeywordCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	assertKeywords(t, g, b12Push(g, me.ID, "Head of Security", "Creature — Andorian Soldier", trekB43HeadOfSecOracle, 2, 2), "first strike")
	assertKeywords(t, g, b12Push(g, me.ID, "Sickbay Orderly", "Creature — Caitian Doctor", trekB43SickbayOracle, 3, 2), "lifelink")
	crew := b12Push(g, me.ID, "Starfleet Crew", "Creature — Officer", trekB43StarfleetOracle, 2, 3)
	if c := layeredCard(t, g, crew); len(c.Keywords) != 0 || c.CurrentPower() != 2 || c.CurrentToughness() != 3 {
		t.Errorf("Starfleet Crew is a vanilla 2/3: keywords %v, %d/%d", c.Keywords, c.CurrentPower(), c.CurrentToughness())
	}
}

// Sickbay Orderly's lifelink pays its controller what it deals in combat.
func TestTrekB43SickbayOrderlyGainsWhatItDeals(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	orderly := b12Push(g, me.ID, "Sickbay Orderly", "Creature — Caitian Doctor", trekB43SickbayOracle, 3, 2)
	life := me.Life
	declareAttack(t, g, opp.ID, orderly)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("lifelink: life %d -> %d, want +3", life, me.Life)
	}
}
