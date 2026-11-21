package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- creatures ---------------------------------------------------

func TestCaptainStormCountersAPirateOnArtifactETB(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Captain Storm, Cosmium Raider",
		"Legendary Creature — Human Pirate", effectstest.CaptainStormOracle, false)
	mako := effectstest.PushCatalogPermanent(g, me.ID, "Marauding Mako", "Creature — Shark Pirate", effectstest.MaraudingMakoOracle, false)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	effectstest.PickCard(t, g, me.ID, mako)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountersOn(g, mako, "+1/+1"); got != 1 {
		t.Errorf("Mako counters = %d, want 1", got)
	}
}

func TestImperialRecruiterFetchesASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
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

	effectstest.CastCatalogSpell(t, g, "Imperial Recruiter", "Creature — Human Advisor", effectstest.ImperialRecruiterOrc, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)

	if !me.Hand.Contains(small) {
		t.Errorf("the power-2 creature should be in hand")
	}
	if me.Hand.Contains(big) {
		t.Errorf("a 5/5 is not a legal fetch")
	}
}

func TestMagmakinArtilleristBurnsOnEveryDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Magmakin Artillerist", "Creature — Elemental Pirate", effectstest.MagmakinOracle, false)
	before := effectstest.LifeOfOpponents(g)
	fodder := effectstest.HandCard(me, "Fodder", "Sorcery")

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	_ = fodder

	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d -> %d, want -1", i+1, b, got)
		}
	}
}

func TestMalcolmMakesATreasureOnPirateDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Malcolm, Keen-Eyed Navigator",
		"Legendary Creature — Siren Pirate", effectstest.MalcolmNavigatorOrc, false)
	pirate := effectstest.PushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)

	effectstest.AttackWith(t, g, victim.ID, pirate)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Errorf("a Pirate connecting should make one Treasure")
	}
}

func TestMalcolmIgnoresNonPirateDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Malcolm, Keen-Eyed Navigator",
		"Legendary Creature — Siren Pirate", effectstest.MalcolmNavigatorOrc, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	effectstest.AttackWith(t, g, victim.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 0 {
		t.Errorf("a non-Pirate should not make a Treasure")
	}
}

// Malcolm prints "deal damage", not "deal combat damage" — a
// Pirate's noncombat damage should still make a Treasure.
func TestMalcolmTriggersOnNoncombatPirateDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Malcolm, Keen-Eyed Navigator",
		"Legendary Creature — Siren Pirate", effectstest.MalcolmNavigatorOrc, false)
	pirate := effectstest.PushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)

	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(pirate, victim.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Errorf("noncombat damage from a Pirate you control should still make a Treasure")
	}
}

// "for each opponent dealt damage": the count is opponents damaged,
// not Pirates that dealt the damage — two Pirates hitting two
// DIFFERENT opponents in one batch make two Treasures, and two
// Pirates hitting the SAME opponent make one, not two.
func TestMalcolmCountsOpponentsDamagedNotPiratesThatDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, v1, v2 := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Malcolm, Keen-Eyed Navigator",
		"Legendary Creature — Siren Pirate", effectstest.MalcolmNavigatorOrc, false)
	p1 := effectstest.PushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)
	p2 := effectstest.PushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)

	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(p1, v1.ID, 1)
		_ = g.DealDamageToPlayerForEffect(p2, v2.ID, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("two Pirates hitting two different opponents should make 2 Treasures, got %d", n)
	}
}

func TestMalcolmCollapsesTwoPiratesHittingTheSameOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Malcolm, Keen-Eyed Navigator",
		"Legendary Creature — Siren Pirate", effectstest.MalcolmNavigatorOrc, false)
	p1 := effectstest.PushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)
	p2 := effectstest.PushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)

	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(p1, victim.ID, 1)
		_ = g.DealDamageToPlayerForEffect(p2, victim.ID, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("two Pirates hitting the same opponent should make 1 Treasure, got %d", n)
	}
}

func TestScroungingSkyrayGrowsOnDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	skyray := effectstest.PushCatalogPermanent(g, me.ID, "Scrounging Skyray", "Creature — Fish Pirate", effectstest.SkyrayOracle, false)
	effectstest.HandCard(me, "Fodder", "Sorcery")

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountersOn(g, skyray, "+1/+1"); got != 1 {
		t.Errorf("Skyray counters = %d, want 1", got)
	}
}

// "Another creature or artifact" — the Ardent's own arrival is not a
// trigger, and a land isn't either.
func TestWeftstalkerArdentPingsOnAnotherPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ardent := effectstest.PushCatalogPermanent(g, me.ID, "Weftstalker Ardent", "Creature — Drix Artificer", effectstest.WeftstalkerOracle, false)
	before := effectstest.LifeOfOpponents(g)

	// Its own ETB does nothing.
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: ardent}) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if g.Seats[i+1].Life != b {
			t.Fatalf(`"another" should exclude the Ardent's own arrival`)
		}
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d -> %d, want -1", i+1, b, got)
		}
	}
}

func TestGemcutterBuccaneerMakesTappedTreasuresForPirates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gem := effectstest.PushCatalogPermanent(g, me.ID, "Gemcutter Buccaneer", "Creature — Orc Pirate Artificer", effectstest.GemcutterOracle, false)

	// Its own arrival counts — the trigger doesn't say "another".
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: gem}) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Fatalf("the Buccaneer's own ETB should make a Treasure")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && !c.Tapped {
			t.Errorf("Gemcutter's Treasures enter tapped")
		}
	}

	// A non-Pirate entering does not.
	effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure")
	bearID := effectstest.FindBattlefieldByName(g, "Bear")
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: bearID}) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != before {
		t.Errorf("a non-Pirate should not make a Treasure")
	}
}

