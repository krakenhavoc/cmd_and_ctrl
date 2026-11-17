package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 ONE batch 19 — copies and casting from exile (#5479).

const (
	s82OneBladeOfSharedSouls = "bfecd486-8326-426b-ac2b-54946f6ae707"
	s82OneBlueSunsTwilight   = "11e637b4-6ce0-4514-98d6-02a2d073f73b"
	s82OneRedSunsTwilight    = "21c77cdb-9ba1-4621-b096-7eb94c076cef"
	s82OneHellraiser         = "eb7b024c-da9b-4544-8648-0ac0395cf7fc"
	s82OneSpelldancer        = "5b90f8ed-84b2-4306-9dfa-b5654ea0e4cd"
	s82OneArgentumMasticore  = "3ceee5c2-321c-49ec-ba49-1fac7f6ee2cc"
	s82OneChissGoria         = "3f4d1ae1-d5c6-4afb-94c8-8e0ae4891501"
	s82OneIxhel              = "9fb7982f-fe35-4dea-a515-a0039ddbbaec"
)

func TestS82OneB19CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		s82OneBladeOfSharedSouls: "Blade of Shared Souls", s82OneBlueSunsTwilight: "Blue Sun's Twilight",
		s82OneRedSunsTwilight: "Red Sun's Twilight", s82OneHellraiser: "Capricious Hellraiser",
		s82OneSpelldancer: "Mercurial Spelldancer", s82OneArgentumMasticore: "Argentum Masticore",
		s82OneChissGoria: "Chiss-Goria, Forge Tyrant", s82OneIxhel: "Ixhel, Scion of Atraxa",
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

// s82OneB19CastStrict casts a card from the active seat's hand with real
// payment, at sorcery timing.
func s82OneB19CastStrict(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, x int, targets ...game.TargetRef) error {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: x, Targets: targets})
}

// --- Blue Sun's Twilight ----------------------------------------------

func TestS82OneBlueSunsTwilightTakesTheCreatureAndDoesNotCopyBelowFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ox := b27Push(g, opp.ID, "Ox", "Creature — Ox", "", "{3}", 3, 3)
	castXSpell(t, g, "Blue Sun's Twilight", "Sorcery", s82OneBlueSunsTwilight, "{X}{U}{U}", 4, cardTarget(ox))
	passPriorityAroundTable(t, g)
	if got := aegisCardNow(g, ox).Controller; got != me.ID {
		t.Errorf("the Ox is controlled by %s, want me", got)
	}
	if n := countBattlefieldByName(g, "Ox"); n != 1 {
		t.Errorf("%d Oxen, want 1 (X is 4, no copy)", n)
	}
}

func TestS82OneBlueSunsTwilightAtFiveMakesACopyOfTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ox := b27Push(g, opp.ID, "Ox", "Creature — Ox", "", "{3}", 3, 3)
	castXSpell(t, g, "Blue Sun's Twilight", "Sorcery", s82OneBlueSunsTwilight, "{X}{U}{U}", 5, cardTarget(ox))
	passPriorityAroundTable(t, g)
	if got := aegisCardNow(g, ox).Controller; got != me.ID {
		t.Errorf("the Ox is controlled by %s, want me", got)
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ox" {
			n++
			if c.Controller != me.ID {
				t.Errorf("an Ox is controlled by %s, want me (the copy is mine)", c.Controller)
			}
			if c.IsToken() && (c.Power != 3 || c.Toughness != 3) {
				t.Errorf("the token copy is %d/%d, want 3/3", c.Power, c.Toughness)
			}
		}
	}
	if n != 2 {
		t.Errorf("%d Oxen, want 2 (the stolen one and its token copy)", n)
	}
}

func TestS82OneBlueSunsTwilightRefusesAMoreExpensiveCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ox := b27Push(g, opp.ID, "Ox", "Creature — Ox", "", "{3}", 3, 3)
	fillPoolColored(me, "U", 2)
	fillPool(me, 2)
	if err := s82OneB19CastStrict(t, g, "Blue Sun's Twilight", "Sorcery", s82OneBlueSunsTwilight, "{X}{U}{U}", 2, cardTarget(ox)...); err == nil {
		t.Error("X=2 may not target a creature with mana value 3")
	}
}

// --- Red Sun's Twilight -----------------------------------------------

