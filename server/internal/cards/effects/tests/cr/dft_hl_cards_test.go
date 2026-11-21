package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dft_hl_cards_test.go — Aetherdrift (DFT), cards from H to L (#4464).
// Each test plays the real catalog entry and asserts what its text says.
// The fixtures are dft_ab_cards_test.go's, dft_st_cards_test.go's,
// dft_mp_cards_test.go's and dft_cg_cards_test.go's.

const (
	dftHLHellrideOracle    = "47944bf1-b954-4143-b479-45c10071e35a"
	dftHLHauntOracle       = "7e074b14-0ec8-439b-9b83-5f7c1a904384"
	dftHLHazardOracle      = "799664fc-279c-40c1-8e40-44a68c38801b"
	dftHLHowlersOracle     = "8abeaa77-d455-47e7-bf60-aefb94ab83fb"
	dftHLHulldrifterOracle = "a0a7319d-9441-4aed-a351-9ff3fb3b1d9e"
	dftHLTacticsOracle     = "121bd1af-da21-4937-a911-377c5b430796"
	dftHLJibbirikOracle    = "9cda6cb8-c254-413a-babd-0e24e7b89cc4"
	dftHLKalakscionOracle  = "d85f5edb-60a8-4abd-82b0-f9a58f3d4185"
	dftHLBuccaneerOracle   = "8cadf9c0-9a63-4cde-a6be-0ae86890c831"
	dftHLParryOracle       = "ecef9e3d-dce8-405f-bba3-14732444258f"
	dftHLLotusguardOracle  = "1139ee31-cb6e-4ae6-a08b-252c46b2db97"
)

func TestDftHLEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dftHLHellrideOracle:    "Haunted Hellride",
		dftHLHauntOracle:       "Haunt the Network",
		dftHLHazardOracle:      "Hazard of the Dunes",
		dftHLHowlersOracle:     "Howler's Heavy",
		dftHLHulldrifterOracle: "Hulldrifter",
		dftHLTacticsOracle:     "Intimidation Tactics",
		dftHLJibbirikOracle:    "Jibbirik Omnivore",
		dftHLKalakscionOracle:  "Kalakscion, Hunger Tyrant",
		dftHLBuccaneerOracle:   "Keen Buccaneer",
		dftHLParryOracle:       "Lightshield Parry",
		dftHLLotusguardOracle:  "Lotusguard Disciple",
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

// --- fixtures ------------------------------------------------------------------

// dftHLTokensNamed counts the permanents named name that p controls.
func dftHLTokensNamed(g *game.Game, p uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == p && c.Name == name {
			n++
		}
	}
	return n
}

// --- Haunted Hellride -------------------------------------------------------------

func TestDftHLHauntedHellridePumpsGrantsDeathtouchAndUntapsWhenYouAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hellride := effectstest.DftABPush(g, me.ID, "Haunted Hellride", effectstest.DftABVehicle, dftHLHellrideOracle, 3, 3)
	attacker := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// The uncrewed Hellride does not attack; "whenever you attack" fires
	// anyway.
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt: whenever you attack")
	}
	if effectstest.HasID(pick.PickTargetCards, hellride) {
		t.Error("the uncrewed Hellride is not a creature, so not a legal target")
	}
	effectstest.DftCGPickRefused(t, g, me.ID, theirs)
	effectstest.PickCard(t, g, me.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, attacker)
	if !ok {
		t.Fatal("the attacker left the battlefield")
	}
	if c.Tapped {
		t.Error("the attacker was not untapped")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, attacker); p != 3 || tough != 2 {
		t.Errorf("the attacker is %d/%d, want 3/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, attacker, "deathtouch") {
		t.Error("the attacker has no deathtouch")
	}
	effectstest.DftMPPassTurn(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, attacker); p != 2 || tough != 2 {
		t.Errorf("the attacker is %d/%d after the turn ended, want 2/2", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, attacker, "deathtouch") {
		t.Error("deathtouch outlived the turn")
	}
}

func TestDftHLHauntedHellrideCrewsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hellride := effectstest.DftABPush(g, me.ID, "Haunted Hellride", effectstest.DftABVehicle, dftHLHellrideOracle, 3, 3)
	one := effectstest.B12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.DftABCrew(t, g, me, hellride, 0, one)
	if p, tough := effectstest.EffectivePTOf(t, g, hellride); p != 3 || tough != 3 {
		t.Errorf("the crewed Hellride is %d/%d, want 3/3", p, tough)
	}
}

