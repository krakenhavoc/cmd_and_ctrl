package testsru

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b95_test.go — set GAP, batch 95: Doctor Who instants, lands and
// sorceries (S82, #5948).

const (
	gapB95EverybodyLives = "39213de3-6a4a-4879-a7f9-70f45013765e"
	gapB95Council        = "26e4b49e-77e7-41d9-94c5-924669a82591"
	gapB95Cemetery       = "d002391f-1dad-4966-ac36-56cc3ec015b2"
	gapB95Clocktower     = "69143645-97b6-4c7c-9fa2-844fb3b99822"
	gapB95Exterminate    = "90b540ab-b05a-4d29-b896-7ff5da0b101d"
	gapB95FiveDoctors    = "dbdd93d2-f2c6-43f9-93e6-2185c7df83bc"
	gapB95Coward         = "21591447-c045-4317-8d0e-705627aa157b"
	gapB95Delete         = "01d0156f-c26a-4668-a3a4-15ee0790d1b6"
	gapB95Ecstatic       = "724d933c-6e5f-458c-9b39-83b4662693ea"
)

func gapB95TimeCounters(g *game.Game, id uuid.UUID) int {
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.Counters[game.CounterTime]
}

func gapB95Exiled(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

func TestS82GapB95EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB95EverybodyLives: "Everybody Lives!", gapB95Council: "Gallifrey Council Chamber", gapB95Cemetery: "Ominous Cemetery",
		gapB95Clocktower: "Trenzalore Clocktower", gapB95Exterminate: "Exterminate!", gapB95FiveDoctors: "The Five Doctors",
		gapB95Coward: "Coward", gapB95Coward + "#1": "Killer", gapB95Delete: "Delete", gapB95Ecstatic: "Ecstatic Beauty",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%+v", name, ok, spec)
		}
	}
}

// --- Everybody Lives! ---------------------------------------------------------

func TestS82GapB95EverybodyLivesProtectsCreaturesPlayersLifeAndTheGame(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Everybody Lives!", "Instant", gapB95EverybodyLives, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		for _, kw := range []string{"hexproof", "indestructible"} {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("a creature lacks %s", kw)
			}
		}
	}
	g.ReadSnapshot(func() {
		for _, p := range []*game.Player{me, opp} {
			if !g.PlayerHasKeywordLocked(p, game.KeywordHexproof) {
				t.Errorf("a player lacks hexproof")
			}
		}
	})
	// Life can't be lost: a drain, damage and a payment all leave it alone.
	life := opp.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -5) })
	if opp.Life != life {
		t.Errorf("a life loss landed: %d -> %d", life, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 3) })
	if opp.Life != life+3 {
		t.Errorf("a life GAIN was stopped: %d, want %d (only losing is barred)", opp.Life, life+3)
	}
	g.WithWriteLock(func() {
		if g.CanPayLifeLocked(opp, 1) {
			t.Error("a player who can't lose life paid life")
		}
	})
	// Lethal damage doesn't kill an indestructible creature.
	if err := g.MarkCombatDamage(mine, theirs, 5); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	g.RunStateChecksForTest()
	if _, here := effectstest.BattlefieldCard(g, theirs); !here {
		t.Error("an indestructible creature died to damage")
	}
	// A player at 0 life doesn't lose the game this turn.
	g.WithWriteLock(func() { opp.Life = 0 })
	g.RunStateChecksForTest()
	if g.State != game.StateActive || opp.Eliminated {
		t.Errorf("a player at 0 life lost the game this turn (state %v, eliminated %v)", g.State, opp.Eliminated)
	}
	// It ends with the turn.
	g.WithWriteLock(func() { opp.Life = 7 })
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -2) })
	if opp.Life != 5 {
		t.Errorf("after the turn life loss still barred: %d, want 5", opp.Life)
	}
	if effectstest.HasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("indestructible outlasted the turn")
	}
}

// --- Gallifrey Council Chamber ---------------------------------------------------

func TestS82GapB95GallifreyCouncilChamberSurveilsOnEntering(t *testing.T) {
	g, me, _ := effectstest.C2LandTable(t)
	effectstest.PushLibraryTyped(me, "Top Card", "Basic Land — Forest")
	effectstest.B31PlayLand(t, g, "Gallifrey Council Chamber", "Legendary Land", gapB95Council)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Fatal("entering did not ask to surveil")
	}
}