func s82OneB19Artifact(g *game.Game, owner uuid.UUID, name string, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Artifact", ManaCost: "{2}",
		Keywords: keywords, Owner: owner, Controller: owner,
	})
}

func TestS82OneRedSunsTwilightDestroysUpToXArtifactsAndCopiesAtFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := s82OneB19Artifact(g, opp.ID, "Rock")
	b := s82OneB19Artifact(g, opp.ID, "Gem")
	castXSpell(t, g, "Red Sun's Twilight", "Sorcery", s82OneRedSunsTwilight, "{X}{R}{R}", 5, cardTargets(a, b))
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, a) != nil || findBattlefieldCardByID(g, b) != nil {
		t.Fatal("an artifact was not destroyed")
	}
	rock, gem := countBattlefieldByName(g, "Rock"), countBattlefieldByName(g, "Gem")
	if rock != 1 || gem != 1 {
		t.Fatalf("%d Rock and %d Gem tokens, want one copy of each", rock, gem)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Rock" && c.Name != "Gem" {
			continue
		}
		if !c.IsToken() || c.Controller != me.ID {
			t.Errorf("%s: token %v controller %s, want my token", c.Name, c.IsToken(), c.Controller)
		}
		if !hasEffectiveKeyword(t, g, c.InstanceID, "haste") {
			t.Errorf("the %s token has no haste", c.Name)
		}
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if countBattlefieldByName(g, "Rock") != 0 || countBattlefieldByName(g, "Gem") != 0 {
		t.Error("the tokens were not exiled at the beginning of the next end step")
	}
}

func TestS82OneRedSunsTwilightBelowFiveMakesNoCopies(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := s82OneB19Artifact(g, opp.ID, "Rock")
	castXSpell(t, g, "Red Sun's Twilight", "Sorcery", s82OneRedSunsTwilight, "{X}{R}{R}", 4, cardTargets(a))
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, a) != nil {
		t.Error("the Rock was not destroyed")
	}
	if n := countBattlefieldByName(g, "Rock"); n != 0 {
		t.Errorf("%d Rock tokens, want none below X=5", n)
	}
}

func TestS82OneRedSunsTwilightDoesNotCopyAnArtifactThatSurvives(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := s82OneB19Artifact(g, opp.ID, "Rock", "indestructible")
	b := s82OneB19Artifact(g, opp.ID, "Gem")
	castXSpell(t, g, "Red Sun's Twilight", "Sorcery", s82OneRedSunsTwilight, "{X}{R}{R}", 5, cardTargets(a, b))
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, a) == nil {
		t.Fatal("the indestructible Rock was destroyed")
	}
	if n := countBattlefieldByName(g, "Rock"); n != 1 {
		t.Errorf("%d Rocks, want only the survivor", n)
	}
	if n := countBattlefieldByName(g, "Gem"); n != 1 {
		t.Errorf("%d Gems, want the destroyed one's token copy", n)
	}
}

func TestS82OneRedSunsTwilightCannotTargetMoreThanX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := s82OneB19Artifact(g, opp.ID, "Rock")
	b := s82OneB19Artifact(g, opp.ID, "Gem")
	fillPoolColored(me, "R", 2)
	fillPool(me, 1)
	if err := s82OneB19CastStrict(t, g, "Red Sun's Twilight", "Sorcery", s82OneRedSunsTwilight, "{X}{R}{R}", 1, cardTargets(a, b)...); err == nil {
		t.Error("X=1 may not target two artifacts")
	}
}

// --- Mercurial Spelldancer --------------------------------------------

func s82OneB19Spelldancer(g *game.Game, owner uuid.UUID, oil int) uuid.UUID {
	id := b27Push(g, owner, "Mercurial Spelldancer", "Creature — Phyrexian Rogue", s82OneSpelldancer, "{1}{U}", 2, 1, "U")
	if oil > 0 {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Counters = map[string]int{s82OneOilCounter: oil}
				}
			}
		})
	}
	return id
}

func TestS82OneSpelldancerGainsOilForNoncreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dancer := s82OneB19Spelldancer(g, me.ID, 0)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if n := countersOn(g, dancer, s82OneOilCounter); n != 1 {
		t.Fatalf("%d oil counters after a noncreature spell, want 1", n)
	}
	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, dancer, s82OneOilCounter); n != 1 {
		t.Errorf("%d oil counters after a creature spell, want still 1", n)
	}
}

