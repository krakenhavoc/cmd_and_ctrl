package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pirates_batch2_test.go — the second Pirates batch: fourteen cards
// that needed no engine work, written after the impulse-exile and
// additional-cost sub-PRs opened them up.

const (
	captainStormOracle   = "431e85e3-15e6-471b-ba71-6058394c9a96"
	imperialRecruiterOrc = "4d6a1391-817a-4ddc-840d-886b138eeb3f"
	magmakinOracle       = "900b9409-9c16-414d-8674-2ea42c2415a1"
	malcolmNavigatorOrc  = "a66f8b44-0163-4456-b152-4acefab896a4"
	skyrayOracle         = "3a46d85b-ce1a-4842-a342-92a5bddb1053"
	weftstalkerOracle    = "926d52a5-4db1-46ce-9567-17c28bf56ae7"
	gemcutterOracle      = "68e45c07-96c5-4f87-a816-d9fa4f119740"
	solphimOracle        = "895f23a2-55b7-4cc0-8939-2efaaf097e6f"
	gambleOracle         = "a54f0869-94c8-42af-9080-166efb9486a4"
	windfallOracle       = "08becc07-28bc-4a2f-a6b0-28a2998d2f50"
	offerOracle          = "234a734b-ba28-4f1b-9d01-3c3e7d516590"
	timeLoopOracle       = "13e94530-defb-4c9b-9ec7-bb7789ec2630"
	pullFromTomorrowOrc  = "b1a23235-3076-475c-a68a-db29cf2a9dba"
	franticSearchOracle  = "16e015b2-f8a3-4b1a-80be-58a8f5fb5e8c"
)

// --- creatures ---------------------------------------------------

func TestCaptainStormCountersAPirateOnArtifactETB(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Captain Storm, Cosmium Raider",
		"Legendary Creature — Human Pirate", captainStormOracle, false)
	mako := pushCatalogPermanent(g, me.ID, "Marauding Mako", "Creature — Shark Pirate", maraudingMakoOracle, false)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	pickCard(t, g, me.ID, mako)
	passPriorityAroundTable(t, g)

	if got := countersOn(g, mako, "+1/+1"); got != 1 {
		t.Errorf("Mako counters = %d, want 1", got)
	}
}

func TestImperialRecruiterFetchesASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	big := uuid.New()
	me.Library.PushTop(game.Card{
		InstanceID: big, Name: "Big Guy", TypeLine: "Creature — Giant",
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	small := uuid.New()
	me.Library.PushTop(game.Card{
		InstanceID: small, Name: "Ragavan", TypeLine: "Creature — Monkey Pirate",
		Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	me.Hand.Cards = nil

	castCatalogSpell(t, g, "Imperial Recruiter", "Creature — Human Advisor", imperialRecruiterOrc, nil)
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(small) {
		t.Errorf("the power-2 creature should be in hand")
	}
	if me.Hand.Contains(big) {
		t.Errorf("a 5/5 is not a legal fetch")
	}
}

func TestMagmakinArtilleristBurnsOnEveryDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Magmakin Artillerist", "Creature — Elemental Pirate", magmakinOracle, false)
	before := lifeOfOpponents(g)
	fodder := handCard(me, "Fodder", "Sorcery")

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	_ = fodder

	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d -> %d, want -1", i+1, b, got)
		}
	}
}

func TestMalcolmMakesATreasureOnPirateDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Malcolm, Keen-Eyed Navigator",
		"Legendary Creature — Siren Pirate", malcolmNavigatorOrc, false)
	pirate := pushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)

	attackWith(t, g, victim.ID, pirate)
	passPriorityAroundTable(t, g)

	if countBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Errorf("a Pirate connecting should make one Treasure")
	}
}

func TestMalcolmIgnoresNonPirateDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Malcolm, Keen-Eyed Navigator",
		"Legendary Creature — Siren Pirate", malcolmNavigatorOrc, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	attackWith(t, g, victim.ID, bear)
	passPriorityAroundTable(t, g)

	if countBattlefieldNamed(g, me.ID, "Treasure") != 0 {
		t.Errorf("a non-Pirate should not make a Treasure")
	}
}

func TestScroungingSkyrayGrowsOnDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	skyray := pushCatalogPermanent(g, me.ID, "Scrounging Skyray", "Creature — Fish Pirate", skyrayOracle, false)
	handCard(me, "Fodder", "Sorcery")

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)

	if got := countersOn(g, skyray, "+1/+1"); got != 1 {
		t.Errorf("Skyray counters = %d, want 1", got)
	}
}