// --- Haunt the Network ------------------------------------------------------------

func TestDftHLHauntTheNetworkMakesTwoThoptersThenDrainsByArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Permanent(g, me.ID, "Mind Stone", "Artifact")
	effectstest.B12Permanent(g, opp.ID, "Their Stone", "Artifact")
	effectstest.AdvanceToMain(t, g)
	mine, theirs := me.Life, opp.Life
	effectstest.CastAt(t, g, "Haunt the Network", "Sorcery", dftHLHauntOracle, effectstest.PbPlayer(opp.ID))
	if got := dftHLTokensNamed(g, me.ID, "Thopter"); got != 2 {
		t.Fatalf("%d Thopters, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Thopter" && !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") {
			t.Error("a Thopter has no flying")
		}
	}
	// X counts the two new Thopters and the Mind Stone, not the
	// opponent's artifact.
	if opp.Life != theirs-3 || me.Life != mine+3 {
		t.Errorf("life: opponent %d -> %d, you %d -> %d; want -3 and +3", theirs, opp.Life, mine, me.Life)
	}
}

func TestDftHLHauntTheNetworkTargetsAnOpponentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := effectstest.CastCatalogSpellErr(t, g, "Haunt the Network", "Sorcery", dftHLHauntOracle,
		[]game.TargetRef{effectstest.PbPlayer(me.ID)}); err == nil {
		t.Fatal("Haunt the Network targeted its caster; it wants an opponent")
	}
}

// --- Hazard of the Dunes ----------------------------------------------------------

func TestDftHLHazardOfTheDunesExhaustsForThreeCountersOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wurm := effectstest.DftABPush(g, me.ID, "Hazard of the Dunes", "Creature — Wurm", dftHLHazardOracle, 4, 4)
	for _, kw := range []string{"reach", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, wurm, kw) {
			t.Errorf("Hazard of the Dunes has no %s", kw)
		}
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 7)
	if err := g.ActivateCatalogAbility(me.ID, wurm, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("exhausted for colourless mana; it costs {6}{G}")
	}
	me.ManaPool = nil
	effectstest.FillPool(me, 6)
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, wurm, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, wurm, game.CounterPlusOne); got != 3 {
		t.Errorf("%d +1/+1 counters, want 3", got)
	}
	if p, tough := effectstest.PtOf(t, g, wurm); p != 7 || tough != 7 {
		t.Errorf("the Wurm is %d/%d, want 7/7", p, tough)
	}
	effectstest.FillPool(me, 6)
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, wurm, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the exhaust ability was activated twice")
	}
}

// --- Howler's Heavy ---------------------------------------------------------------

func TestDftHLHowlersHeavyCyclesAndShrinksAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	effectstest.SeedLibrary(me, "Drawn")
	hand := me.Hand.Size()
	id, _ := effectstest.CycleFromHand(t, g, "Howler's Heavy", "Creature — Seal Pirate", dftHLHowlersOracle, "{C}{U}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Heavy is not in the graveyard")
	}
	effectstest.DftCGPickRefused(t, g, me.ID, mine)
	effectstest.PickTriggerTarget(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d after cycling, want %d (the Heavy out, a card in)", got, hand+1)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 1 || tough != 4 {
		t.Errorf("the target is %d/%d, want 1/4", p, tough)
	}
	effectstest.DftMPPassTurn(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d after the turn ended, want 4/4", p, tough)
	}
}