func TestS82OneSpelldancerCopiesTheNextInstantAfterTwoOilCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dancer := s82OneB19Spelldancer(g, me.ID, 2)
	attackWith(t, g, opp.ID, dancer)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, dancer, s82OneOilCounter); n != 0 {
		t.Fatalf("%d oil counters left, want 0", n)
	}
	before := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	settleKeepingTarget(t, g, me.ID, opp.ID)
	if got := lifeOf(g, opp.ID); got != before-6 {
		t.Errorf("life %d -> %d, want -6: the Bolt and its copy", before, got)
	}
	// The delayed trigger is spent: a second instant is not copied.
	before = lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	settleKeepingTarget(t, g, me.ID, opp.ID)
	if got := lifeOf(g, opp.ID); got != before-3 {
		t.Errorf("life %d -> %d, want -3: the second Bolt is not copied", before, got)
	}
}

func TestS82OneSpelldancerDecliningKeepsTheCountersAndCopiesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dancer := s82OneB19Spelldancer(g, me.ID, 2)
	attackWith(t, g, opp.ID, dancer)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, dancer, s82OneOilCounter); n != 2 {
		t.Fatalf("%d oil counters, want 2 kept", n)
	}
	before := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != before-3 {
		t.Errorf("life %d -> %d, want -3: nothing is copied", before, got)
	}
}

func TestS82OneSpelldancerWithOneOilCounterCannotPay(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dancer := s82OneB19Spelldancer(g, me.ID, 1)
	attackWith(t, g, opp.ID, dancer)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, dancer, s82OneOilCounter); n != 1 {
		t.Fatalf("%d oil counters, want the one kept (two cannot be removed)", n)
	}
	before := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != before-3 {
		t.Errorf("life %d -> %d, want -3", before, got)
	}
}

// --- Argentum Masticore -----------------------------------------------

func s82OneB19Masticore(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Argentum Masticore", "Artifact Creature — Phyrexian Masticore", s82OneArgentumMasticore, "{5}", 5, 5)
}

func TestS82OneMasticoreHasFirstStrikeAndProtectionFromMulticolored(t *testing.T) {
	g := newCatalogGame(t)
	m := s82OneB19Masticore(g, g.Seats[0].ID)
	assertKeywords(t, g, m, "first strike")
	if !hasEffectiveKeyword(t, g, m, "protection from multicolored") {
		t.Error("no protection from multicolored")
	}
}

func TestS82OneMasticoreDiscardKeepsItAndDestroysAPermanentAtMostThatManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	m := s82OneB19Masticore(g, me.ID)
	cheap := b27Push(g, opp.ID, "Cheap Rock", "Artifact", "", "{2}", 0, 0)
	dear := b27Push(g, opp.ID, "Dear Rock", "Artifact", "", "{5}", 0, 0)
	land := b27Push(g, opp.ID, "Their Forest", "Land", "", "", 0, 0)
	me.Hand.Cards = nil
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID})
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, pitch)
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, cheap)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, m) == nil {
		t.Error("the Masticore was sacrificed although a card was discarded")
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the card was not discarded")
	}
	if findBattlefieldCardByID(g, cheap) != nil {
		t.Error("the mana value 2 permanent survived a mana value 3 discard")
	}
	if findBattlefieldCardByID(g, dear) == nil {
		t.Error("the mana value 5 permanent was destroyed by a mana value 3 discard")
	}
	if findBattlefieldCardByID(g, land) == nil {
		t.Error("a land was destroyed")
	}
}

func TestS82OneMasticoreDeclineingTheDiscardSacrificesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	m := s82OneB19Masticore(g, me.ID)
	keep := uuid.New()
	me.Hand.Cards = nil
	me.Hand.PushTop(game.Card{InstanceID: keep, Name: "Keeper", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, m) != nil {
		t.Error("the Masticore stayed although nothing was discarded")
	}
	if !me.Hand.Contains(keep) {
		t.Error("the card left the hand")
	}
}

func TestS82OneMasticoreWithAnEmptyHandIsSacrificedAtOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	m := s82OneB19Masticore(g, me.ID)
	me.Hand.Cards = nil
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, m) != nil {
		t.Error("the Masticore stayed with no card to discard")
	}
}