// "Another creature or artifact" — the Ardent's own arrival is not a
// trigger, and a land isn't either.
func TestWeftstalkerArdentPingsOnAnotherPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ardent := pushCatalogPermanent(g, me.ID, "Weftstalker Ardent", "Creature — Drix Artificer", weftstalkerOracle, false)
	before := lifeOfOpponents(g)

	// Its own ETB does nothing.
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: ardent}) })
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if g.Seats[i+1].Life != b {
			t.Fatalf(`"another" should exclude the Ardent's own arrival`)
		}
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d -> %d, want -1", i+1, b, got)
		}
	}
}

func TestGemcutterBuccaneerMakesTappedTreasuresForPirates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gem := pushCatalogPermanent(g, me.ID, "Gemcutter Buccaneer", "Creature — Orc Pirate Artificer", gemcutterOracle, false)

	// Its own arrival counts — the trigger doesn't say "another".
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: gem}) })
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Fatalf("the Buccaneer's own ETB should make a Treasure")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && !c.Tapped {
			t.Errorf("Gemcutter's Treasures enter tapped")
		}
	}

	// A non-Pirate entering does not.
	pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := countBattlefieldNamed(g, me.ID, "Treasure")
	bearID := findBattlefieldByName(g, "Bear")
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: bearID}) })
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Treasure") != before {
		t.Errorf("a non-Pirate should not make a Treasure")
	}
}

// Solphim doubles noncombat damage to opponents only — not combat
// damage, and not damage to you.
func TestSolphimDoublesNoncombatDamageToOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Solphim, Mayhem Dominus",
		"Legendary Creature — Phyrexian Horror", solphimOracle, false)
	fireweaver := pushCatalogPermanent(g, me.ID, "Reckless Fireweaver",
		"Creature — Human Artificer", recklessFireweaverOracle, false)
	_ = fireweaver
	before := lifeOfOpponents(g)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	passPriorityAroundTable(t, g)

	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d -> %d, want -2 (1 doubled)", i+1, b, got)
		}
	}
}

func TestSolphimLeavesCombatDamageAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Solphim, Mayhem Dominus",
		"Legendary Creature — Phyrexian Horror", solphimOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := victim.Life

	attackWith(t, g, victim.ID, bear)
	passPriorityAroundTable(t, g)

	if got := victim.Life; got != before-2 {
		t.Errorf("combat damage: %d -> %d, want -2 (undoubled)", before, got)
	}
}

// --- spells ------------------------------------------------------

func TestGambleFetchesThenDiscardsAtRandom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	me.Library.Cards = nil
	prize := uuid.New()
	me.Library.PushTop(game.Card{
		InstanceID: prize, Name: "Prize", TypeLine: "Sorcery",
		Owner: me.ID, Controller: me.ID,
	})

	castCatalogSpell(t, g, "Gamble", "Sorcery", gambleOracle, nil)
	passPriorityAroundTable(t, g)

	// One card fetched, one card discarded — and with an otherwise
	// empty hand it's the same card, which is the joke.
	if me.Hand.Size() != 0 {
		t.Errorf("hand = %d, want 0 (fetched then pitched)", me.Hand.Size())
	}
	if !me.Graveyard.Contains(prize) {
		t.Errorf("the fetched card should be in the graveyard")
	}
}

func TestWindfallRefillsEveryoneToTheLargestHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for _, p := range g.Seats {
		p.Hand.Cards = nil
	}
	for i := 0; i < 3; i++ {
		handCard(opp, "Theirs", "Sorcery")
	}
	handCard(me, "Mine", "Sorcery")

	castCatalogSpell(t, g, "Windfall", "Sorcery", windfallOracle, nil)
	passPriorityAroundTable(t, g)

	// Opponent pitched 3, so everyone draws 3.
	if me.Hand.Size() != 3 {
		t.Errorf("caster hand = %d, want 3", me.Hand.Size())
	}
	if opp.Hand.Size() != 3 {
		t.Errorf("opponent hand = %d, want 3", opp.Hand.Size())
	}
	if me.Graveyard.Size() == 0 || opp.Graveyard.Size() == 0 {
		t.Errorf("both hands should have been discarded first")
	}
}