func TestDftHLHowlersHeavyCyclingWantsBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Howler's Heavy", "Creature — Seal Pirate", dftHLHowlersOracle)
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("cycled for colourless mana; cycling costs {1}{U}")
	}
}

// --- Hulldrifter ------------------------------------------------------------------

func TestDftHLHulldrifterDrawsTwoOnEntryAndCrewsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	ship := effectstest.DftABCast(t, g, me, game.Card{Name: "Hulldrifter", TypeLine: effectstest.DftABVehicle,
		OracleID: dftHLHulldrifterOracle, Power: 3, Toughness: 2})
	if !g.Battlefield.Contains(ship) {
		t.Fatal("Hulldrifter did not enter")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want two cards drawn", hand, got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, ship, "flying") {
		t.Error("Hulldrifter has no flying")
	}
	two := effectstest.B12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	one := effectstest.B12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{two}}); err == nil {
		t.Fatal("crewed with total power 2; crew 3 wants 3")
	}
	effectstest.DftABCrew(t, g, me, ship, 0, two, one)
	if p, tough := effectstest.EffectivePTOf(t, g, ship); p != 3 || tough != 2 {
		t.Errorf("the crewed Hulldrifter is %d/%d, want 3/2", p, tough)
	}
}

// --- Intimidation Tactics ---------------------------------------------------------

func TestDftHLIntimidationTacticsExilesAnArtifactOrCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(),
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"})
	effectstest.CastAt(t, g, "Intimidation Tactics", "Sorcery", dftHLTacticsOracle, effectstest.PbPlayer(victim.ID))
	pick := effectstest.OpenVariant(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, ids[2:]) || pick.PickDestination != game.PickExile || pick.PickOptional {
		t.Fatalf("pick = options %v destination %q optional %v; want the creature and the artifact, exile, mandatory",
			pick.DiscardOptions, pick.PickDestination, pick.PickOptional)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, ids[3:]); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !g.Exile.Contains(ids[3]) || victim.Graveyard.Contains(ids[3]) {
		t.Error("the artifact card was not exiled")
	}
	if victim.Hand.Size() != 3 {
		t.Errorf("victim's hand is %d, want 3", victim.Hand.Size())
	}
}

func TestDftHLIntimidationTacticsTargetsAnOpponentAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := effectstest.CastCatalogSpellErr(t, g, "Intimidation Tactics", "Sorcery", dftHLTacticsOracle,
		[]game.TargetRef{effectstest.PbPlayer(me.ID)}); err == nil {
		t.Fatal("Intimidation Tactics targeted its caster; it wants an opponent")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.SeedLibrary(me2, "Drawn")
	hand := me2.Hand.Size()
	id, _ := effectstest.CycleFromHand(t, g2, "Intimidation Tactics", "Sorcery", dftHLTacticsOracle, "{C}{C}{C}")
	effectstest.PassPriorityAroundTable(t, g2)
	if !me2.Graveyard.Contains(id) || me2.Hand.Size() != hand+1 {
		t.Errorf("cycling did not discard the card for a card: graveyard %v, hand %d want %d",
			me2.Graveyard.Contains(id), me2.Hand.Size(), hand+1)
	}
}

// --- the vanillas -----------------------------------------------------------------

func TestDftHLTheVanillasResolveAsPlainCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Jibbirik Omnivore", "Creature — Beast", dftHLJibbirikOracle, 3, 2},
		{"Kalakscion, Hunger Tyrant", "Legendary Creature — Crocodile", dftHLKalakscionOracle, 7, 2},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		id := effectstest.DftABCast(t, g, me, game.Card{Name: tc.name, TypeLine: tc.typeLine, OracleID: tc.oracle,
			Power: tc.power, Toughness: tc.toughness})
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s did not enter", tc.name)
			continue
		}
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != tc.power || tough != tc.toughness {
			t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.toughness)
		}
		if len(g.PendingChoices) != 0 || g.Stack.Size() != 0 {
			t.Errorf("%s left something behind: choices %v, stack %d", tc.name, g.PendingChoices, g.Stack.Size())
		}
	}
}

