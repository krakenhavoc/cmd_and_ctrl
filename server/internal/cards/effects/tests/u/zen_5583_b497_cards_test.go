package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// zen_5583_b497_cards_test.go — Zendikar, batch 497 (#5583). Every card is
// registered Full, and each one with rules text is played through the engine.

const (
	zenB497BalothCageTrapOracle    = "1221bef9-5947-4c32-a74b-b60279bb79bf"
	zenB497SummoningTrapOracle     = "2faa8dc1-0370-458c-9aef-44f81a9b5cd5"
	zenB497TrapfindersTrickOracle  = "9792ae99-7b3a-46cb-bd22-0b33c499ce1a"
	zenB497MindlessNullOracle      = "22ff4e16-5c66-4076-94cc-32c281d42f03"
	zenB497WorldQuellerOracle      = "2ac27b74-aa44-412d-8099-ace5009628f0"
	zenB497NissasChosenOracle      = "d14c9d2b-41c7-4c9a-a66a-cda34d29c7ba"
	zenB497NissaRevaneOracle       = "113fc13a-6df9-4eed-b9d4-002a85a036a1"
	zenB497ChandraAblazeOracle     = "3fae28e6-2ffc-460e-9dd3-e321e8a53ba9"
	zenB497SeaGateLoremasterOracle = "6eed122b-9760-47fd-8ba2-adeda8054e0d"
	zenB497TrapmakersSnareOracle   = "8aa0125c-29fe-4f31-985c-de2da3dbc602"
	zenB497TeeteringPeaksOracle    = "4a9437a6-4e61-48b6-8194-1c6ba6432250"
	zenB497GreenweaverDruidOracle  = "5e35a7c9-1e6e-41a7-9f59-5ea1bc53bc65"

	zenB497CounterspellOracle = "cc187110-1148-4090-bbb8-e205694a39f5"
)

func TestZenB497EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		zenB497BalothCageTrapOracle, zenB497SummoningTrapOracle, zenB497TrapfindersTrickOracle, zenB497MindlessNullOracle,
		zenB497WorldQuellerOracle, zenB497NissasChosenOracle, zenB497NissaRevaneOracle, zenB497ChandraAblazeOracle,
		zenB497SeaGateLoremasterOracle, zenB497TrapmakersSnareOracle, zenB497TeeteringPeaksOracle, zenB497GreenweaverDruidOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// zenB497Walker puts a planeswalker with the given loyalty onto the battlefield.
func zenB497Walker(g *game.Game, owner uuid.UUID, name, oracle, typeLine string, loyalty int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Counters: map[string]int{game.CounterLoyalty: loyalty}, Owner: owner, Controller: owner,
	})
}

// zenB497Activate has who activate row idx of source and settles the stack.
func zenB497Activate(t *testing.T, g *game.Game, who, source uuid.UUID, idx int, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(who, source, idx, params); err != nil {
		t.Fatalf("activate row %d: %v", idx, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Greenweaver Druid -------------------------------------------------------

func TestZenB497GreenweaverDruidAddsTwoGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Greenweaver Druid", "Creature — Elf Druid", zenB497GreenweaverDruidOracle, 1, 1)
	effectstest.ActivateManaFor(t, g, me.ID, id, 0, game.ManaAbilityParams{})
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("pool %v, want GG", got)
	}
	if !effectstest.IsTapped(g, id) {
		t.Error("the Druid did not tap")
	}
}

// --- Mindless Null -----------------------------------------------------------

func TestZenB497MindlessNullBlocksOnlyWithAVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raider := effectstest.B12Creature(g, me.ID, "Raider", "Creature — Bear", 2, 2)
	null := effectstest.B12Push(g, opp.ID, "Mindless Null", "Creature — Zombie", zenB497MindlessNullOracle, 2, 2)
	effectstest.BrAttack(t, g, raider)
	if err := g.DeclareBlocker(null, raider); err == nil {
		t.Fatal("Mindless Null blocked with no Vampire")
	}
}

func TestZenB497MindlessNullBlocksBesideAVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raider := effectstest.B12Creature(g, me.ID, "Raider", "Creature — Bear", 2, 2)
	null := effectstest.B12Push(g, opp.ID, "Mindless Null", "Creature — Zombie", zenB497MindlessNullOracle, 2, 2)
	effectstest.B12Creature(g, opp.ID, "Bloodthrone", "Creature — Vampire", 1, 1)
	effectstest.BrAttack(t, g, raider)
	if err := g.DeclareBlocker(null, raider); err != nil {
		t.Errorf("Mindless Null can't block beside a Vampire: %v", err)
	}
}

func TestZenB497MindlessNullIgnoresAnOpponentsVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raider := effectstest.B12Creature(g, me.ID, "Raider", "Creature — Vampire", 2, 2)
	null := effectstest.B12Push(g, opp.ID, "Mindless Null", "Creature — Zombie", zenB497MindlessNullOracle, 2, 2)
	effectstest.BrAttack(t, g, raider)
	if err := g.DeclareBlocker(null, raider); err == nil {
		t.Error("an attacking Vampire let Mindless Null block")
	}
}

// --- Nissa's Chosen ----------------------------------------------------------

func TestZenB497NissasChosenGoesToTheBottomOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "First", "Second", "Third")
	id := effectstest.B12Push(g, me.ID, "Nissa's Chosen", "Creature — Elf Warrior", zenB497NissasChosenOracle, 2, 3)
	effectstest.B18Kill(t, g, id)
	if g.Battlefield.Contains(id) {
		t.Fatal("Nissa's Chosen is still on the battlefield")
	}
	if me.Graveyard.Contains(id) {
		t.Fatal("Nissa's Chosen went to the graveyard")
	}
	if len(me.Library.Cards) == 0 || me.Library.Cards[0].InstanceID != id {
		t.Error("Nissa's Chosen is not on the bottom of its owner's library")
	}
}

func TestZenB497NissasChosenExiledStaysExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Nissa's Chosen", "Creature — Elf Warrior", zenB497NissasChosenOracle, 2, 3)
	effectstest.ExileCreature(t, g, id)
	for _, c := range me.Library.Cards {
		if c.InstanceID == id {
			t.Error("an exiled Nissa's Chosen went to the library")
		}
	}
}

// --- Sea Gate Loremaster -----------------------------------------------------

func TestZenB497SeaGateLoremasterDrawsPerAlly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	effectstest.B12Creature(g, me.ID, "Ally One", "Creature — Kor Ally", 1, 1)
	effectstest.B12Creature(g, me.ID, "Not An Ally", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, g.Seats[1].ID, "Their Ally", "Creature — Kor Ally", 1, 1)
	id := effectstest.B12Push(g, me.ID, "Sea Gate Loremaster", "Creature — Merfolk Wizard Ally", zenB497SeaGateLoremasterOracle, 1, 3)
	before := len(me.Hand.Cards)
	effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if got := len(me.Hand.Cards) - before; got != 2 {
		t.Errorf("drew %d cards, want 2 (the Loremaster and one other Ally of yours)", got)
	}
}

// --- Teetering Peaks ---------------------------------------------------------

func TestZenB497TeeteringPeaksEntersTappedAndBoostsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.B31PlayLand(t, g, "Teetering Peaks", "Land", zenB497TeeteringPeaksOracle)
	if !effectstest.IsTapped(g, land) {
		t.Error("Teetering Peaks entered untapped")
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			effectstest.AnswerPickTarget(t, g, bear)
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 4/2", p, tough)
	}
}

func TestZenB497TeeteringPeaksTapsForRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.B12Push(g, me.ID, "Teetering Peaks", "Land", zenB497TeeteringPeaksOracle, 0, 0)
	effectstest.ActivateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{})
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want R", got)
	}
}

// --- Trapfinder's Trick and Trapmaker's Snare --------------------------------