// The Treasures go to the countered spell's controller, not to the
// player who countered it.
func TestOfferCountersAndPaysTheVictim(t *testing.T) {
	g := newCatalogGame(t)
	// The active seat casts the sorcery (sorcery speed); the seat
	// behind it answers with the instant.
	caster := g.Seats[g.Turn.ActiveSeat]
	interrupter := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	spell := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: spell, Name: "Divination", TypeLine: "Sorcery",
		Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the spell to be countered: %v", err)
	}

	offer := uuid.New()
	interrupter.Hand.PushTop(game.Card{
		InstanceID: offer, Name: "An Offer You Can't Refuse", TypeLine: "Instant",
		OracleID: offerOracle, Owner: interrupter.ID, Controller: interrupter.ID,
	})
	if err := g.CastSpell(interrupter.ID, offer, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err != nil {
		t.Fatalf("cast Offer: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Stack.Contains(spell) {
		t.Errorf("the target spell should have been countered")
	}
	// Leaving the stack is what RESOLVING does too — the graveyard
	// is the only assertion that tells a counter from a resolution.
	if !caster.Graveyard.Contains(spell) {
		t.Errorf("the countered spell did not reach its owner's graveyard")
	}
	if n := countBattlefieldNamed(g, caster.ID, "Treasure"); n != 2 {
		t.Errorf("the countered spell's controller gets 2 Treasures, got %d", n)
	}
	if n := countBattlefieldNamed(g, interrupter.ID, "Treasure"); n != 0 {
		t.Errorf("the player who countered gets none, got %d", n)
	}
}

func TestDecayingTimeLoopSwapsTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	for i := 0; i < 3; i++ {
		handCard(me, "Fodder", "Sorcery")
	}
	before := me.Library.Size()

	castCatalogSpell(t, g, "Decaying Time Loop", "Instant", timeLoopOracle, nil)
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != 3 {
		t.Errorf("hand = %d, want 3", me.Hand.Size())
	}
	if drawn := before - me.Library.Size(); drawn != 3 {
		t.Errorf("drew %d, want 3", drawn)
	}
	// Three pitched cards plus the Time Loop itself, which is put
	// into its owner's graveyard as it finishes resolving.
	if me.Graveyard.Size() != 4 {
		t.Errorf("graveyard = %d, want 3 pitched + the spell", me.Graveyard.Size())
	}
}

func TestPullFromTomorrowDrawsXThenQueuesADiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	active := g.Seats[g.Turn.ActiveSeat]
	before := active.Library.Size()

	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Pull from Tomorrow", TypeLine: "Instant",
		OracleID: pullFromTomorrowOrc, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if drawn := before - active.Library.Size(); drawn != 3 {
		t.Errorf("drew %d, want X=3", drawn)
	}
	if discardOwed(g, active.ID) != 1 {
		t.Errorf("a discard of 1 should be pending, got %d", discardOwed(g, active.ID))
	}
}

func TestFranticSearchLootsAndUntapsThreeLands(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	active.Hand.Cards = nil
	var lands []uuid.UUID
	for i := 0; i < 4; i++ {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: "Island", TypeLine: "Basic Land — Island",
			Tapped: true, Owner: active.ID, Controller: active.ID,
		})
		lands = append(lands, id)
	}

	castCatalogSpell(t, g, "Frantic Search", "Instant", franticSearchOracle, nil)
	passPriorityAroundTable(t, g)

	untapped := 0
	for _, id := range lands {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id && !c.Tapped {
				untapped++
			}
		}
	}
	if untapped != 3 {
		t.Errorf("untapped %d lands, want exactly 3 (up to three)", untapped)
	}
	if discardOwed(g, active.ID) != 2 {
		t.Errorf("loot should queue a 2-card discard, got %d", discardOwed(g, active.ID))
	}
}

// Every card in the batch is registered and reachable by oracle ID.
func TestBatch2CardsAreRegistered(t *testing.T) {
	for _, id := range []string{
		captainStormOracle, imperialRecruiterOrc, magmakinOracle, malcolmNavigatorOrc,
		skyrayOracle, weftstalkerOracle, gemcutterOracle, solphimOracle,
		gambleOracle, windfallOracle, offerOracle, timeLoopOracle,
		pullFromTomorrowOrc, franticSearchOracle,
	} {
		if _, ok := Lookup(id); !ok {
			t.Errorf("%s is not registered", id)
		}
	}
}
