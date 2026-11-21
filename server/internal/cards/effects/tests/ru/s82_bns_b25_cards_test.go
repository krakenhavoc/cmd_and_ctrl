package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Akroan Crusader ---------------------------------------------------------

func TestBnsB25AkroanCrusaderMakesAHastySoldierWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	c := effectstest.B12Push(g, me.ID, "Akroan Crusader", "Creature — Human Soldier", bnsB25AkroanCrusaderOracle, 1, 1)
	effectstest.BnsB25Heroic(t, g, c)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCountNamed(g, me.ID, "Soldier"); got != 1 {
		t.Fatalf("%d Soldier tokens, want 1", got)
	}
	for _, card := range g.BattlefieldCardsForEffect() {
		if card.Name == "Soldier" {
			if card.Power != 1 || card.Toughness != 1 || !effectstest.HasEffectiveKeyword(t, g, card.InstanceID, "haste") {
				t.Errorf("token %d/%d, want a 1/1 with haste", card.Power, card.Toughness)
			}
		}
	}
}

func TestBnsB25AkroanCrusaderIgnoresASpellThatDoesNotTargetIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Akroan Crusader", "Creature — Human Soldier", bnsB25AkroanCrusaderOracle, 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.BnsB25Heroic(t, g, other)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCountNamed(g, me.ID, "Soldier"); got != 0 {
		t.Errorf("%d Soldier tokens for a spell aimed elsewhere, want 0", got)
	}
}

// --- Akroan Skyguard ---------------------------------------------------------

func TestBnsB25AkroanSkyguardGrowsWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	c := effectstest.B12Push(g, me.ID, "Akroan Skyguard", "Creature — Human Soldier", bnsB25AkroanSkyguardOracle, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, c, "flying") {
		t.Error("no flying")
	}
	effectstest.BnsB25Heroic(t, g, c)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B18Counters(t, g, c, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

// --- Ambassador Oak ----------------------------------------------------------

func TestBnsB25AmbassadorOakMakesAnElfWarrior(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Ambassador Oak", "Creature — Treefolk Warrior", bnsB25AmbassadorOakOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCountNamed(g, me.ID, "Elf Warrior"); got != 1 {
		t.Errorf("%d Elf Warrior tokens, want 1", got)
	}
}

// --- Angelic Cub -------------------------------------------------------------

func TestBnsB25AngelicCubGrowsOnlyOnTheFirstTargetingEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	c := effectstest.B12Push(g, me.ID, "Angelic Cub", "Creature — Cat Angel", bnsB25AngelicCubOracle, 1, 1)
	effectstest.BnsB25Heroic(t, g, c)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B18Counters(t, g, c, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after the first targeting, want 1", got)
	}
	effectstest.BnsB25Heroic(t, g, c)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B18Counters(t, g, c, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after a second targeting the same turn, want 1", got)
	}
}

func TestBnsB25AngelicCubFliesAtThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	c := effectstest.B12Push(g, me.ID, "Angelic Cub", "Creature — Cat Angel", bnsB25AngelicCubOracle, 1, 1)
	effectstest.SetCounters(g, c, map[string]int{game.CounterPlusOne: 2})
	if effectstest.HasEffectiveKeyword(t, g, c, "flying") {
		t.Fatal("flying with two counters")
	}
	effectstest.SetCounters(g, c, map[string]int{game.CounterPlusOne: 3})
	if !effectstest.HasEffectiveKeyword(t, g, c, "flying") {
		t.Error("no flying with three counters")
	}
}

// --- Archfiend's Vessel ------------------------------------------------------

func TestBnsB25ArchfiendsVesselIsQuietWhenCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.CastHoldCreature(t, g, "Archfiend's Vessel", "Creature — Human Cleric", bnsB25ArchfiendsVesselOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCountNamed(g, me.ID, "Demon"); got != 0 {
		t.Errorf("%d Demons from a hand cast, want 0", got)
	}
	if !effects.OnBattlefield(g, id) {
		t.Error("the Vessel left the battlefield")
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "lifelink") {
		t.Error("no lifelink")
	}
}

func TestBnsB25ArchfiendsVesselBecomesADemonWhenCastFromYourGraveyard(t *testing.T) {
	g, me, _ := effectstest.MuldrothaTable(t)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Archfiend's Vessel", TypeLine: "Creature — Human Cleric", ManaCost: "{0}",
		OracleID: bnsB25ArchfiendsVesselOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	if err := effectstest.MuldrothaCast(g, me, id, ""); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g) // the spell resolves; the trigger goes up
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, id) {
		t.Error("the Vessel stayed on the battlefield")
	}
	demons := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Demon" && c.Controller == me.ID {
			demons++
			if c.Power != 5 || c.Toughness != 5 || !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") {
				t.Errorf("Demon %d/%d, want 5/5 flying", c.Power, c.Toughness)
			}
		}
	}
	if demons != 1 {
		t.Errorf("%d Demons, want 1", demons)
	}
}