func TestS82GapB95GallifreyCouncilChamberManaPaysOnlyForTimeLordsAndAliens(t *testing.T) {
	g, me, _ := effectstest.C2LandTable(t)
	land := effectstest.C2Land(g, me.ID, "Gallifrey Council Chamber", "Legendary Land", gapB95Council)
	cast := func(name, typeLine string) error {
		if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
			t.Fatalf("tap for any colour: %v", err)
		}
		id := effectstest.HandSpell(me, name, typeLine, "{U}")
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
		g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, land).Tapped = false })
		return err
	}
	for name, typeLine := range map[string]string{"Time Lord Bear": "Creature — Time Lord Doctor", "Alien Bear": "Creature — Alien"} {
		if err := cast(name, typeLine); err != nil {
			t.Errorf("%s with the restricted mana: %v", name, err)
		}
		g, me, _ = effectstest.C2LandTable(t)
		land = effectstest.C2Land(g, me.ID, "Gallifrey Council Chamber", "Legendary Land", gapB95Council)
	}
	if err := cast("Human Bear", "Creature — Human"); err == nil {
		t.Error("the restricted mana paid for a Human spell")
	}
	// The colourless ability is unrestricted.
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {C}: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("the pool has %d mana, want 1", me.ManaPool.Total())
	}
}

// --- Ominous Cemetery -----------------------------------------------------------

func TestS82GapB95OminousCemeteryShufflesATargetIntoItsOwnersLibraryAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Ominous Cemetery", "Land", gapB95Cemetery, false)
	stolen := effectstest.B12Creature(g, me.ID, "Stolen Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, stolen).Owner = opp.ID })
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(stolen), Strict: true}); err == nil {
		t.Fatal("activated without {5}")
	}
	effectstest.C2Pay(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(stolen)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !g.Exile.Contains(land) {
		t.Error("the land was not exiled as a cost")
	}
	if g.Battlefield.Contains(stolen) || !opp.Library.Contains(stolen) {
		t.Error("the creature did not go into its OWNER's library")
	}
}

// --- Trenzalore Clocktower ------------------------------------------------------

func TestS82GapB95TrenzaloreClocktowerCountsUpAndShufflesHandAndGraveyardAwayWithATimeLord(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tower := effectstest.PushCatalogPermanent(g, me.ID, "Trenzalore Clocktower", "Legendary Land", gapB95Clocktower, false)
	if err := g.ActivateManaAbility(me.ID, tower, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {U}: %v", err)
	}
	if got := gapB95TimeCounters(g, tower); got != 1 {
		t.Fatalf("%d time counters after one tap, want 1", got)
	}
	g.WithWriteLock(func() {
		c := effectstest.BattlefieldCardPtr(t, g, tower)
		c.Counters[game.CounterTime] = 12
		c.Tapped = false
	})
	for i := 0; i < 3; i++ {
		effectstest.HandCardForTest(me, "Hand Card", "Sorcery", "")
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Grave Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	for i := 0; i < 12; i++ {
		effectstest.PushLibraryTyped(me, "Library Card", "Basic Land — Forest")
	}
	effectstest.C2Pay(t, g, me, "{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, tower, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without controlling a Time Lord")
	}
	effectstest.B12Creature(g, me.ID, "The Doctor", "Legendary Creature — Time Lord Doctor", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, tower, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tower) || !g.Exile.Contains(tower) {
		t.Error("the Clocktower was not exiled as a cost")
	}
	if got := me.Hand.Size(); got != 7 {
		t.Errorf("hand %d, want 7 (the old hand shuffled away, seven drawn)", got)
	}
	if got := me.Graveyard.Size(); got != 0 {
		t.Errorf("graveyard %d, want 0", got)
	}
}

// --- Exterminate! ----------------------------------------------------------------

func gapB95CastExterminate(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID, replicate int, taps ...uuid.UUID) error {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Exterminate!", TypeLine: "Sorcery", OracleID: gapB95Exterminate, Owner: me.ID, Controller: me.ID})
	var opt []int
	for range replicate {
		opt = append(opt, 0)
	}
	return g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(target), OptionalCosts: opt, TeamworkIDs: taps, Strict: true})
}