func TestS82OneMasticoreDiscardingALandDestroysOnlyAManaValueZeroPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	m := s82OneB19Masticore(g, me.ID)
	rock := b27Push(g, opp.ID, "Cheap Rock", "Artifact", "", "{2}", 0, 0)
	me.Hand.Cards = nil
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, m) == nil {
		t.Error("the Masticore was sacrificed")
	}
	if findBattlefieldCardByID(g, rock) == nil {
		t.Error("a mana value 2 permanent was destroyed by a land discard")
	}
}

// --- Capricious Hellraiser --------------------------------------------

func TestS82OneHellraiserCostsThreeLessWithNineCardsInTheGraveyard(t *testing.T) {
	for _, tc := range []struct {
		cards int
		pool  int
		ok    bool
	}{{9, 3, true}, {8, 3, false}, {8, 6, true}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		for i := 0; i < tc.cards; i++ {
			gyCard(me, "Filler", "Sorcery", "", "")
		}
		fillPoolColored(me, "R", 3)
		fillPool(me, tc.pool-3)
		err := s82OneB19CastStrict(t, g, "Capricious Hellraiser", "Creature — Phyrexian Dragon", s82OneHellraiser, "{3}{R}{R}{R}", 0)
		if (err == nil) != tc.ok {
			t.Errorf("%d cards in the graveyard, %d mana: err %v, want ok=%v", tc.cards, tc.pool, err, tc.ok)
		}
	}
}

func TestS82OneHellraiserExilesThreeAndCopiesANoncreatureNonlandCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Graveyard.Cards = nil
	shock := gyCard(me, "Shock", "Instant", "{R}", shockOracleForCopies)
	bear := gyCard(me, "Bear", "Creature — Bear", "{1}{G}", "")
	forest := gyCard(me, "Forest", "Basic Land — Forest", "", "")
	extra := gyCard(me, "Extra", "Sorcery", "{1}", "")
	_ = extra
	advanceToMain(t, g)
	castFromHandForTest(t, g, me, "Capricious Hellraiser", "Creature — Phyrexian Dragon", "{3}{R}{R}{R}", s82OneHellraiser, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	exiledFromYard := 0
	for _, id := range []uuid.UUID{shock, bear, forest, extra} {
		if inExile(g, id) {
			exiledFromYard++
		}
	}
	if exiledFromYard != 3 {
		t.Fatalf("%d graveyard cards exiled, want exactly three", exiledFromYard)
	}
	if !inExile(g, shock) {
		// The random draw left the Shock behind: nothing to copy unless
		// another noncreature, nonland card was drawn.
		t.Skip("the seeded random stream did not draw the Shock")
	}
	if p := chooseCardsChoiceFor(g, me.ID); p != nil {
		answerChooseCards(t, g, me.ID, shock)
	}
	cp := copyOf(t, g, "Shock")
	before := opp.Life
	if err := castCopy(g, me.ID, cp.InstanceID, true, playerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the copy for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life %d -> %d, want -2 from the copied Shock", before, opp.Life)
	}
	if !inExile(g, shock) {
		t.Error("the original left exile")
	}
}

func TestS82OneHellraiserWithOnlyCreaturesAndLandsCopiesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Graveyard.Cards = nil
	gyCard(me, "Bear", "Creature — Bear", "{1}{G}", "")
	gyCard(me, "Forest", "Basic Land — Forest", "", "")
	advanceToMain(t, g)
	castFromHandForTest(t, g, me, "Capricious Hellraiser", "Creature — Phyrexian Dragon", "{3}{R}{R}{R}", s82OneHellraiser, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("asked to choose with no eligible card")
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies, want none", n)
	}
	if me.Graveyard.Size() != 0 {
		t.Errorf("%d cards left in the graveyard, want all exiled", me.Graveyard.Size())
	}
}

// --- Blade of Shared Souls --------------------------------------------

func TestS82OneBladeOfSharedSoulsMakesTheRebelACopyOfAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	giant := b27Push(g, me.ID, "Hill Giant", "Creature — Giant", "", "{3}{R}", 3, 3, "R")
	advanceToMain(t, g)
	blade := castCatalogSpell(t, g, "Blade of Shared Souls", "Artifact — Equipment", s82OneBladeOfSharedSouls, nil)
	passPriorityAroundTable(t, g)
	rebel := findBattlefieldByName(g, "Rebel")
	if rebel == uuid.Nil {
		t.Fatal("no Rebel was made")
	}
	if host := attachmentHostOf(t, g, blade); host.ID != rebel {
		t.Fatalf("the Blade is attached to %+v, want the Rebel", host)
	}
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, giant)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	aegisWantName(t, g, rebel, "Hill Giant", 3)
	// Moving the Blade ends the copy.
	equipTo(t, g, me.ID, blade, giant)
	if got := aegisCardNow(g, rebel).Name; got != "Rebel" {
		t.Errorf("the old host is still %q after the Blade moved", got)
	}
}

