package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_tz2_cards_test.go — Foundations, the second T–Z batch (#4465): the
// cards that needed no new seam. Each test plays the real catalog entry
// through the engine and asserts what its text says.

const (
	fdnTZ2DjinnOracle    = "b94c233e-11d3-4556-a1da-bd193e58b663"
	fdnTZ2SeraphOracle   = "3250466f-82a4-45e1-9722-ff247ddf8c15"
	fdnTZ2TerrorOracle   = "b9d0f2e1-62c2-44fd-ad38-471daf17bb0a"
	fdnTZ2BansheeOracle  = "888f58c0-1860-4df4-ad75-53e634150680"
	fdnTZ2PaladinOracle  = "d373dc71-5196-43a0-8cc5-67e2a591aad0"
	fdnTZ2VeteranOracle  = "bf4268e6-170a-445b-b215-718f7494ae28"
	fdnTZ2WardensOracle  = "fa7d09aa-2ebc-4571-81da-836cd80cf554"
	fdnTZ2GourmandOracle = "3386414c-32fa-40ae-9b30-dab976a158c1"
)

func TestFdnTZ2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnTZ2DjinnOracle:    "Tempest Djinn",
		fdnTZ2SeraphOracle:   "Vanguard Seraph",
		fdnTZ2TerrorOracle:   "Tolarian Terror",
		fdnTZ2BansheeOracle:  "Tragic Banshee",
		fdnTZ2PaladinOracle:  "Twinblade Paladin",
		fdnTZ2VeteranOracle:  "Volley Veteran",
		fdnTZ2WardensOracle:  "Wardens of the Cycle",
		fdnTZ2GourmandOracle: "Vampire Gourmand",
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

func TestFdnTZ2PrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Tempest Djinn", "Creature — Djinn", fdnTZ2DjinnOracle, 0, 4},
		{"Vanguard Seraph", "Creature — Angel Warrior", fdnTZ2SeraphOracle, 3, 3},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		assertKeywords(t, g, id, "flying")
	}
}

// --- Tempest Djinn -----------------------------------------------------------

func TestFdnTZ2DjinnCountsOnlyYourBasicIslands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	djinn := blbB1Push(g, me.ID, "Tempest Djinn", "Creature — Djinn", fdnTZ2DjinnOracle, 0, 4)
	if p, tough := ptOf(t, g, djinn); p != 0 || tough != 4 {
		t.Errorf("with no Islands it is %d/%d, want 0/4", p, tough)
	}
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, me.ID, "Steam Vents", "Land — Island Mountain")
	b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	if p, tough := ptOf(t, g, djinn); p != 2 || tough != 4 {
		t.Errorf("with two basic Islands of yours it is %d/%d, want 2/4", p, tough)
	}
}

// --- Vanguard Seraph ---------------------------------------------------------

func TestFdnTZ2SeraphSurveilsOnTheFirstGainOfATurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	seraph := b12Push(g, me.ID, "Vanguard Seraph", "Creature — Angel Warrior", fdnTZ2SeraphOracle, 3, 3)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(seraph, me.ID, 2) })
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the first gain of the turn queued no surveil")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
}

func TestFdnTZ2SeraphIgnoresALaterGainInTheSameTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	seraph := b12Push(g, me.ID, "Vanguard Seraph", "Creature — Angel Warrior", fdnTZ2SeraphOracle, 3, 3)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(seraph, me.ID, 2) })
	passPriorityAroundTable(t, g)
	if c := surveilChoiceFor(g, me.ID); c != nil {
		t.Errorf("the turn's second gain surveilled: %+v", c)
	}
}

// --- Tolarian Terror ---------------------------------------------------------

