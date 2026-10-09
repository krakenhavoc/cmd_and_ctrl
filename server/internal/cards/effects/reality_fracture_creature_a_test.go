package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_creature_a_test.go — slice fra-creature-a (tracker
// #2795): twenty Reality Fracture creatures, one or more tests each.

const (
	oracleAerid           = "172a8f58-c420-4497-b962-c5853f923667"
	oracleAfterthought    = "f0c533a2-751a-4c5a-8cec-b83da3c8c62f"
	oracleApexWitchstalk  = "253d0cf7-65d3-4f13-aaf2-aae5bb60ac46"
	oracleArchiveArbiter  = "36cb3d18-019c-464d-b591-8c87e1b81f31"
	oracleArniScribe      = "a680d1ab-99f2-43d0-b289-5fdba4083d03"
	oracleArniChampion    = "1a929c4f-ae1f-44d9-a0a7-2ca2dc928dee"
	oracleBlessedGhoul    = "fe99d90e-4809-4310-a3f7-ceb540511d55"
	oracleBloombrute      = "b7040175-7411-43a8-9b46-dba1d7c2135d"
	oracleBuddingInsurg   = "723669e0-4a04-45ec-8c76-b1a369aa6892"
	oracleChandrasEmber   = "d3576617-7867-4c3d-aa31-d6742483465b"
	oracleCraterclaw      = "71355b51-e70e-47b0-8460-933af3567141"
	oracleCryotheoryAdept = "a9b3ee21-1af1-4e7e-8e48-d7c90d056f5c"
	oracleCurseMarred     = "f30a084c-5705-40f5-8e0d-2083d33973c7"
	oracleDanithaSpear    = "4b5c7f11-75f2-4bad-86cd-f0232861f569"
	oracleDanithaSword    = "f7b754d3-9863-420d-895d-00e70b8b078a"
	oracleDarkMatterManip = "6a3862f0-8dfc-4ac9-8eb9-faefcad685d4"
	oracleDarklightPhoen  = "e4b51c2d-7e56-435b-89d6-85e33df65879"
	oracleDarksteelAngel  = "d0259f7d-9bd7-4722-8fb9-439768308516"
	oracleDenzilore       = "0f8de7f0-c61d-415d-b4db-4959374343da"
)

// rfCrACast puts a creature card with printed P/T in the active seat's
// hand, walks to the precombat main phase and casts it free. castCatalogSpell
// gives a card no P/T, which a creature that must survive cannot use.
func rfCrACast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// rfCrAPush seats a permanent on the active seat's battlefield.
func rfCrAPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func rfCrACounters(g *game.Game, id uuid.UUID, kind string) int {
	c, ok := battlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.Counters[kind]
}

func rfCrADestroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(id, game.DestroyOptions{}) })
	if err != nil {
		t.Fatalf("destroy: %v", err)
	}
}

func rfCrAPower(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	var p int
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("%s not found", id)
		}
		p = c.CurrentPower()
	})
	return p
}

func rfCrAAddMana(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
}

// --- Aerid Konstrari -----------------------------------------------

func TestAeridKonstrariMakesAHeartwoodWhenItEntersAndWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	rfCrACast(t, g, "Aerid Konstrari", "Legendary Creature — Elder Sphinx", oracleAerid, 5, 4)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Heartwood"); n != 1 {
		t.Fatalf("%d Heartwood tokens after entering, want 1", n)
	}
	var aerid uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Aerid Konstrari" {
			aerid = c.InstanceID
		}
	}
	rfCrADestroy(t, g, aerid)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Heartwood"); n != 2 {
		t.Fatalf("%d Heartwood tokens after dying, want 2", n)
	}
}