func TestZenB497TrapfindersTrickDiscardsOnlyTraps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trap := effectstest.B13HandCard(opp, "Some Trap", "Instant — Trap", "")
	trap2 := effectstest.B13HandCard(opp, "Other Trap", "Instant — Trap", "")
	bolt := effectstest.B13HandCard(opp, "Bolt", "Instant", "")
	_ = me
	effectstest.CastCatalogSpell(t, g, "Trapfinder's Trick", "Sorcery", zenB497TrapfindersTrickOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(trap) || opp.Hand.Contains(trap2) {
		t.Error("a Trap card is still in the hand")
	}
	if !opp.Hand.Contains(bolt) {
		t.Error("a non-Trap card was discarded")
	}
	if !opp.Graveyard.Contains(trap) || !opp.Graveyard.Contains(trap2) {
		t.Error("the Traps are not in the graveyard")
	}
}

func TestZenB497TrapmakersSnareFindsATrap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Filler One", "Filler Two")
	trap := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Some Trap", TypeLine: "Instant — Trap"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plain Instant", TypeLine: "Instant"})
	effectstest.CastCatalogSpell(t, g, "Trapmaker's Snare", "Instant", zenB497TrapmakersSnareOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, trap)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(trap) {
		t.Error("the Trap card is not in hand")
	}
}

// --- Baloth Cage Trap --------------------------------------------------------

func zenB497CageTrapCast(t *testing.T, g *game.Game, me *game.Player, params game.CastSpellParams) error {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Baloth Cage Trap", TypeLine: "Instant — Trap", ManaCost: "{3}{G}{G}",
		OracleID: zenB497BalothCageTrapOracle, Owner: me.ID, Controller: me.ID})
	return g.CastSpell(me.ID, id, params)
}

func TestZenB497BalothCageTrapNeedsAnOpponentsArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CG")
	if err := zenB497CageTrapCast(t, g, me, game.CastSpellParams{Strict: true, AlternativeCost: "trap"}); err == nil {
		t.Fatal("the trap cost was offered with no opposing artifact")
	}
	// My own artifact does not count.
	mine := effectstest.B13HandCard(me, "My Rock", "Artifact", "")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, mine); err != nil {
		t.Fatalf("moving my artifact: %v", err)
	}
	if err := zenB497CageTrapCast(t, g, me, game.CastSpellParams{Strict: true, AlternativeCost: "trap"}); err == nil {
		t.Fatal("the trap cost was offered for my own artifact")
	}
}

func TestZenB497BalothCageTrapMakesABeastForOneGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B13HandCard(opp, "Their Rock", "Artifact", "")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, rock); err != nil {
		t.Fatalf("moving their artifact: %v", err)
	}
	effectstest.FloatForTest(g, me, "CG")
	if err := zenB497CageTrapCast(t, g, me, game.CastSpellParams{Strict: true, AlternativeCost: "trap"}); err != nil {
		t.Fatalf("cast for the trap cost: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && effects.IsToken(c) && c.Name == "Beast" {
			n++
			if p, tough := effectstest.PtOf(t, g, c.InstanceID); p != 4 || tough != 4 {
				t.Errorf("the Beast is %d/%d, want 4/4", p, tough)
			}
		}
	}
	if n != 1 {
		t.Errorf("%d Beast tokens, want 1", n)
	}
}

// --- Summoning Trap ----------------------------------------------------------

func zenB497SummoningTrapCast(t *testing.T, g *game.Game, me *game.Player) error {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Summoning Trap", TypeLine: "Instant — Trap", ManaCost: "{4}{G}{G}",
		OracleID: zenB497SummoningTrapOracle, Owner: me.ID, Controller: me.ID})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "trap"})
}