func TestS82OneBladeOfSharedSoulsMayDeclineTheCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	giant := b27Push(g, me.ID, "Hill Giant", "Creature — Giant", "", "{3}{R}", 3, 3, "R")
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Blade of Shared Souls", "Artifact — Equipment", s82OneBladeOfSharedSouls, nil)
	passPriorityAroundTable(t, g)
	rebel := findBattlefieldByName(g, "Rebel")
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, giant)
	}
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := aegisCardNow(g, rebel).Name; got != "Rebel" {
		t.Errorf("the Rebel is %q after declining", got)
	}
}

func TestS82OneBladeOfSharedSoulsCannotCopyACreatureYouDoNotControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, opp.ID, "Their Giant", "Creature — Giant", "", "{3}{R}", 3, 3, "R")
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Blade of Shared Souls", "Artifact — Equipment", s82OneBladeOfSharedSouls, nil)
	passPriorityAroundTable(t, g)
	rebel := findBattlefieldByName(g, "Rebel")
	if rebel == uuid.Nil {
		t.Fatal("no Rebel was made")
	}
	if latestPickTarget(g, me.ID) != nil || confirmForTriggerPrompt(g, me.ID) != nil {
		t.Error("the trigger asked although I control no other creature")
	}
	if got := aegisCardNow(g, rebel).Name; got != "Rebel" {
		t.Errorf("the Rebel is %q", got)
	}
}

// --- Chiss-Goria, Forge Tyrant ----------------------------------------

func s82OneB19ChissAttack(t *testing.T, g *game.Game) (me, opp *game.Player, artifact, nonArtifact uuid.UUID) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	chiss := b27Push(g, me.ID, "Chiss-Goria, Forge Tyrant", "Legendary Creature — Dragon", s82OneChissGoria, "{6}{R}{R}{R}", 5, 4, "R")
	s82OneB19Artifact(g, me.ID, "Mine One")
	s82OneB19Artifact(g, me.ID, "Mine Two")
	me.Library.Cards = nil
	seedLibraryTop(me, "Filler", "Sorcery", "")
	seedLibraryTop(me, "Filler", "Sorcery", "")
	nonArtifact = seedLibraryTop(me, "Big Spell", "Sorcery", "{1}")
	artifact = seedLibraryTop(me, "Big Rock", "Artifact", "{4}")
	seedLibraryTop(me, "Second Rock", "Artifact", "{1}")
	attackWithTrigger(t, g, chiss, opp.ID, uuid.Nil)
	return me, opp, artifact, nonArtifact
}

func TestS82OneChissGoriaExilesFiveAndLetsYouCastAnArtifactWithAffinity(t *testing.T) {
	g := newCatalogGame(t)
	me, _, rock, spell := s82OneB19ChissAttack(t, g)
	if !inExile(g, rock) || !inExile(g, spell) {
		t.Fatal("the top five cards were not exiled")
	}
	if me.Library.Size() != 0 {
		t.Errorf("%d cards left in the library, want the five exiled", me.Library.Size())
	}
	advanceTo(t, g, game.StepPostcombatMain)
	// The Big Rock costs {4}; two artifacts make it {2} (affinity).
	fillPool(me, 1)
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("cast with {1} though the rock costs {2} after affinity")
	}
	fillPool(me, 1)
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast with {2} (affinity for two artifacts): %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Big Rock") == uuid.Nil {
		t.Error("the Big Rock did not resolve")
	}
}

func TestS82OneChissGoriaLetsYouCastOnlyOneArtifactAndNoNonartifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, _, rock, spell := s82OneB19ChissAttack(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	fillPool(me, 10)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("cast a nonartifact spell through the permission")
	}
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the artifact: %v", err)
	}
	passPriorityAroundTable(t, g)
	var second uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.Name == "Second Rock" {
			second = c.InstanceID
		}
	}
	if second == uuid.Nil {
		t.Fatal("the second artifact is not in exile")
	}
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("cast a second artifact: the permission is for one spell")
	}
}

