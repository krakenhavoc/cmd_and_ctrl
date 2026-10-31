package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_6_cards_test.go — Marvel Super Heroes Commander, sixth batch
// (#4455). Each test plays the real catalog entry and asserts what its
// text says.

const (
	msc6DivineVisitationOracle = "5917216f-5b42-41fa-8976-401bdbeb782e"
	msc6DocOckOracle           = "dbaa70e6-26de-4ce9-9543-e652cbe387c1"
	msc6BugleOracle            = "2e677ded-7af3-4af0-81df-b70d6ee616ec"
	msc6JaneOracle             = "b1c8fddb-d686-4b4a-8aff-d79ea747801c"
	msc6NefariaOracle          = "5ec4f81e-ede4-4d45-b768-16583f6f5b8d"
	msc6DaredevilOracle        = "9436a736-9139-4d8a-9eb0-9dbf706da51c"
	msc6CrystalOracle          = "32c14d7f-fdaa-4340-af94-1f10bf064f51"
	msc6ChaseOracle            = "161ff063-b784-43c7-b608-f78be3daa5b2"
	msc6CommuneOracle          = "27ca27f0-95a5-4292-b30b-1db3e9d639f8"
	msc6WaveOracle             = "25012c1f-c069-43d7-b930-a3a5a7eb1389"
	msc6DoomOracle             = "9c9cd0fd-bfe9-409c-a9a9-146f5f4a2ef4"
)

func TestMsc6EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc6DivineVisitationOracle: "Divine Visitation",
		msc6DocOckOracle:           "Doc Ock, Sinister Scientist",
		msc6BugleOracle:            "Daily Bugle Reporters",
		msc6JaneOracle:             "Doctor Jane Foster",
		msc6NefariaOracle:          "Count Nefaria",
		msc6DaredevilOracle:        "Daredevil, Fearless Fighter",
		msc6CrystalOracle:          "Crystal, Inhuman Princess",
		msc6ChaseOracle:            "Chase Stein, Runaway",
		msc6CommuneOracle:          "Commune with Dinosaurs",
		msc6WaveOracle:             "Dismantling Wave",
		msc6DoomOracle:             "Doctor Doom, King of Latveria",
		msc6ZemoOracleID:           "Helmut Zemo, Mastermind",
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

// msc6SettleOpponentTargets answers every "target opponent" prompt with
// opp and passes until the table is quiet.
func msc6SettleOpponentTargets(t *testing.T, g *game.Game, me, opp uuid.UUID) {
	t.Helper()
	for i := 0; i < 6; i++ {
		passPriorityAroundTable(t, g)
		if latestPickTarget(g, me) == nil {
			return
		}
		b16PickPlayer(t, g, me, opp)
	}
}

func TestMsc6DivineVisitationMakesAngelsInsteadOfYourCreatureTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Divine Visitation", "Enchantment", msc6DivineVisitationOracle, 0, 0)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 white Soldier"), 2)
		_ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1)
		_ = g.CreateTokenForEffect(opp.ID, TokenCard("1/1 white Soldier"), 1)
	})
	if n := countBattlefieldNamed(g, me.ID, "Angel"); n != 2 {
		t.Errorf("two Soldier tokens became %d Angels, want 2", n)
	}
	if n := countBattlefieldNamed(g, me.ID, "Soldier"); n != 0 {
		t.Errorf("%d Soldiers were made anyway", n)
	}
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("the Treasure is not a creature token and stays: %d", n)
	}
	if n := countBattlefieldNamed(g, opp.ID, "Soldier"); n != 1 {
		t.Errorf("an opponent's Soldier is untouched: %d", n)
	}
	angel := findBattlefieldByName(g, "Angel")
	assertKeywords(t, g, angel, "flying", "vigilance")
	if p, tg := effectivePower(t, g, angel), effectiveToughness(t, g, angel); p != 4 || tg != 4 {
		t.Errorf("the Angel is %d/%d, want 4/4", p, tg)
	}
}