func TestS82GapB95ExterminateDestroysAndDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	life := opp.Life
	if err := gapB95CastExterminate(t, g, me, victim, 0); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not destroyed")
	}
	if opp.Life != life-3 {
		t.Errorf("its controller's life %d, want %d", opp.Life, life-3)
	}
}

func TestS82GapB95ExterminateDrainsEvenWhenTheCreatureSurvives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Tough Victim", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		effectstest.BattlefieldCardPtr(t, g, victim).Keywords = append(effectstest.BattlefieldCardPtr(t, g, victim).Keywords, "indestructible")
	})
	life := opp.Life
	if err := gapB95CastExterminate(t, g, me, victim, 0); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d: the controller loses 3 life whether or not it died", opp.Life, life-3)
	}
}

func TestS82GapB95ExterminateReplicateTapsAnUntappedDalekPerCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	dalek := effectstest.B12Creature(g, me.ID, "Dalek", "Artifact Creature — Dalek", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := gapB95CastExterminate(t, g, me, victim, 1, bear); err == nil {
		t.Fatal("replicate paid by tapping a non-Dalek")
	}
	if err := gapB95CastExterminate(t, g, me, victim, 1, dalek); err != nil {
		t.Fatalf("cast with replicate: %v", err)
	}
	if c, _ := effectstest.BattlefieldCard(g, dalek); !c.Tapped {
		t.Error("the Dalek was not tapped to pay replicate")
	}
	if err := gapB95CastExterminate(t, g, me, victim, 1, dalek); err == nil {
		t.Error("a tapped Dalek paid replicate")
	}
}

// --- Delete ------------------------------------------------------------------------

func TestS82GapB95DeleteDamagesNonartifactCreaturesAndEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 2, 5)
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Delete", TypeLine: "Sorcery", OracleID: gapB95Delete, Owner: me.ID, Controller: me.ID})
	mine, theirs := me.Life, opp.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3, Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a 2/2 survived 3 damage")
	}
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(golem) {
		t.Error("the 2/5 or the artifact creature died")
	}
	if bnsMarked(g, golem) != 0 {
		t.Errorf("the artifact creature was dealt %d damage", bnsMarked(g, golem))
	}
	if me.Life != mine-3 || opp.Life != theirs-3 {
		t.Errorf("life %d / %d, want %d / %d (every player takes X)", me.Life, opp.Life, mine-3, theirs-3)
	}
}

// --- The Five Doctors --------------------------------------------------------------

func gapB95FiveDoctorsTable(t *testing.T) (*game.Game, *game.Player, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	var doctors []uuid.UUID
	for i := 0; i < 4; i++ {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Library Doctor", TypeLine: "Legendary Creature — Time Lord Doctor", Owner: me.ID, Controller: me.ID})
		doctors = append(doctors, id)
	}
	for i := 0; i < 3; i++ {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Graveyard Doctor", TypeLine: "Legendary Creature — Time Lord Doctor", Owner: me.ID, Controller: me.ID})
		doctors = append(doctors, id)
	}
	effectstest.PushLibraryTyped(me, "Not A Doctor", "Basic Land — Forest")
	return g, me, doctors
}

func TestS82GapB95FiveDoctorsFindsUpToFiveDoctorsFromLibraryAndGraveyardIntoHand(t *testing.T) {
	g, me, doctors := gapB95FiveDoctorsTable(t)
	effectstest.CastCatalogSpell(t, g, "The Five Doctors", "Sorcery", gapB95FiveDoctors, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	got := 0
	for _, id := range doctors {
		if me.Hand.Contains(id) {
			got++
		}
	}
	if got != 5 {
		t.Errorf("%d Doctor cards in hand, want 5 (up to five of the seven)", got)
	}
}

func TestS82GapB95FiveDoctorsKickedPutsThemOntoTheBattlefield(t *testing.T) {
	g, me, _ := gapB95FiveDoctorsTable(t)
	if _, err := effectstest.CastWithOptionalCosts(t, g, "The Five Doctors", "Sorcery", gapB95FiveDoctors, nil, []int{0}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Library Doctor", me.ID) + effectstest.CountOnBattlefield(g, "Graveyard Doctor", me.ID); got != 5 {
		t.Errorf("%d Doctors on the battlefield, want 5", got)
	}
}

// --- Coward // Killer ------------------------------------------------------------------

func TestS82GapB95CowardStopsBlockingAddsATypeAndTimeTravels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	clock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Clock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterTime: 1}})
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Coward", "Sorcery", gapB95Coward, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, victim); !slices.Contains(c.Effective().Subtypes, "Coward") {
		t.Error("the creature did not become a Coward")
	}
	if got := gapB95TimeCounters(g, clock); got != 2 {
		t.Errorf("the Clock has %d time counters, want 2 (time travel)", got)
	}
	if r := effectstest.BattlefieldCardPtr(t, g, victim).Effective().Restrictions; r&game.CantBlock == 0 {
		t.Error("the creature can still block")
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	if c, _ := effectstest.BattlefieldCard(g, victim); slices.Contains(c.Effective().Subtypes, "Coward") {
		t.Error("the Coward type outlasted the turn")
	}
}