// Solphim doubles noncombat damage to opponents only — not combat
// damage, and not damage to you.
func TestSolphimDoublesNoncombatDamageToOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Solphim, Mayhem Dominus",
		"Legendary Creature — Phyrexian Horror", effectstest.SolphimOracle, false)
	fireweaver := effectstest.PushCatalogPermanent(g, me.ID, "Reckless Fireweaver",
		"Creature — Human Artificer", effectstest.RecklessFireweaverOracle, false)
	_ = fireweaver
	before := effectstest.LifeOfOpponents(g)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)

	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d -> %d, want -2 (1 doubled)", i+1, b, got)
		}
	}
}

func TestSolphimLeavesCombatDamageAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Solphim, Mayhem Dominus",
		"Legendary Creature — Phyrexian Horror", effectstest.SolphimOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := victim.Life

	effectstest.AttackWith(t, g, victim.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if got := victim.Life; got != before-2 {
		t.Errorf("combat damage: %d -> %d, want -2 (undoubled)", before, got)
	}
}

// --- spells ------------------------------------------------------

func TestGambleFetchesThenDiscardsAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	me.Library.Cards = nil
	prize := uuid.New()
	me.Library.PushTop(game.Card{
		InstanceID: prize, Name: "Prize", TypeLine: "Sorcery",
		Owner: me.ID, Controller: me.ID,
	})

	effectstest.CastCatalogSpell(t, g, "Gamble", "Sorcery", effectstest.GambleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for _, p := range g.Seats {
		p.Hand.Cards = nil
	}
	for i := 0; i < 3; i++ {
		effectstest.HandCard(opp, "Theirs", "Sorcery")
	}
	effectstest.HandCard(me, "Mine", "Sorcery")

	effectstest.CastCatalogSpell(t, g, "Windfall", "Sorcery", effectstest.WindfallOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
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
		OracleID: effectstest.OfferOracle, Owner: interrupter.ID, Controller: interrupter.ID,
	})
	if err := g.CastSpell(interrupter.ID, offer, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err != nil {
		t.Fatalf("cast Offer: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Stack.Contains(spell) {
		t.Errorf("the target spell should have been countered")
	}
	// Leaving the stack is what RESOLVING does too — the graveyard
	// is the only assertion that tells a counter from a resolution.
	if !caster.Graveyard.Contains(spell) {
		t.Errorf("the countered spell did not reach its owner's graveyard")
	}
	if n := effectstest.CountBattlefieldNamed(g, caster.ID, "Treasure"); n != 2 {
		t.Errorf("the countered spell's controller gets 2 Treasures, got %d", n)
	}
	if n := effectstest.CountBattlefieldNamed(g, interrupter.ID, "Treasure"); n != 0 {
		t.Errorf("the player who countered gets none, got %d", n)
	}
}

func TestDecayingTimeLoopSwapsTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	for i := 0; i < 3; i++ {
		effectstest.HandCard(me, "Fodder", "Sorcery")
	}
	before := me.Library.Size()

	effectstest.CastCatalogSpell(t, g, "Decaying Time Loop", "Instant", effectstest.TimeLoopOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	active := g.Seats[g.Turn.ActiveSeat]
	before := active.Library.Size()

	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Pull from Tomorrow", TypeLine: "Instant",
		OracleID: effectstest.PullFromTomorrowOrc, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if drawn := before - active.Library.Size(); drawn != 3 {
		t.Errorf("drew %d, want X=3", drawn)
	}
	if effectstest.DiscardOwed(g, active.ID) != 1 {
		t.Errorf("a discard of 1 should be pending, got %d", effectstest.DiscardOwed(g, active.ID))
	}
}

func TestFranticSearchLootsAndUntapsThreeLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
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

	effectstest.CastCatalogSpell(t, g, "Frantic Search", "Instant", effectstest.FranticSearchOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// The loot's discard is asked first, and the untap waits for its
	// answer (CR 608.2c, #4155): only the discard is open now. The untap
	// prompt is told from it by the cards it offers.
	untapPrompt := func() *game.PendingChoice {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == active.ID && effectstest.HasID(c.ChooseCards, lands[0]) {
				return c
			}
		}
		return nil
	}
	if effectstest.DiscardOwed(g, active.ID) != 2 {
		t.Fatalf("loot should queue a 2-card discard, got %d", effectstest.DiscardOwed(g, active.ID))
	}
	if untapPrompt() != nil {
		t.Fatal("the untap prompt is open before the discard is answered")
	}
	effectstest.DiscardFromHand(t, g, active.ID)
	untapPick := untapPrompt()
	if untapPick == nil {
		t.Fatalf("no untap prompt for the controller after the discard: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(untapPick.ID, active.ID, lands[:3]); err != nil {
		t.Fatalf("ResolveChooseCards(untap): %v", err)
	}
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
}

// Every card in the batch is registered and reachable by oracle ID.
func TestBatch2CardsAreRegistered(t *testing.T) {
	for _, id := range []string{
		effectstest.CaptainStormOracle, effectstest.ImperialRecruiterOrc, effectstest.MagmakinOracle, effectstest.MalcolmNavigatorOrc,
		effectstest.SkyrayOracle, effectstest.WeftstalkerOracle, effectstest.GemcutterOracle, effectstest.SolphimOracle,
		effectstest.GambleOracle, effectstest.WindfallOracle, effectstest.OfferOracle, effectstest.TimeLoopOracle,
		effectstest.PullFromTomorrowOrc, effectstest.FranticSearchOracle,
	} {
		if _, ok := effects.Lookup(id); !ok {
			t.Errorf("%s is not registered", id)
		}
	}
}