func TestHeartwoodTokenIsARedGreenArtifactThatTapsForRedOrGreen(t *testing.T) {
	tok := HeartwoodToken()
	if tok.Name != "Heartwood" || tok.TypeLine != "Token Artifact — Heartwood" {
		t.Fatalf("token printed as %q / %q", tok.Name, tok.TypeLine)
	}
	if len(tok.Colors) != 2 {
		t.Errorf("colours %v, want red and green", tok.Colors)
	}
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var made []uuid.UUID
	g.WithWriteLock(func() {
		made, _ = g.CreateTokensForEffect(me.ID, HeartwoodToken(), 1, game.TokenEntryOptions{})
	})
	if len(made) != 1 {
		t.Fatalf("made %d tokens", len(made))
	}
	c, _ := battlefieldCard(g, made[0])
	if abs := game.ManaAbilitiesForCard(c); len(abs) != 1 || abs[0].Produced != "{R|G}" {
		t.Fatalf("mana abilities %+v, want one {R|G}", abs)
	}
}

func TestAeridKonstrariSixManaMakesATokenThenPumpsByArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	aerid := rfCrAPush(g, me.ID, "Aerid Konstrari", "Legendary Creature — Elder Sphinx", oracleAerid, 5, 4)
	rfCrAPush(g, me.ID, "Rock", "Artifact", "", 0, 0)
	rfCrAAddMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	b16Activate(t, g, me.ID, aerid, 0, game.ActivateAbilityParams{})
	if n := b16CountNamed(g, "Heartwood"); n != 1 {
		t.Fatalf("%d Heartwood tokens, want 1", n)
	}
	// Rock + the new Heartwood: X = 2, counted after the token exists.
	if got := rfCrAPower(t, g, aerid); got != 7 {
		t.Fatalf("power %d, want 7 (5 + 2 artifacts)", got)
	}
}

// --- Archive Arbiter -----------------------------------------------

func TestArchiveArbiterGainsFourLifeMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	rfCrACast(t, g, "Archive Arbiter", "Artifact Creature — Sphinx", oracleArchiveArbiter, 4, 4)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Fatalf("life %d, want %d", me.Life, life+4)
	}
}

func TestArchiveArbiterDestroysANoncreatureNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := rfCrAPush(g, opp.ID, "Rock", "Artifact", "", 0, 0)
	bear := rfCrAPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	land := rfCrAPush(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	rfCrACast(t, g, "Archive Arbiter", "Artifact Creature — Sphinx", oracleArchiveArbiter, 4, 4)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	// The creature and the land are not legal picks.
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the destroy mode")
	}
	for _, bad := range []uuid.UUID{bear, land} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was accepted as a target", bad)
		}
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Fatal("the artifact survived")
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(land) {
		t.Fatal("the creature or land was destroyed")
	}
}

// --- Chandra's Emberling -------------------------------------------

func TestChandrasEmberlingGrowsOnNoncreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	em := rfCrAPush(g, me.ID, "Chandra's Emberling", "Creature — Gremlin Elemental", oracleChandrasEmber, 2, 2)
	castCatalogSpell(t, g, "Bolt", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := rfCrACounters(g, em, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters after a noncreature spell: %d, want 1", got)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := rfCrACounters(g, em, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters after a creature spell: %d, want 1", got)
	}
}

// --- Craterclaw Colossus -------------------------------------------

func TestCraterclawColossusGivesTrampleAndPowerPerArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := rfCrAPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	rfCrAPush(g, me.ID, "Rock", "Artifact", "", 0, 0)
	rfCrAPush(g, me.ID, "Rock", "Artifact", "", 0, 0)
	colossus := rfCrACast(t, g, "Craterclaw Colossus", "Artifact Creature — Beast Construct", oracleCraterclaw, 5, 5)
	passPriorityAroundTable(t, g)
	// Two Rocks and the Colossus itself.
	if got := rfCrAPower(t, g, bear); got != 5 {
		t.Fatalf("Bear power %d, want 5 (2 + 3 artifacts)", got)
	}
	if !effectiveAbilitiesContain(t, g, bear, "trample") {
		t.Fatal("Bear has no trample")
	}
	if got := rfCrAPower(t, g, colossus); got != 8 {
		t.Fatalf("Colossus power %d, want 8", got)
	}
	late := rfCrAPush(g, me.ID, "Late Bear", "Creature — Bear", "", 2, 2)
	if got := rfCrAPower(t, g, late); got != 2 {
		t.Fatalf("a creature that arrived later has power %d, want 2", got)
	}
}