// --- Keen Buccaneer ---------------------------------------------------------------

func TestDftHLKeenBuccaneerLootsThenGrowsOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pirate := effectstest.DftABPush(g, me.ID, "Keen Buccaneer", "Creature — Octopus Pirate", dftHLBuccaneerOracle, 2, 3)
	if !effectstest.HasEffectiveKeyword(t, g, pirate, "vigilance") {
		t.Error("Keen Buccaneer has no vigilance")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "Drawn", "Next")
	junk := effectstest.PushCatalogHandCard(me, "Forest", "Basic Land — Forest", "")
	hand := me.Hand.Size()
	effectstest.FillPool(me, 1)
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, pirate, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt after the draw")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d before the discard, want one card drawn", hand, got)
	}
	// The counter waits for the discard (CR 608.2c).
	if got := effectstest.CountersOn(g, pirate, game.CounterPlusOne); got != 0 {
		t.Errorf("%d +1/+1 counters before the discard, want 0", got)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{junk}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(junk) || me.Hand.Size() != hand {
		t.Errorf("the loot did not discard the Forest: graveyard %v, hand %d want %d",
			me.Graveyard.Contains(junk), me.Hand.Size(), hand)
	}
	if got := effectstest.CountersOn(g, pirate, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	effectstest.FillPool(me, 1)
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, pirate, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the exhaust ability was activated twice")
	}
}

// --- Lightshield Parry ------------------------------------------------------------

func TestDftHLLightshieldParryPumpsTwoTwoUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastAt(t, g, "Lightshield Parry", "Instant", dftHLParryOracle, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	effectstest.DftMPPassTurn(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the Bear is %d/%d after the turn ended, want 2/2", p, tough)
	}
}

func TestDftHLLightshieldParryCycles(t *testing.T) {
	effectstest.DftSTCycles(t, "Lightshield Parry", "Instant", dftHLParryOracle, 0)
}

// --- Lotusguard Disciple ----------------------------------------------------------

func TestDftHLLotusguardDiscipleGrantsLifelinkAndIndestructibleToACreatureOrVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	stone := effectstest.B12Permanent(g, me.ID, "Mind Stone", "Artifact")
	wagon := effectstest.B12Permanent(g, opp.ID, "Their Wagon", effectstest.DftABVehicle)
	effectstest.AdvanceToMain(t, g)
	disciple := effectstest.DftABCast(t, g, me, game.Card{Name: "Lotusguard Disciple", TypeLine: "Creature — Bird Cleric",
		OracleID: dftHLLotusguardOracle, Power: 2, Toughness: 2})
	if !effectstest.HasEffectiveKeyword(t, g, disciple, "flying") {
		t.Error("Lotusguard Disciple has no flying")
	}
	effectstest.DftCGPickRefused(t, g, me.ID, stone)
	effectstest.PickTriggerTarget(t, g, me.ID, wagon)
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"lifelink", "indestructible"} {
		if !effectstest.HasEffectiveKeyword(t, g, wagon, kw) {
			t.Errorf("the Vehicle has no %s", kw)
		}
	}
	effectstest.DftMPPassTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, wagon, "indestructible") {
		t.Error("indestructible outlived the turn")
	}
}

func TestDftHLLotusguardDiscipleCanTargetItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	disciple := effectstest.DftABCast(t, g, me, game.Card{Name: "Lotusguard Disciple", TypeLine: "Creature — Bird Cleric",
		OracleID: dftHLLotusguardOracle, Power: 2, Toughness: 2})
	effectstest.PickTriggerTarget(t, g, me.ID, disciple)
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"lifelink", "indestructible"} {
		if !effectstest.HasEffectiveKeyword(t, g, disciple, kw) {
			t.Errorf("the Disciple has no %s", kw)
		}
	}
}
