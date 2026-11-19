package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b25_cards_test.go — Foundations Jumpstart, batch 25 (#5929, S82).
// Each test plays the real catalog entry.

const (
	bnsB25AkroanCrusaderOracle   = "d981d85c-e97b-4fe7-b415-6e73a3e7e35c"
	bnsB25AkroanSkyguardOracle   = "f4e8be6f-2a1f-47b2-a4d8-7e576e25ab60"
	bnsB25AmbassadorOakOracle    = "4045c24e-098f-4582-8c50-576a94c7b705"
	bnsB25AngelicCubOracle       = "c8c110db-4d98-4087-a3b0-c7fe88b5e38a"
	bnsB25ArchfiendsVesselOracle = "62c240ac-f2dd-4e39-9e82-50307efa833e"
	bnsB25ArchonOracle           = "38e0a1d0-2a4c-4510-9610-abea7fa1de7c"
	bnsB25ArenaAthleteOracle     = "4dfea3bf-dfd1-4abd-a17e-f3a07488a069"
	bnsB25AudaciousThiefOracle   = "5905c55a-f1d1-4021-a9af-f85b02ae21f3"
	bnsB25AugurOfBolasOracle     = "be5d6779-ecc9-44b9-8c4d-97a243059d55"
	bnsB25AvenReedstalkerOracle  = "d9f43942-4296-417d-a180-5b4f64e4b12d"
	bnsB25BalothGorgerOracle     = "7f9bc0c9-0333-4c02-bc14-e6affe848b5d"
)

func TestBnsB25EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB25AkroanCrusaderOracle:   "Akroan Crusader",
		bnsB25AkroanSkyguardOracle:   "Akroan Skyguard",
		bnsB25AmbassadorOakOracle:    "Ambassador Oak",
		bnsB25AngelicCubOracle:       "Angelic Cub",
		bnsB25ArchfiendsVesselOracle: "Archfiend's Vessel",
		bnsB25ArchonOracle:           "Archon of Falling Stars",
		bnsB25ArenaAthleteOracle:     "Arena Athlete",
		bnsB25AudaciousThiefOracle:   "Audacious Thief",
		bnsB25AugurOfBolasOracle:     "Augur of Bolas",
		bnsB25AvenReedstalkerOracle:  "Aven Reedstalker",
		bnsB25BalothGorgerOracle:     "Baloth Gorger",
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

// bnsB25Heroic casts Hearts on Fire at the creature and lets it resolve.
func bnsB25Heroic(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Hearts on Fire", "Instant", j25b1HeartsOnFireOracle, cardRefs(target))
}

// --- Akroan Crusader ---------------------------------------------------------

func TestBnsB25AkroanCrusaderMakesAHastySoldierWhenTargeted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	c := b12Push(g, me.ID, "Akroan Crusader", "Creature — Human Soldier", bnsB25AkroanCrusaderOracle, 1, 1)
	bnsB25Heroic(t, g, c)
	passPriorityAroundTable(t, g)
	if got := battlefieldCountNamed(g, me.ID, "Soldier"); got != 1 {
		t.Fatalf("%d Soldier tokens, want 1", got)
	}
	for _, card := range g.BattlefieldCardsForEffect() {
		if card.Name == "Soldier" {
			if card.Power != 1 || card.Toughness != 1 || !hasEffectiveKeyword(t, g, card.InstanceID, "haste") {
				t.Errorf("token %d/%d, want a 1/1 with haste", card.Power, card.Toughness)
			}
		}
	}
}

func TestBnsB25AkroanCrusaderIgnoresASpellThatDoesNotTargetIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Akroan Crusader", "Creature — Human Soldier", bnsB25AkroanCrusaderOracle, 1, 1)
	other := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bnsB25Heroic(t, g, other)
	passPriorityAroundTable(t, g)
	if got := battlefieldCountNamed(g, me.ID, "Soldier"); got != 0 {
		t.Errorf("%d Soldier tokens for a spell aimed elsewhere, want 0", got)
	}
}

// --- Akroan Skyguard ---------------------------------------------------------