func TestS82GapB95KillerHitsTheTargetAndEveryOtherCreatureSharingAType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	target := effectstest.B12Creature(g, opp.ID, "Elf Warrior", "Creature — Elf Warrior", 2, 5)
	sharesWarrior := effectstest.B12Creature(g, me.ID, "Human Warrior", "Creature — Human Warrior", 2, 5)
	sharesNothing := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 5)
	noType := effectstest.B12Creature(g, me.ID, "Typeless", "Creature", 2, 5)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Killer", TypeLine: "Sorcery", OracleID: gapB95Coward + "#1", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(target), Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, want := range map[uuid.UUID]int{target: 3, sharesWarrior: 3, sharesNothing: 0, noType: 0} {
		if got := bnsMarked(g, name); got != want {
			t.Errorf("a creature was dealt %d damage, want %d", got, want)
		}
	}
}

// --- Ecstatic Beauty ---------------------------------------------------------------

func TestS82GapB95EcstaticBeautyExilesThreeMayPlayThemAndSuspendsTheOnesWithSuspend(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	suspendOracle := "9728dec9-d482-4c7a-8cdc-44d010dc878d" // Ancestral Vision, suspend 4
	plain := effectstest.PushLibraryTyped(me, "Plain Card", "Sorcery")
	vision := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: vision, Name: "Ancestral Vision", TypeLine: "Sorcery", OracleID: suspendOracle, Owner: me.ID, Controller: me.ID})
	land := effectstest.PushLibraryTyped(me, "Exiled Forest", "Basic Land — Forest")
	below := effectstest.PushLibraryTyped(me, "Fourth Card", "Basic Land — Island")
	_ = below
	// Library top first: land, vision, plain, then the card below... ordering by push:
	// the last pushed is the top, so the fourth card is the top; reorder.
	me.Library.Cards = nil
	for _, c := range []game.Card{
		{InstanceID: below, Name: "Fourth Card", TypeLine: "Basic Land — Island"},
		{InstanceID: plain, Name: "Plain Card", TypeLine: "Sorcery"},
		{InstanceID: vision, Name: "Ancestral Vision", TypeLine: "Sorcery", OracleID: suspendOracle},
		{InstanceID: land, Name: "Exiled Forest", TypeLine: "Basic Land — Forest"},
	} {
		c.Owner, c.Controller = me.ID, me.ID
		me.Library.PushTop(c)
	}
	effectstest.CastCatalogSpell(t, g, "Ecstatic Beauty", "Sorcery", gapB95Ecstatic, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"land": land, "vision": vision, "plain": plain} {
		if _, ok := gapB95Exiled(g, id); !ok {
			t.Errorf("the %s was not exiled", name)
		}
	}
	if me.Library.Size() != 1 || !me.Library.Contains(below) {
		t.Error("only the top three cards were exiled")
	}
	if c, _ := gapB95Exiled(g, vision); c.Counters[game.CounterTime] != 4 {
		t.Errorf("the card with suspend has %d time counters, want 4", c.Counters[game.CounterTime])
	}
	for _, id := range []uuid.UUID{land, plain} {
		if c, _ := gapB95Exiled(g, id); c.Counters[game.CounterTime] != 0 {
			t.Errorf("a card without suspend got %d time counters", c.Counters[game.CounterTime])
		}
	}
	granted := 0
	for _, perm := range me.CastPermissions {
		if perm.Zone == game.ZoneExile && perm.Label == "Ecstatic Beauty — you may play the exiled cards this turn" {
			granted++
		}
	}
	if granted != 1 {
		t.Errorf("%d play permissions over the exiled cards, want 1", granted)
	}
}