func TestBnsB25ArchfiendsVesselBecomesADemonWhenItEntersFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogGraveyardCard(me, "Archfiend's Vessel", "Creature — Human Cleric", bnsB25ArchfiendsVesselOracle, 1, 1)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, id) {
		t.Error("the Vessel stayed on the battlefield")
	}
	if got := effectstest.BattlefieldCountNamed(g, me.ID, "Demon"); got != 1 {
		t.Errorf("%d Demons, want 1", got)
	}
}

// --- Archon of Falling Stars -------------------------------------------------

func TestBnsB25ArchonReturnsAnEnchantmentWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	archon := effectstest.B12Push(g, me.ID, "Archon of Falling Stars", "Creature — Archon", bnsB25ArchonOracle, 4, 4)
	if !effectstest.HasEffectiveKeyword(t, g, archon, "flying") {
		t.Error("no flying")
	}
	ench := effectstest.PushGraveyardPermanent(me, "Old Enchantment", "Enchantment", "{1}{W}")
	bear := effectstest.PushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	effectstest.KillCreature(t, g, me.ID, archon)
	effectstest.PickCard(t, g, me.ID, ench)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, ench) {
		t.Error("the enchantment did not return")
	}
	if effects.OnBattlefield(g, bear) {
		t.Error("a creature card returned")
	}
}

func TestBnsB25ArchonMayDeclineTheReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	archon := effectstest.B12Push(g, me.ID, "Archon of Falling Stars", "Creature — Archon", bnsB25ArchonOracle, 4, 4)
	ench := effectstest.PushGraveyardPermanent(me, "Old Enchantment", "Enchantment", "{1}{W}")
	effectstest.KillCreature(t, g, me.ID, archon)
	effectstest.PickCard(t, g, me.ID, ench)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, ench) {
		t.Error("the enchantment returned on a no")
	}
}

// --- Arena Athlete -----------------------------------------------------------

func TestBnsB25ArenaAthleteStopsAnOpposingBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	a := effectstest.B12Push(g, me.ID, "Arena Athlete", "Creature — Human", bnsB25ArenaAthleteOracle, 2, 1)
	blocker := effectstest.B12Creature(g, them.ID, "Wall", "Creature — Wall", 0, 4)
	other := effectstest.B12Creature(g, them.ID, "Other Wall", "Creature — Wall", 0, 4)
	effectstest.BnsB25Heroic(t, g, a)
	effectstest.MscM2PickTargets(t, g, blocker)
	effectstest.AttackIntoBlocks(t, g, them.ID, a)
	if effectstest.BlockRefusal(t, g, blocker, a) == "" {
		t.Error("the target blocked")
	}
	if got := effectstest.BlockRefusal(t, g, other, a); got != "" {
		t.Errorf("another creature was refused: %q", got)
	}
}

// --- Audacious Thief ---------------------------------------------------------

func TestBnsB25AudaciousThiefDrawsAndLosesLifeOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	thief := effectstest.B12Push(g, me.ID, "Audacious Thief", "Creature — Human Rogue", bnsB25AudaciousThiefOracle, 2, 2)
	hand, life := me.Hand.Size(), me.Life
	effectstest.DeclareAttack(t, g, opp.ID, thief)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, me.Life)
	}
}

// --- Augur of Bolas ----------------------------------------------------------

func TestBnsB25AugurOfBolasTakesAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	sorc := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: sorc, Name: "Test Sorcery", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	for _, id := range ids {
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	effectstest.CastHoldCreature(t, g, "Augur of Bolas", "Creature — Merfolk Wizard", bnsB25AugurOfBolasOracle, 1, 3)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
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
	spec, _ := effects.Lookup(bnsB25AvenReedstalkerOracle)
	if len(spec.PrintedKeywords) != 2 || spec.PrintedKeywords[0] != "flash" || spec.PrintedKeywords[1] != "flying" {
		t.Errorf("keywords %v, want flash and flying", spec.PrintedKeywords)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	c := effectstest.B12Push(g, me.ID, "Aven Reedstalker", "Creature — Bird Warrior", bnsB25AvenReedstalkerOracle, 2, 3)
	if !effectstest.HasEffectiveKeyword(t, g, c, "flying") || !effectstest.HasEffectiveKeyword(t, g, c, "flash") {
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
			g := effectstest.NewCatalogGame(t)
			id, err := effectstest.CastWithOptionalCosts(t, g, "Baloth Gorger", "Creature — Beast", bnsB25BalothGorgerOracle, nil, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.PlusOneCounters(g, id); n != tc.want {
				t.Errorf("+1/+1 counters %d, want %d", n, tc.want)
			}
		})
	}
}