func TestMsc6DocOckIsEightEightWithEightCardsInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ock := b12Push(g, me.ID, "Doc Ock, Sinister Scientist", "Legendary Creature — Human Scientist Villain", msc6DocOckOracle, 4, 5)
	for i := 0; i < 7; i++ {
		pushGraveyardCardForTest(me, "Filler")
	}
	if p, tg := effectivePower(t, g, ock), effectiveToughness(t, g, ock); p != 4 || tg != 5 {
		t.Errorf("with seven cards Doc Ock is %d/%d, want 4/5", p, tg)
	}
	// The eighth card arrives by a real zone move, which drops the layer cache.
	b27Kill(g, b12Creature(g, me.ID, "Doomed Bear", "Creature — Bear", 1, 1))
	if p, tg := effectivePower(t, g, ock), effectiveToughness(t, g, ock); p != 8 || tg != 8 {
		t.Errorf("with eight cards Doc Ock is %d/%d, want 8/8", p, tg)
	}
	// An opponent's graveyard does not count.
	g2 := newCatalogGame(t)
	ock2 := b12Push(g2, g2.Seats[0].ID, "Doc Ock, Sinister Scientist", "Legendary Creature — Human Scientist Villain", msc6DocOckOracle, 4, 5)
	for i := 0; i < 8; i++ {
		pushGraveyardCardForTest(g2.Seats[1], "Filler")
	}
	if p := effectivePower(t, g2, ock2); p != 4 {
		t.Errorf("their graveyard made Doc Ock %d power", p)
	}
}

func TestMsc6DocOckHasHexproofOnlyWithAnotherVillain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ock := b12Push(g, me.ID, "Doc Ock, Sinister Scientist", "Legendary Creature — Human Scientist Villain", msc6DocOckOracle, 4, 5)
	if containsString(effectiveAbilities(t, g, ock), "hexproof") {
		t.Fatal("alone, Doc Ock has hexproof")
	}
	b12Creature(g, me.ID, "Henchman", "Creature — Human Villain", 1, 1)
	assertKeywords(t, g, ock, "hexproof")

	g2 := newCatalogGame(t)
	ock2 := b12Push(g2, g2.Seats[0].ID, "Doc Ock, Sinister Scientist", "Legendary Creature — Human Scientist Villain", msc6DocOckOracle, 4, 5)
	b12Creature(g2, g2.Seats[1].ID, "Their Villain", "Creature — Human Villain", 1, 1)
	if containsString(effectiveAbilities(t, g2, ock2), "hexproof") {
		t.Error("an opponent's Villain gave Doc Ock hexproof")
	}
}

func TestMsc6DailyBugleReportersPutsCountersOnUpToTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)

	castAndResolveCreature(t, g, "Daily Bugle Reporters", "Creature — Human Citizen", msc6BugleOracle)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	answerTriggerTargets(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	if got := b18Counters(t, g, a, "+1/+1"); got != 1 {
		t.Errorf("Bear A has %d counters, want 1", got)
	}
	if got := b18Counters(t, g, b, "+1/+1"); got != 1 {
		t.Errorf("Bear B has %d counters, want 1", got)
	}
	if got := b18Counters(t, g, c, "+1/+1"); got != 0 {
		t.Errorf("Bear C has %d counters, want 0", got)
	}
}

func TestMsc6DailyBugleReportersReturnsASmallCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	small := gyCard(me, "Small Friend", "Creature — Human", "{1}{W}", "")
	big := gyCard(me, "Big Friend", "Creature — Giant", "{2}{W}", "")

	castAndResolveCreature(t, g, "Daily Bugle Reporters", "Creature — Human Citizen", msc6BugleOracle)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
		t.Error("a mana value 3 creature card was accepted")
	}
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(small) || me.Graveyard.Contains(small) {
		t.Error("the mana value 2 creature card did not return to hand")
	}
	if !me.Graveyard.Contains(big) {
		t.Error("the larger card left the graveyard")
	}
}