func TestFdnTZ2TerrorCostsLessForEachInstantAndSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	if got := selfPricedMV(t, g, me, fdnTZ2TerrorOracle, "Creature — Serpent", "{6}{U}"); got != 7 {
		t.Errorf("an empty graveyard: it costs %d, want 7", got)
	}
	pushGraveyardCardTyped(me, "Old Instant", "Instant")
	pushGraveyardCardTyped(me, "Old Instant", "Instant")
	pushGraveyardCardTyped(me, "Old Sorcery", "Sorcery")
	pushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	if got := selfPricedMV(t, g, me, fdnTZ2TerrorOracle, "Creature — Serpent", "{6}{U}"); got != 4 {
		t.Errorf("three instants and sorceries: it costs %d, want 4", got)
	}
	for i := 0; i < 6; i++ {
		pushGraveyardCardTyped(me, "Old Sorcery", "Sorcery")
	}
	if got := selfPricedMV(t, g, me, fdnTZ2TerrorOracle, "Creature — Serpent", "{6}{U}"); got != 1 {
		t.Errorf("nine instants and sorceries: it costs %d, want 1 (only the generic part is reduced)", got)
	}
}

func TestFdnTZ2TerrorHasWardTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	terror := pushCatalogPermanent(g, me.ID, "Tolarian Terror", "Creature — Serpent", fdnTZ2TerrorOracle, false)
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, terror)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {2} payment for the opponent targeting Tolarian Terror")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(terror) || !opp.Graveyard.Contains(blade) {
		t.Error("declining ward did not counter the spell")
	}
}

// --- Tragic Banshee ----------------------------------------------------------

func TestFdnTZ2BansheeShrinksByOneOrThirteenWithMorbid(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		giant := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
		bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		advanceToMain(t, g)
		if died {
			killCreature(t, g, opp.ID, bear)
		}
		banshee := castHoldCreature(t, g, "Tragic Banshee", "Creature — Spirit", fdnTZ2BansheeOracle, 5, 3)
		p := latestPickTarget(g, me.ID)
		if p == nil {
			t.Fatalf("died %v: the enters trigger asked for no target", died)
		}
		if hasID(p.PickTargetCards, mine) || hasID(p.PickTargetCards, banshee) {
			t.Errorf("died %v: the trigger offers a creature you control", died)
		}
		pickCard(t, g, me.ID, giant)
		passPriorityAroundTable(t, g)
		if died {
			if g.Battlefield.Contains(giant) {
				t.Error("with morbid the 5/5 survived -13/-13")
			}
			continue
		}
		if pw, tt := ptOf(t, g, giant); pw != 4 || tt != 4 {
			t.Errorf("without morbid the 5/5 is %d/%d, want 4/4", pw, tt)
		}
	}
}

// --- Twinblade Paladin -------------------------------------------------------

func TestFdnTZ2PaladinHasDoubleStrikeAtTwentyFiveAndGrowsOnGains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	paladin := b12Push(g, me.ID, "Twinblade Paladin", "Creature — Human Knight", fdnTZ2PaladinOracle, 3, 3)
	advanceToMain(t, g)
	if me.Life < 25 {
		t.Fatalf("the table starts at %d life", me.Life)
	}
	if !hasEffectiveKeyword(t, g, paladin, "double strike") {
		t.Errorf("at %d life it has no double strike", me.Life)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 24-me.Life) })
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, paladin, "double strike") {
		t.Error("at 24 life it still has double strike")
	}
	if n := plusOnes(g, paladin); n != 0 {
		t.Errorf("a life loss grew it to %d counters", n)
	}
	for i := 0; i < 2; i++ {
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
		passPriorityAroundTable(t, g)
	}
	if n := plusOnes(g, paladin); n != 2 {
		t.Errorf("%d counters after two gains, want 2", n)
	}
	if me.Life != 26 || !hasEffectiveKeyword(t, g, paladin, "double strike") {
		t.Errorf("at %d life: double strike %v, want it back at 26", me.Life, hasEffectiveKeyword(t, g, paladin, "double strike"))
	}
}

// --- Volley Veteran ----------------------------------------------------------