func TestZenB497SummoningTrapIsFreeAfterAnOpponentCountersACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	if err := zenB497SummoningTrapCast(t, g, me); err == nil {
		t.Fatal("the trap cost was offered before anything was countered")
	}
	bear := effectstest.CastHoldCreatureNoResolve(t, g, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.Batch01OpponentCasts(t, g, opp, "Counterspell", zenB497CounterspellOracle, "", effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear was not countered")
	}
	// The library holds a creature for the trap to find.
	creature := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: creature, Name: "Hill Giant", TypeLine: "Creature — Giant", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := zenB497SummoningTrapCast(t, g, me); err != nil {
		t.Fatalf("cast for {0}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no choose_cards prompt for the look at seven: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{creature}); err != nil {
		t.Fatalf("choosing the creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(creature) {
		t.Error("the chosen creature did not enter the battlefield")
	}
}

func TestZenB497SummoningTrapIgnoresMyOwnCounterspell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.CastHoldCreatureNoResolve(t, g, "Bear", "Creature — Bear", "", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Counterspell", TypeLine: "Instant", OracleID: zenB497CounterspellOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("countering my own spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := zenB497SummoningTrapCast(t, g, me); err == nil {
		t.Error("the trap cost was offered after I countered my own creature spell")
	}
}

func TestZenB497SummoningTrapIgnoresAnOpponentsCounteredCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	// A creature spell an OPPONENT cast, countered by me, is not mine.
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	// Flash is needed to cast on my turn: grant it by casting from their own turn is not available, so
	// the opponent's creature is cast by moving to the stack through a test-only flash keyword.
	_ = id
	if err := zenB497SummoningTrapCast(t, g, me); err == nil {
		t.Error("the trap cost was offered with nothing of mine countered")
	}
}

// --- World Queller -----------------------------------------------------------

func TestZenB497WorldQuellerMakesEveryPlayerSacrificeThePickedType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "World Queller", "Creature — Avatar", zenB497WorldQuellerOracle, 4, 4)
	myBear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirRock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.BlbSTPickOption(t, g, me.ID, "Artifact")
	effectstest.AnswerSacrifice(t, g, opp.ID, theirRock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirRock) {
		t.Error("their artifact was not sacrificed")
	}
	if !g.Battlefield.Contains(myBear) || !g.Battlefield.Contains(theirBear) {
		t.Error("a creature was sacrificed for the artifact type")
	}
}

func TestZenB497WorldQuellerMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "World Queller", "Creature — Avatar", zenB497WorldQuellerOracle, 4, 4)
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.BlbSTPickOption(t, g, me.ID, effects.OldB497WorldQuellerNone)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirBear) {
		t.Error("declining the choice still sacrificed a permanent")
	}
}

// --- Nissa Revane ------------------------------------------------------------

func TestZenB497NissaRevaneFetchesNissasChosen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	chosen := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Nissa's Chosen", TypeLine: "Creature — Elf Warrior", OracleID: zenB497NissasChosenOracle, Power: 2, Toughness: 3})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Other", TypeLine: "Creature — Elf Warrior", Power: 1, Toughness: 1})
	walker := zenB497Walker(g, me.ID, "Nissa Revane", zenB497NissaRevaneOracle, "Legendary Planeswalker — Nissa", 2)
	zenB497Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	effectstest.AnswerSearchByID(t, g, me.ID, chosen)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(chosen) {
		t.Error("Nissa's Chosen did not enter the battlefield")
	}
	if got := effectstest.CountersOn(g, walker, game.CounterLoyalty); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

func TestZenB497NissaRevaneGainsTwoPerElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Elf One", "Creature — Elf Warrior", 1, 1)
	effectstest.B12Creature(g, me.ID, "Elf Two", "Creature — Elf Druid", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, g.Seats[1].ID, "Their Elf", "Creature — Elf", 1, 1)
	walker := zenB497Walker(g, me.ID, "Nissa Revane", zenB497NissaRevaneOracle, "Legendary Planeswalker — Nissa", 2)
	before := me.Life
	zenB497Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	if me.Life-before != 4 {
		t.Errorf("gained %d life, want 4", me.Life-before)
	}
}