func TestBnsB25AkroanSkyguardGrowsWhenTargeted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	c := b12Push(g, me.ID, "Akroan Skyguard", "Creature — Human Soldier", bnsB25AkroanSkyguardOracle, 1, 1)
	if !hasEffectiveKeyword(t, g, c, "flying") {
		t.Error("no flying")
	}
	bnsB25Heroic(t, g, c)
	passPriorityAroundTable(t, g)
	if got := b18Counters(t, g, c, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

// --- Ambassador Oak ----------------------------------------------------------

func TestBnsB25AmbassadorOakMakesAnElfWarrior(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Ambassador Oak", "Creature — Treefolk Warrior", bnsB25AmbassadorOakOracle, 3, 3)
	passPriorityAroundTable(t, g)
	if got := battlefieldCountNamed(g, me.ID, "Elf Warrior"); got != 1 {
		t.Errorf("%d Elf Warrior tokens, want 1", got)
	}
}

// --- Angelic Cub -------------------------------------------------------------

func TestBnsB25AngelicCubGrowsOnlyOnTheFirstTargetingEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	c := b12Push(g, me.ID, "Angelic Cub", "Creature — Cat Angel", bnsB25AngelicCubOracle, 1, 1)
	bnsB25Heroic(t, g, c)
	passPriorityAroundTable(t, g)
	if got := b18Counters(t, g, c, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after the first targeting, want 1", got)
	}
	bnsB25Heroic(t, g, c)
	passPriorityAroundTable(t, g)
	if got := b18Counters(t, g, c, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after a second targeting the same turn, want 1", got)
	}
}

func TestBnsB25AngelicCubFliesAtThreeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	c := b12Push(g, me.ID, "Angelic Cub", "Creature — Cat Angel", bnsB25AngelicCubOracle, 1, 1)
	setCounters(g, c, map[string]int{game.CounterPlusOne: 2})
	if hasEffectiveKeyword(t, g, c, "flying") {
		t.Fatal("flying with two counters")
	}
	setCounters(g, c, map[string]int{game.CounterPlusOne: 3})
	if !hasEffectiveKeyword(t, g, c, "flying") {
		t.Error("no flying with three counters")
	}
}

// --- Archfiend's Vessel ------------------------------------------------------

func TestBnsB25ArchfiendsVesselIsQuietWhenCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := castHoldCreature(t, g, "Archfiend's Vessel", "Creature — Human Cleric", bnsB25ArchfiendsVesselOracle, 1, 1)
	passPriorityAroundTable(t, g)
	if got := battlefieldCountNamed(g, me.ID, "Demon"); got != 0 {
		t.Errorf("%d Demons from a hand cast, want 0", got)
	}
	if !onBattlefield(g, id) {
		t.Error("the Vessel left the battlefield")
	}
	if !hasEffectiveKeyword(t, g, id, "lifelink") {
		t.Error("no lifelink")
	}
}

func TestBnsB25ArchfiendsVesselBecomesADemonWhenCastFromYourGraveyard(t *testing.T) {
	g, me, _ := muldrothaTable(t)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Archfiend's Vessel", TypeLine: "Creature — Human Cleric", ManaCost: "{0}",
		OracleID: bnsB25ArchfiendsVesselOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	if err := muldrothaCast(g, me, id, ""); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g) // the spell resolves; the trigger goes up
	passPriorityAroundTable(t, g)
	if onBattlefield(g, id) {
		t.Error("the Vessel stayed on the battlefield")
	}
	demons := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Demon" && c.Controller == me.ID {
			demons++
			if c.Power != 5 || c.Toughness != 5 || !hasEffectiveKeyword(t, g, c.InstanceID, "flying") {
				t.Errorf("Demon %d/%d, want 5/5 flying", c.Power, c.Toughness)
			}
		}
	}
	if demons != 1 {
		t.Errorf("%d Demons, want 1", demons)
	}
}

func TestBnsB25ArchfiendsVesselBecomesADemonWhenItEntersFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := pushCatalogGraveyardCard(me, "Archfiend's Vessel", "Creature — Human Cleric", bnsB25ArchfiendsVesselOracle, 1, 1)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, id) {
		t.Error("the Vessel stayed on the battlefield")
	}
	if got := battlefieldCountNamed(g, me.ID, "Demon"); got != 1 {
		t.Errorf("%d Demons, want 1", got)
	}
}

// --- Archon of Falling Stars -------------------------------------------------