func TestS82OneChissGoriaAffinityAppliesOnlyToTheSpellCastThisWay(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _, _ := s82OneB19ChissAttack(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	// A {4} artifact from hand does not get the affinity; Chiss-Goria's own
	// affinity is for itself, so with a pool of {2} it is refused.
	fillPool(me, 2)
	if err := s82OneB19CastStrict(t, g, "Hand Rock", "Artifact", "", "{4}", 0); err == nil {
		t.Error("a {4} artifact cast from hand got affinity from Chiss-Goria's permission")
	}
}

func TestS82OneChissGoriaCostsOneLessPerArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	s82OneB19Artifact(g, me.ID, "Mine One")
	s82OneB19Artifact(g, me.ID, "Mine Two")
	advanceToMain(t, g)
	fillPoolColored(me, "R", 3)
	fillPool(me, 3)
	// {6}{R}{R}{R} with two artifacts is 7 mana; six is not enough.
	if err := s82OneB19CastStrict(t, g, "Chiss-Goria, Forge Tyrant", "Legendary Creature — Dragon", s82OneChissGoria, "{6}{R}{R}{R}", 0); err == nil {
		t.Error("cast for six with two artifacts, want seven")
	}
	fillPool(me, 1)
	if err := g.CastSpell(me.ID, me.Hand.Cards[len(me.Hand.Cards)-1].InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("cast for seven with two artifacts: %v", err)
	}
}

// --- Ixhel, Scion of Atraxa -------------------------------------------

func TestS82OneIxhelExilesFromEachCorruptedOpponentAndLetsYouPlayThemWithAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me, poisoned, clean := g.Seats[0], g.Seats[1], g.Seats[2]
	b27Push(g, me.ID, "Ixhel, Scion of Atraxa", "Legendary Creature — Phyrexian Angel", s82OneIxhel, "{1}{W}{B}{G}", 2, 5, "W", "B", "G")
	assertKeywords(t, g, findBattlefieldByName(g, "Ixhel, Scion of Atraxa"), "flying", "vigilance")
	poisoned.Poison = 3
	clean.Poison = 2
	poisoned.Library.Cards = nil
	clean.Library.Cards = nil
	seedLibraryTop(poisoned, "Rocky Top", "Instant", "{R}{R}")
	seedLibraryTop(clean, "Safe Top", "Artifact", "{1}")
	topA := poisoned.Library.Cards[len(poisoned.Library.Cards)-1].InstanceID
	cleanSize := clean.Library.Size()
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if !inExile(g, topA) {
		t.Fatal("the corrupted opponent's top card was not exiled")
	}
	if clean.Library.Size() != cleanSize {
		t.Error("an opponent with fewer than three poison counters exiled a card")
	}
	// Ixhel's controller may cast it with green mana for a red cost.
	fillPoolColored(me, "G", 2)
	if err := g.CastSpell(me.ID, topA, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the exiled card with any colour of mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !poisoned.Graveyard.Contains(topA) {
		t.Error("the card did not resolve")
	}
}

func TestS82OneIxhelDoesNothingWhenNoOpponentIsCorrupted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Ixhel, Scion of Atraxa", "Legendary Creature — Phyrexian Angel", s82OneIxhel, "{1}{W}{B}{G}", 2, 5, "W", "B", "G")
	sizes := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		sizes[i] = p.Library.Size()
	}
	g.Seats[1].Poison = 2
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Library.Size() != sizes[i] {
			t.Errorf("seat %d's library changed from %d to %d", i, sizes[i], p.Library.Size())
		}
	}
}

func TestS82OneIxhelOnlyTriggersOnItsControllersEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Ixhel, Scion of Atraxa", "Legendary Creature — Phyrexian Angel", s82OneIxhel, "{1}{W}{B}{G}", 2, 5, "W", "B", "G")
	opp.Poison = 5
	// Seat 0's own end step legitimately exiles from the opponent; start
	// counting once seat 1's turn begins.
	for i := 0; i < 400 && g.Turn.ActiveSeat != 1; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	size := opp.Library.Size()
	for i := 0; i < 400 && !(g.Turn.ActiveSeat == 1 && g.Turn.Step == game.StepEnd); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	// Drawing on seat 1's turn takes one card; nothing else may leave.
	if opp.Library.Size() < size-1 {
		t.Errorf("library %d -> %d: Ixhel triggered on another player's end step", size, opp.Library.Size())
	}
}