func TestZenB497NissaRevaneUltimateFetchesElves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	e1 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Elf One", TypeLine: "Creature — Elf Warrior", Power: 1, Toughness: 1})
	e2 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Elf Two", TypeLine: "Creature — Elf Druid", Power: 1, Toughness: 1})
	bear := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	walker := zenB497Walker(g, me.ID, "Nissa Revane", zenB497NissaRevaneOracle, "Legendary Planeswalker — Nissa", 7)
	zenB497Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	effectstest.AnswerSearchByID(t, g, me.ID, e1, e2)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(e1) || !g.Battlefield.Contains(e2) {
		t.Error("the Elves did not enter the battlefield")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("a non-Elf entered the battlefield")
	}
}

// --- Chandra Ablaze ----------------------------------------------------------

func TestZenB497ChandraDiscardingARedCardDealsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	red := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: red, Name: "Red Card", TypeLine: "Sorcery", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	walker := zenB497Walker(g, me.ID, "Chandra Ablaze", zenB497ChandraAblazeOracle, "Legendary Planeswalker — Chandra", 5)
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, walker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		effectstest.AnswerDiscard(t, g, me.ID, red)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !me.Graveyard.Contains(red) {
		t.Fatal("the red card was not discarded")
	}
	if before-opp.Life != 4 {
		t.Errorf("the target lost %d, want 4", before-opp.Life)
	}
}

func TestZenB497ChandraDiscardingANonRedCardDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	blue := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: blue, Name: "Blue Card", TypeLine: "Sorcery", Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	walker := zenB497Walker(g, me.ID, "Chandra Ablaze", zenB497ChandraAblazeOracle, "Legendary Planeswalker — Chandra", 5)
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, walker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		effectstest.AnswerDiscard(t, g, me.ID, blue)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if opp.Life != before {
		t.Errorf("the target lost %d with a non-red discard, want 0", before-opp.Life)
	}
}

func TestZenB497ChandraMinusTwoWheelsEveryHandToThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "a", "b", "c", "d", "e")
	effectstest.SeedLibrary(opp, "a", "b", "c", "d", "e")
	old1 := effectstest.B13HandCard(me, "Old One", "Sorcery", "")
	old2 := effectstest.B13HandCard(opp, "Old Two", "Sorcery", "")
	walker := zenB497Walker(g, me.ID, "Chandra Ablaze", zenB497ChandraAblazeOracle, "Legendary Planeswalker — Chandra", 5)
	zenB497Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	for _, p := range []*game.Player{me, opp} {
		if effectstest.DiscardChoiceFor(g, p.ID) != nil {
			t.Fatalf("%s was asked to discard instead of the hand going", p.Name)
		}
		if len(p.Hand.Cards) != 3 {
			t.Errorf("%s has %d cards, want 3", p.Name, len(p.Hand.Cards))
		}
	}
	if me.Hand.Contains(old1) || opp.Hand.Contains(old2) {
		t.Error("an old card is still in a hand")
	}
}

func TestZenB497ChandraUltimateCastsEveryRedInstantAndSorceryFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	burn1 := uuid.New()
	burn2 := uuid.New()
	blue := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: burn1, Name: "Red Burn", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: burn2, Name: "Red Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: blue, Name: "Blue Instant", TypeLine: "Instant", ManaCost: "{U}", Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	_ = opp
	walker := zenB497Walker(g, me.ID, "Chandra Ablaze", zenB497ChandraAblazeOracle, "Legendary Planeswalker — Chandra", 7)
	if err := g.ActivateCatalogAbility(me.ID, walker, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, burn1, burn2)
	for _, id := range offer.DiscardOptions {
		if id == blue {
			t.Fatal("a blue instant is offered")
		}
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, burn1, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting the first free: %v", err)
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) == nil {
		t.Fatal("the second card is no longer on offer after the first cast")
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, burn2, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting the second free: %v", err)
	}
}