func TestMsc6DoctorJaneFosterReturnsToHandOrToTheBattlefieldAfterLifegain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	friend := gyCard(me, "Friend", "Creature — Human", "{2}{W}", "")
	big := gyCard(me, "Giant", "Creature — Giant", "{3}{W}", "")
	castHoldCreature(t, g, "Doctor Jane Foster", "Legendary Creature — Human Doctor", msc6JaneOracle, 2, 3)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
		t.Error("a mana value 4 creature card was accepted")
	}
	pickCard(t, g, me.ID, friend)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(friend) || g.Battlefield.Contains(friend) {
		t.Error("with no lifegain the card goes to hand")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	friend2 := gyCard(me2, "Friend", "Creature — Human", "{2}{W}", "")
	g2.WithWriteLock(func() { _ = g2.ChangePlayerLifeForEffect(uuid.Nil, me2.ID, 2) })
	castHoldCreature(t, g2, "Doctor Jane Foster", "Legendary Creature — Human Doctor", msc6JaneOracle, 2, 3)
	pickCard(t, g2, me2.ID, friend2)
	passPriorityAroundTable(t, g2)
	if !g2.Battlefield.Contains(friend2) || me2.Hand.Contains(friend2) {
		t.Error("after gaining life the card returns to the battlefield")
	}
}

func TestMsc6CountNefariaCostsThreeLessAfterYouSacrificeAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := selfPricedMV(t, g, me, msc6NefariaOracle, "Legendary Creature — Human Villain", "{6}{B}"); got != 7 {
		t.Fatalf("Count Nefaria before a sacrifice: %d, want 7", got)
	}
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	if err := g.SacrificePermanent(me.ID, rock); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	if got := selfPricedMV(t, g, me, msc6NefariaOracle, "Legendary Creature — Human Villain", "{6}{B}"); got != 4 {
		t.Errorf("Count Nefaria after a sacrifice: %d, want 4", got)
	}

	// An opponent's sacrifice does not count for you.
	g2 := newCatalogGame(t)
	opp2 := g2.Seats[1]
	theirs := b12Permanent(g2, opp2.ID, "Their Rock", "Artifact")
	if err := g2.SacrificePermanent(opp2.ID, theirs); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	if got := selfPricedMV(t, g2, g2.Seats[0], msc6NefariaOracle, "Legendary Creature — Human Villain", "{6}{B}"); got != 7 {
		t.Errorf("Count Nefaria after an opponent's sacrifice: %d, want 7", got)
	}
}

func TestMsc6CountNefariaHasFlying(t *testing.T) {
	g := newCatalogGame(t)
	nef := b12Push(g, g.Seats[0].ID, "Count Nefaria", "Legendary Creature — Human Villain", msc6NefariaOracle, 6, 6)
	assertKeywords(t, g, nef, "flying")
}

func TestMsc6DaredevilReflectsDamageDealtToYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Daredevil, Fearless Fighter", "Legendary Creature — Human Hero", msc6DaredevilOracle, 3, 3)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// An opponent's source hurting me does nothing.
	before := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, me.ID, 2) })
	msc6SettleOpponentTargets(t, g, me.ID, opp.ID)
	if opp.Life != before {
		t.Errorf("their own damage was reflected: %d -> %d", before, opp.Life)
	}

	// A source I control hurting me reflects that much to the target opponent.
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(mine, me.ID, 3) })
	msc6SettleOpponentTargets(t, g, me.ID, opp.ID)
	if me.Life != life-3 {
		t.Fatalf("I took %d, want 3", life-me.Life)
	}
	if opp.Life != before-3 {
		t.Errorf("the opponent took %d, want 3", before-opp.Life)
	}
}

func TestMsc6DaredevilAttackExilesTheTopCardAndHurtsYouForItsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dd := b12Push(g, me.ID, "Daredevil, Fearless Fighter", "Legendary Creature — Human Hero", msc6DaredevilOracle, 3, 3)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Costly Spell", TypeLine: "Sorcery", ManaCost: "{2}{R}"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	life, oppLife := me.Life, opp.Life
	declareAttack(t, g, opp.ID, dd)
	msc6SettleOpponentTargets(t, g, me.ID, opp.ID)

	if !g.Exile.Contains(ids[0]) {
		t.Fatal("the top card was not exiled")
	}
	if me.Life != life-3 {
		t.Errorf("Daredevil dealt %d to me, want the card's mana value 3", life-me.Life)
	}
	// Its own damage to me is damage from a source I control: reflected.
	if opp.Life != oppLife-3 {
		t.Errorf("the opponent took %d, want 3 reflected", oppLife-opp.Life)
	}
	perm := exiledPermission(g, ids[0])
	if perm.Player != me.ID || !permissionLive(g, perm, me.ID) {
		t.Errorf("%+v: I may play the exiled card this turn", perm)
	}
}

func TestMsc6DaredevilAttackWithALandCostsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dd := b12Push(g, me.ID, "Daredevil, Fearless Fighter", "Legendary Creature — Human Hero", msc6DaredevilOracle, 3, 3)
	ids := seedSearchLibrary(me, game.Card{Name: "Top Land", TypeLine: "Basic Land — Mountain"})
	life := me.Life
	declareAttack(t, g, opp.ID, dd)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[0]) {
		t.Fatal("the top card was not exiled")
	}
	if me.Life != life {
		t.Errorf("a land cost me %d life", life-me.Life)
	}
}

func TestMsc6CrystalDamagesEachOpponentForTheSpellsColors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Crystal, Inhuman Princess", "Legendary Creature — Inhuman Noble Hero", msc6CrystalOracle, 2, 3)
	toMainForCost(t, g)

	lives := func() [3]int {
		return [3]int{g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life}
	}
	start := lives()

	gold := handCardFull(me, "Gold Spell", "Instant", "{R}{G}", "", []string{"R", "G"})
	floatForTest(g, me, "RG")
	if err := g.CastSpell(me.ID, gold, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the two-color spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	for i, l := range lives() {
		if l != start[i]-2 {
			t.Errorf("opponent %d took %d, want 2", i+1, start[i]-l)
		}
	}

	rock := handCardFull(me, "Plain Rock", "Artifact", "{2}", "", nil)
	floatForTest(g, me, "CC")
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the colorless spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	for i, l := range lives() {
		if l != start[i]-2 {
			t.Errorf("a colorless spell hurt opponent %d: %d", i+1, start[i]-l)
		}
	}

	bear := handCardFull(me, "Green Bear", "Creature — Bear", "{1}{G}", "", []string{"G"})
	floatForTest(g, me, "GG")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the creature spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	for i, l := range lives() {
		if l != start[i]-2 {
			t.Errorf("a creature spell hurt opponent %d: %d", i+1, start[i]-l)
		}
	}
}

func TestMsc6CrystalTapsForOneOfFourColors(t *testing.T) {
	for _, colour := range []string{"R", "G", "W", "U"} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		crystal := b12Push(g, me.ID, "Crystal, Inhuman Princess", "Legendary Creature — Inhuman Noble Hero", msc6CrystalOracle, 2, 3)
		b28TapForMana(t, g, me.ID, crystal, colour)
		if got := poolColors(me); len(got) != 1 || got[0] != colour {
			t.Errorf("tapped for {%s}: pool %v", colour, got)
		}
	}
	g := newCatalogGame(t)
	crystal := b12Push(g, g.Seats[0].ID, "Crystal, Inhuman Princess", "Legendary Creature — Inhuman Noble Hero", msc6CrystalOracle, 2, 3)
	assertKeywords(t, g, crystal, "flying")
}

func TestMsc6ChaseSteinDiscardsToExileAndPlayTheTopCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chase := b12Push(g, me.ID, "Chase Stein, Runaway", "Legendary Creature — Human Hero", msc6ChaseOracle, 2, 2)
	discard := handCardForTest(me, "Pitched Card", "Sorcery", "")
	ids := seedSearchLibrary(me,
		game.Card{Name: "Top Card", TypeLine: "Sorcery"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	b16Activate(t, g, me.ID, chase, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{discard}})
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(discard) {
		t.Error("the discarded card is not in the graveyard")
	}
	if !g.Exile.Contains(ids[0]) || !me.Library.Contains(ids[1]) {
		t.Fatal("only the top card is exiled")
	}
	if c, _ := battlefieldCard(g, chase); !c.Tapped {
		t.Error("Chase Stein did not tap")
	}
	perm := exiledPermission(g, ids[0])
	if perm.Player != me.ID || perm.CastOnly || !permissionLive(g, perm, me.ID) {
		t.Errorf("%+v: I may play the exiled card", perm)
	}
	advanceToMainOf(t, g, 2)
	if !permissionLive(g, exiledPermission(g, ids[0]), me.ID) {
		t.Error("the grant survives the opponents' turns")
	}
	advanceToMainOf(t, g, 1)
	if permissionLive(g, exiledPermission(g, ids[0]), me.ID) {
		t.Error("the grant ends with my next turn")
	}
}

func TestMsc6ChaseSteinNeedsACardToDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chase := b12Push(g, me.ID, "Chase Stein, Runaway", "Legendary Creature — Human Hero", msc6ChaseOracle, 2, 2)
	seedSearchLibrary(me, game.Card{Name: "Top Card", TypeLine: "Sorcery"})
	me.Hand.Cards = nil
	if err := g.ActivateCatalogAbility(me.ID, chase, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated with nothing to discard")
	}
}

func TestMsc6CommuneWithDinosaursTakesADinosaurOrALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	var top [5]uuid.UUID
	cards := []struct{ name, typeLine string }{
		{"Raptor", "Creature — Dinosaur"},
		{"Bear", "Creature — Bear"},
		{"Forest", "Basic Land — Forest"},
		{"Shock", "Instant"},
		{"Wall", "Creature — Wall"},
	}
	for i := len(cards) - 1; i >= 0; i-- {
		top[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: top[i], Name: cards[i].name, TypeLine: cards[i].typeLine, Owner: me.ID, Controller: me.ID})
	}
	castCatalogSpell(t, g, "Commune with Dinosaurs", "Sorcery", msc6CommuneOracle, nil)
	passPriorityAroundTable(t, g)

	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick opened")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if len(offered) != 2 || !offered[top[0]] || !offered[top[2]] || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the Dinosaur and the land, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[0]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[0]) {
		t.Fatal("the Dinosaur is not in hand")
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no order prompt for the rest on the bottom")
	}
}

func TestMsc6CommuneWithDinosaursMayTakeNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := seedLibrary(me, "A", "B", "C", "D", "E")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Commune with Dinosaurs", "Sorcery", msc6CommuneOracle, nil)
	passPriorityAroundTable(t, g)
	if latestChooseCards(g, me.ID) != nil {
		answerChooseCards(t, g, me.ID)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d with nothing eligible", hand, me.Hand.Size())
	}
	_ = top
}

func TestMsc6DismantlingWaveDestroysOneArtifactOrEnchantmentPerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := b12Permanent(g, me.ID, "My Rock", "Artifact")
	aRock := b12Permanent(g, a.ID, "A's Rock", "Artifact")
	aSecond := b12Permanent(g, a.ID, "A's Aura", "Enchantment")
	bAura := b12Permanent(g, b.ID, "B's Enchantment", "Enchantment")
	bBear := b12Creature(g, b.ID, "B's Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Dismantling Wave", "Sorcery", msc6WaveOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: aRock}, {Kind: game.TargetCard, ID: bAura},
	})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aRock) || g.Battlefield.Contains(bAura) {
		t.Error("the named artifact and enchantment survived")
	}
	for name, id := range map[string]uuid.UUID{"my artifact": mine, "A's second permanent": aSecond, "B's creature": bBear} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed", name)
		}
	}
}