func TestBnsB25ArchonReturnsAnEnchantmentWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	archon := b12Push(g, me.ID, "Archon of Falling Stars", "Creature — Archon", bnsB25ArchonOracle, 4, 4)
	if !hasEffectiveKeyword(t, g, archon, "flying") {
		t.Error("no flying")
	}
	ench := pushGraveyardPermanent(me, "Old Enchantment", "Enchantment", "{1}{W}")
	bear := pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	killCreature(t, g, me.ID, archon)
	pickCard(t, g, me.ID, ench)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, ench) {
		t.Error("the enchantment did not return")
	}
	if onBattlefield(g, bear) {
		t.Error("a creature card returned")
	}
}

func TestBnsB25ArchonMayDeclineTheReturn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	archon := b12Push(g, me.ID, "Archon of Falling Stars", "Creature — Archon", bnsB25ArchonOracle, 4, 4)
	ench := pushGraveyardPermanent(me, "Old Enchantment", "Enchantment", "{1}{W}")
	killCreature(t, g, me.ID, archon)
	pickCard(t, g, me.ID, ench)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, ench) {
		t.Error("the enchantment returned on a no")
	}
}

// --- Arena Athlete -----------------------------------------------------------

func TestBnsB25ArenaAthleteStopsAnOpposingBlocker(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := seatsForCombat(g)
	a := b12Push(g, me.ID, "Arena Athlete", "Creature — Human", bnsB25ArenaAthleteOracle, 2, 1)
	blocker := b12Creature(g, them.ID, "Wall", "Creature — Wall", 0, 4)
	other := b12Creature(g, them.ID, "Other Wall", "Creature — Wall", 0, 4)
	bnsB25Heroic(t, g, a)
	mscM2PickTargets(t, g, blocker)
	attackIntoBlocks(t, g, them.ID, a)
	if blockRefusal(t, g, blocker, a) == "" {
		t.Error("the target blocked")
	}
	if got := blockRefusal(t, g, other, a); got != "" {
		t.Errorf("another creature was refused: %q", got)
	}
}

// --- Audacious Thief ---------------------------------------------------------

func TestBnsB25AudaciousThiefDrawsAndLosesLifeOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	thief := b12Push(g, me.ID, "Audacious Thief", "Creature — Human Rogue", bnsB25AudaciousThiefOracle, 2, 2)
	hand, life := me.Hand.Size(), me.Life
	declareAttack(t, g, opp.ID, thief)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, me.Life)
	}
}

// --- Augur of Bolas ----------------------------------------------------------

func TestBnsB25AugurOfBolasTakesAnInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	sorc := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: sorc, Name: "Test Sorcery", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	for _, id := range ids {
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	castHoldCreature(t, g, "Augur of Bolas", "Creature — Merfolk Wizard", bnsB25AugurOfBolasOracle, 1, 3)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no take-one prompt: %+v", g.PendingChoices)
	}
	for _, c := range pick.ChooseCards {
		if c == ids[0] || c == ids[1] {
			t.Errorf("a non-instant, non-sorcery card %v is offered", c)
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{sorc}); err != nil {
		t.Fatalf("take the sorcery: %v", err)
	}
	if !me.Hand.Contains(sorc) {
		t.Error("the sorcery is not in hand")
	}
}

// --- Aven Reedstalker --------------------------------------------------------

func TestBnsB25AvenReedstalkerHasFlashAndFlying(t *testing.T) {
	spec, _ := Lookup(bnsB25AvenReedstalkerOracle)
	if len(spec.PrintedKeywords) != 2 || spec.PrintedKeywords[0] != "flash" || spec.PrintedKeywords[1] != "flying" {
		t.Errorf("keywords %v, want flash and flying", spec.PrintedKeywords)
	}
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	c := b12Push(g, me.ID, "Aven Reedstalker", "Creature — Bird Warrior", bnsB25AvenReedstalkerOracle, 2, 3)
	if !hasEffectiveKeyword(t, g, c, "flying") || !hasEffectiveKeyword(t, g, c, "flash") {
		t.Error("missing flash or flying")
	}
}

// --- Baloth Gorger -----------------------------------------------------------

func TestBnsB25BalothGorgerEntersWithThreeCountersOnlyWhenKicked(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		want     int
	}{
		{"unkicked", nil, 0},
		{"kicked", []int{0}, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id, err := castWithOptionalCosts(t, g, "Baloth Gorger", "Creature — Beast", bnsB25BalothGorgerOracle, nil, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			if n := plusOneCounters(g, id); n != tc.want {
				t.Errorf("+1/+1 counters %d, want %d", n, tc.want)
			}
		})
	}
}