// --- Curse-Marred Demon --------------------------------------------

func TestCurseMarredDemonTutorsThenDiscardsAtRandom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	me.Library.Cards = nil
	prize := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: prize, Name: "Prize", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	rfCrACast(t, g, "Curse-Marred Demon", "Creature — Demon", oracleCurseMarred, 4, 4)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 || !me.Graveyard.Contains(prize) {
		t.Fatalf("hand %d, graveyard has prize %v: want the tutored card pitched", me.Hand.Size(), me.Graveyard.Contains(prize))
	}
}

// --- Apex Witchstalker ---------------------------------------------

func TestApexWitchstalkerGainsTwoOnEnteringAndOnDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	id := rfCrACast(t, g, "Apex Witchstalker", "Creature — Wolf", oracleApexWitchstalk, 6, 4)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Fatalf("life %d after entering, want %d", me.Life, life+2)
	}
	rfCrADestroy(t, g, id)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Fatalf("life %d after dying, want %d", me.Life, life+4)
	}
}

func TestApexWitchstalkerBasicLandcycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id, _ := cycleFromHand(t, g, "Apex Witchstalker", "Creature — Wolf", oracleApexWitchstalk, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		// One legal match resolves without a prompt.
		if !me.Hand.Contains(ids[0]) {
			t.Fatal("no basic land reached the hand")
		}
		return
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Fatal("no basic land reached the hand")
	}
}

// --- Afterthought Sentry -------------------------------------------

func TestAfterthoughtSentryGainsFlyingForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	s := rfCrAPush(g, me.ID, "Afterthought Sentry", "Artifact Creature — Gargoyle", oracleAfterthought, 2, 2)
	if effectiveAbilitiesContain(t, g, s, "flying") {
		t.Fatal("flying before activating")
	}
	rfCrAAddMana(t, g, me, "{C}{C}")
	b16Activate(t, g, me.ID, s, 0, game.ActivateAbilityParams{})
	if !effectiveAbilitiesContain(t, g, s, "flying") {
		t.Fatal("no flying after activating")
	}
}

func TestAfterthoughtSentryAttackExilesAGraveyardCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	s := rfCrAPush(g, me.ID, "Afterthought Sentry", "Artifact Creature — Gargoyle", oracleAfterthought, 2, 2)
	victim := pushGraveyardCardForTest(opp, "Doomed")
	declareAttack(t, g, opp.ID, s)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
			t.Fatalf("pick: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if !g.Exile.Contains(victim) {
		t.Fatal("the graveyard card was not exiled")
	}
}

// --- Dark Matter Manipulator ---------------------------------------

func TestDarkMatterManipulatorMillsThreeAndGrowsPerSevenGraveyardCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Graveyard.Cards = nil
	for i := 0; i < 20; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	}
	id := rfCrACast(t, g, "Dark Matter Manipulator", "Creature — Human Warlock", oracleDarkMatterManip, 1, 2)
	passPriorityAroundTable(t, g)
	if got := len(me.Graveyard.Cards); got != 3 {
		t.Fatalf("graveyard %d after the mill, want 3", got)
	}
	if got := rfCrAPower(t, g, id); got != 1 {
		t.Fatalf("power %d with 3 cards, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 4) })
	if got := rfCrAPower(t, g, id); got != 3 {
		t.Fatalf("power %d with 7 cards, want 3", got)
	}
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 7) })
	if got := rfCrAPower(t, g, id); got != 5 {
		t.Fatalf("power %d with 14 cards, want 5", got)
	}
}