func TestMsc6DismantlingWaveCyclingDestroysEveryArtifactAndEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	seedLibrary(me, "Drawn")
	hand := me.Hand.Size()

	cycled, _ := cycleFromHand(t, g, "Dismantling Wave", "Sorcery", msc6WaveOracle, "{C}{C}{C}{C}{C}{C}{W}{W}")
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(cycled) {
		t.Error("the cycled card is not in the graveyard")
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("an artifact or enchantment survived the cycling trigger")
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(land) {
		t.Error("a creature or land was destroyed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("cycling drew %d cards", me.Hand.Size()-hand)
	}
}

func TestMsc6DoomKingDrainsOnceForADiscardOfLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Doctor Doom, King of Latveria", "Legendary Creature — Human Noble Villain", msc6DoomOracle, 3, 3)
	me.Hand.Cards = nil
	l1 := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	l2 := handCardForTest(me, "Mountain", "Basic Land — Mountain", "")
	spell := handCardForTest(me, "Sorcery", "Sorcery", "")
	before := [3]int{g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life}

	g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 2) })
	answerDiscard(t, g, me.ID, l1, l2)
	passPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ {
		if got := before[i] - g.Seats[i+1].Life; got != 2 {
			t.Errorf("opponent %d lost %d, want 2 (once for two lands)", i+1, got)
		}
	}

	g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 1) })
	answerDiscard(t, g, me.ID, spell)
	passPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ {
		if got := before[i] - g.Seats[i+1].Life; got != 2 {
			t.Errorf("a nonland discard drained opponent %d: total %d", i+1, got)
		}
	}
}

func TestMsc6DoomKingCombatGivesAVillainMenaceAndConnives(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Doctor Doom, King of Latveria", "Legendary Creature — Human Noble Villain", msc6DoomOracle, 3, 3)
	henchman := b12Creature(g, me.ID, "Henchman", "Creature — Human Villain", 1, 1)
	hero := b12Creature(g, me.ID, "Hero", "Creature — Human Hero", 1, 1)
	seedLibrary(me, "Drawn")
	pitch := handCardForTest(me, "Pitched Sorcery", "Sorcery", "")

	advanceTo(t, g, game.StepBeginCombat)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: hero}); err == nil {
		t.Error("a non-Villain was accepted")
	}
	pickCard(t, g, me.ID, henchman)
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, henchman, "menace")
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	if got := b18Counters(t, g, henchman, "+1/+1"); got != 1 {
		t.Errorf("a nonland discard put %d counters on the Villain, want 1", got)
	}
	if containsString(effectiveAbilities(t, g, hero), "menace") {
		t.Error("the Hero gained menace")
	}
}

func msc6ZemoCastFromGraveyard(g *game.Game, me *game.Player, id uuid.UUID) error {
	return g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true,
	})
}

func TestMsc6ZemoCastsASmallSpellFromTheGraveyardExilesItAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zemo := b12Push(g, me.ID, "Helmut Zemo, Mastermind", "Legendary Creature — Human Noble Villain", msc6ZemoOracleID, 3, 3)
	div := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)
	big := gyCard(me, "Big Spell", "Sorcery", "{3}{U}", "")
	seedLibrary(me, "A", "B", "C")

	declareAttack(t, g, opp.ID, zemo)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
		t.Fatal("a mana value 4 spell was accepted for a 3 power Zemo")
	}
	pickCard(t, g, me.ID, div)
	passPriorityAroundTable(t, g)

	floatForTest(g, me, "UUU")
	hand := me.Hand.Size()
	if err := msc6ZemoCastFromGraveyard(g, me, div); err != nil {
		t.Fatalf("cast Divination from the graveyard: %v", err)
	}
	castInsideTheResolution(t, g, div)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(div) || !inExile(g, div) {
		t.Error("the spell goes to exile instead of the graveyard")
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want Divination's two", got)
	}
	if got := b18Counters(t, g, zemo, "+1/+1"); got != 1 {
		t.Errorf("Zemo has %d counters, want 1 for the cast", got)
	}
}

func TestMsc6ZemoDecliningTheCastGivesNoCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zemo := b12Push(g, me.ID, "Helmut Zemo, Mastermind", "Legendary Creature — Human Noble Villain", msc6ZemoOracleID, 3, 3)
	div := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)

	declareAttack(t, g, opp.ID, zemo)
	pickCard(t, g, me.ID, div)
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, div)
	declineResolutionCast(t, g, me.ID)
	if !me.Graveyard.Contains(div) {
		t.Error("the declined card left the graveyard")
	}
	if got := b18Counters(t, g, zemo, "+1/+1"); got != 0 {
		t.Errorf("Zemo has %d counters after a declined cast", got)
	}
}