func TestFdnTZ2VeteranDealsDamageEqualToYourGoblins(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	b12Creature(g, me.ID, "Goblin Friend", "Creature — Goblin", 1, 1)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	giant := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	veteran := castHoldCreature(t, g, "Volley Veteran", "Creature — Goblin Warrior", fdnTZ2VeteranOracle, 4, 2)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if hasID(p.PickTargetCards, veteran) {
		t.Error("the trigger offers a creature you control")
	}
	pickCard(t, g, me.ID, giant)
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, giant); got != 2 {
		t.Errorf("the Giant has %d damage, want 2 (the Veteran and one other Goblin of yours)", got)
	}
}

// --- Wardens of the Cycle ----------------------------------------------------

func TestFdnTZ2WardensTriggerOnlyAfterADeath(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	blbB1Push(g, me.ID, "Wardens of the Cycle", "Creature — Elf Warlock", fdnTZ2WardensOracle, 3, 4)
	advanceToEndStepOf(t, g, seat)
	noModePick(t, g, me.ID, "no creature died this turn")
}

func TestFdnTZ2WardensModes(t *testing.T) {
	for _, mode := range []int{0, 1} {
		g := newCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me, opp, _ := blbB1Seats(g)
		seedLibrary(me, "A", "B", "C")
		blbB1Push(g, me.ID, "Wardens of the Cycle", "Creature — Elf Warlock", fdnTZ2WardensOracle, 3, 4)
		bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		advanceToMain(t, g)
		killCreature(t, g, opp.ID, bear)
		advanceToEndStepOf(t, g, seat)
		life, hand := me.Life, me.Hand.Size()
		chooseModeNow(t, g, me.ID, []int{0, 1}, mode)
		passPriorityAroundTable(t, g)
		wantLife, wantHand := life+2, hand
		if mode == 1 {
			wantLife, wantHand = life-1, hand+1
		}
		if me.Life != wantLife || me.Hand.Size() != wantHand {
			t.Errorf("mode %d: life %d → %d, hand %d → %d; want life %d, hand %d",
				mode, life, me.Life, hand, me.Hand.Size(), wantLife, wantHand)
		}
	}
}

// --- Vampire Gourmand --------------------------------------------------------

func TestFdnTZ2GourmandSacrificesToDrawAndBeUnblockable(t *testing.T) {
	for _, sac := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		seedLibrary(me, "A", "B")
		gourmand := blbB1Push(g, me.ID, "Vampire Gourmand", "Creature — Vampire", fdnTZ2GourmandOracle, 2, 2)
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		relic := b12Permanent(g, me.ID, "Relic", "Artifact")
		wall := blbB1Push(g, opp.ID, "Their Wall", "Creature — Wall", "", 0, 4)
		declareAttack(t, g, opp.ID, gourmand)
		passPriorityAroundTable(t, g)
		prompt := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
		if prompt == nil || prompt.ChooseMin != 0 || !hasID(prompt.ChooseCards, bear) ||
			hasID(prompt.ChooseCards, gourmand) || hasID(prompt.ChooseCards, relic) {
			t.Fatalf("want a may-sacrifice over another creature: %+v", prompt)
		}
		hand := me.Hand.Size()
		if sac {
			answerOwnPermanents(t, g, me.ID, bear)
		} else {
			answerOwnPermanents(t, g, me.ID)
		}
		passPriorityAroundTable(t, g)
		if got := g.Battlefield.Contains(bear); got == sac {
			t.Errorf("sacrificed %v: the Bear is on the battlefield %v", sac, got)
		}
		wantHand := hand
		if sac {
			wantHand = hand + 1
		}
		if got := me.Hand.Size(); got != wantHand {
			t.Errorf("sacrificed %v: hand %d, want %d", sac, got, wantHand)
		}
		advanceTo(t, g, game.StepDeclareBlockers)
		err := g.DeclareBlocker(wall, gourmand)
		if sac && err == nil {
			t.Error("blocked after the sacrifice")
		}
		if !sac && err != nil {
			t.Errorf("without a sacrifice the block was refused: %v", err)
		}
	}
}
