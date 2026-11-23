package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b02_cards_test.go — S82 SNC batch 2 (#5491): simple instants and
// sorceries. Every test plays the real catalog card through the engine.

const (
	sncB02Strangle  = "1f570526-7fb0-4fe5-8014-b5793eb65f87"
	sncB02Antagon   = "f4e860f1-e637-4640-b257-57c2e42793ce"
	sncB02Daring    = "c777002d-dcd4-4d75-8175-5440a8bb118e"
	sncB02Refuse    = "d05aeada-2a9f-464d-8a66-b6e90f43bae3"
	sncB02Due       = "bf65a5f5-5536-4616-a09e-2416439c7bc6"
	sncB02DealBad   = "e9ede348-f1dc-45ba-9e31-3129ccc182d1"
	sncB02Rumble    = "04dd9395-847b-46f7-a669-9b0df89d01ff"
	sncB02Family    = "f6abd7e5-8c52-4b36-98c5-a4df8b5b6863"
	sncB02Case      = "575d067a-83a0-49e7-8a74-e3e46a946c8c"
	sncB02Prize     = "c0c4ecf6-84e6-443e-8a81-e395b31c631d"
	sncB02Silence   = "2fd01a30-0c0e-4ac5-b96d-b2a6dad31f0c"
	sncB02RunOut    = "96189b18-8cab-4709-b3b6-c4f2003829eb"
	sncB02Aria      = "d031b482-b33d-4eb4-a5bc-ea96e66a1d99"
	sncB02Bear      = "Creature — Bear"
	sncB02TokenLine = "Token Creature — Soldier"
)

func TestSncB02EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		sncB02Strangle: "Strangle", sncB02Antagon: "Antagonize", sncB02Daring: "Daring Escape",
		sncB02Refuse: "Refuse to Yield", sncB02Due: "Demon's Due", sncB02DealBad: "Deal Gone Bad",
		sncB02Rumble: "Ready to Rumble", sncB02Family: "For the Family", sncB02Case: "Case the Joint",
		sncB02Prize: "Prizefight", sncB02Silence: "Buy Your Silence", sncB02RunOut: "Run Out of Town",
		sncB02Aria: "Incandescent Aria",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestSncB02StrangleDealsThreeToACreatureOrPlaneswalkerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	small := effectstest.B12Creature(g, opp.ID, "Small", sncB02Bear, 2, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", sncB02Bear, 2, 4)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Strangle", "Sorcery", sncB02Strangle, effectstest.B16TargetPlayer(me.ID)); err == nil {
		t.Error("targeted a player")
	}
	effectstest.CastCatalogSpell(t, g, "Strangle", "Sorcery", sncB02Strangle, effectstest.B16TargetCard(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("3 damage should kill a 3-toughness creature")
	}
	effectstest.CastCatalogSpell(t, g, "Strangle", "Sorcery", sncB02Strangle, effectstest.B16TargetCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) {
		t.Error("3 damage should not kill a 4-toughness creature")
	}
	pw := effectstest.B12Permanent(g, opp.ID, "Walker", "Legendary Planeswalker — Test")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Strangle", "Sorcery", sncB02Strangle, effectstest.B16TargetCard(pw)); err != nil {
		t.Errorf("a planeswalker is a legal target: %v", err)
	}
}

func TestSncB02AntagonizeGivesPlusFourPlusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", sncB02Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Antagonize", "Instant", sncB02Antagon, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 6 || tg != 5 {
		t.Errorf("Bear is %d/%d, want 6/5", p, tg)
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Antagonize", "Instant", sncB02Antagon, effectstest.B16TargetPlayer(me.ID)); err == nil {
		t.Error("targeted a player")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("the pump outlasted the turn: power %d", p)
	}
}

func TestSncB02DaringEscapePumpsGrantsFirstStrikeAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	bear := effectstest.B12Creature(g, me.ID, "Bear", sncB02Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Daring Escape", "Instant", sncB02Daring, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tg != 2 {
		t.Errorf("Bear is %d/%d, want 3/2", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("Bear lacks first strike")
	}
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("expected a scry 1 prompt: %+v", c)
	}
}

func TestSncB02RefuseToYieldPumpsAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", sncB02Bear, 2, 2)
	if err := g.TapCard(bear, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	effectstest.CastCatalogSpell(t, g, "Refuse to Yield", "Instant", sncB02Refuse, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tg != 9 {
		t.Errorf("Bear is %d/%d, want 4/9", p, tg)
	}
	if effectstest.IsTapped(g, bear) {
		t.Error("Bear should be untapped")
	}
}

func TestSncB02DemonsDueScriesThenDrawsAndLosesLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "Bin Me", "Keep Me", "Third", "Fourth")
	hand, life := effectstest.HandSize(me), me.Life
	effectstest.CastCatalogSpell(t, g, "Demon's Due", "Instant", sncB02Due, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("expected a scry 2 prompt: %+v", c)
	}
	if effectstest.HandSize(me) != hand {
		t.Fatalf("drew before the scry was answered: hand %d, started %d", effectstest.HandSize(me), hand)
	}
	// Keep both on top; bottom none.
	if err := g.ResolveScry(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != hand+2 { // the cast spell left the hand it was pushed into; two cards came
		t.Errorf("hand %d, want %d", effectstest.HandSize(me), hand+2)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

func TestSncB02DealGoneBadShrinksTheCreatureAndMillsThePlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", sncB02Bear, 3, 3)
	lib, gy := opp.Library.Size(), opp.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Deal Gone Bad", "Instant", sncB02DealBad, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Slot: 0},
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a 3/3 under -3/-3 should die")
	}
	if opp.Library.Size() != lib-3 || opp.Graveyard.Size() != gy+4 { // three milled, plus the dead Bear
		t.Errorf("library %d (want %d), graveyard %d (want %d)", opp.Library.Size(), lib-3, opp.Graveyard.Size(), gy+4)
	}
	// The player may be the caster; the creature clause may not be a player.
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Deal Gone Bad", "Instant", sncB02DealBad, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 0},
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 1},
	}); err == nil {
		t.Error("the first target must be a creature")
	}
}

// sncB02CastModal casts a modal spell with the chosen mode and targets.
func sncB02CastModal(t *testing.T, g *game.Game, name, typeLine, oracle string, mode int, targets []game.TargetRef, strict bool) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: active.ID, Controller: active.ID})
	effectstest.AdvanceToMain(t, g)
	return g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: targets, Strict: strict})
}

func TestSncB02ReadyToRumbleBothModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", sncB02Bear, 2, 5)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := sncB02CastModal(t, g, "Ready to Rumble", "Sorcery", sncB02Rumble, 0, effectstest.B16TargetCard(rock), true); err == nil {
		t.Error("the damage mode targeted an artifact that is not a creature")
	}
	if err := sncB02CastModal(t, g, "Ready to Rumble", "Sorcery", sncB02Rumble, 0, effectstest.B16TargetCard(bear), true); err != nil {
		t.Fatalf("damage mode: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("5 damage should kill a 5-toughness creature")
	}
	if err := sncB02CastModal(t, g, "Ready to Rumble", "Sorcery", sncB02Rumble, 1, effectstest.B16TargetCard(rock), true); err != nil {
		t.Fatalf("destroy mode: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact should be destroyed")
	}
	bear2 := effectstest.B12Creature(g, opp.ID, "Bear", sncB02Bear, 2, 5)
	if err := sncB02CastModal(t, g, "Ready to Rumble", "Sorcery", sncB02Rumble, 1, effectstest.B16TargetCard(bear2), true); err == nil {
		t.Error("the destroy mode targeted a nonartifact creature")
	}
}

func TestSncB02ForTheFamilyDependsOnFourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", sncB02Bear, 1, 1)
	effectstest.CastCatalogSpell(t, g, "For the Family", "Instant", sncB02Family, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("with one creature power is %d, want 3", p)
	}
	for i := 0; i < 3; i++ {
		effectstest.B12Creature(g, me.ID, "Extra", sncB02Bear, 1, 1)
	}
	other := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", sncB02Bear, 1, 1) // opponents' creatures do not count
	effectstest.CastCatalogSpell(t, g, "For the Family", "Instant", sncB02Family, effectstest.B16TargetCard(other))
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, other); p != 5 { // four creatures of ours: +4/+4 instead
		t.Errorf("with four creatures the target is %d power, want 5", p)
	}
}

func TestSncB02ForTheFamilyDoesNotCountOpponentCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", sncB02Bear, 1, 1)
	for i := 0; i < 4; i++ {
		effectstest.B12Creature(g, opp.ID, "Theirs", sncB02Bear, 1, 1)
	}
	effectstest.CastCatalogSpell(t, g, "For the Family", "Instant", sncB02Family, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, mine); p != 3 {
		t.Errorf("power %d, want 3: the opponent's creatures do not count", p)
	}
}

func TestSncB02CaseTheJointDrawsTwoAndLooksAtEveryTopCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hand := effectstest.HandSize(me)
	effectstest.CastCatalogSpell(t, g, "Case the Joint", "Instant", sncB02Case, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != hand+2 {
		t.Errorf("hand %d, want %d", effectstest.HandSize(me), hand+2)
	}
	for _, p := range g.Seats {
		top := p.Library.Cards[p.Library.Size()-1]
		if !top.IsKnownTo(me.ID) {
			t.Errorf("seat %s: the caster did not look at the top card", p.Name)
		}
		if p.ID != me.ID && top.IsKnownTo(opp.ID) && p.ID != opp.ID {
			t.Errorf("seat %s: an uninvolved player learned the top card", p.Name)
		}
	}
}

func TestSncB02PrizefightFightsAndCreatesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", sncB02Bear, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", sncB02Bear, 2, 2)
	before := effectstest.CountTreasuresControlledBy(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Prizefight", "Instant", sncB02Prize, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("their 2/2 should have died to 3 damage")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("our 3/3 took only 2 damage")
	}
	if got := effectstest.CountTreasuresControlledBy(g, me.ID); got != before+1 {
		t.Errorf("Treasures %d, want %d", got, before+1)
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Prizefight", "Instant", sncB02Prize, []game.TargetRef{
		{Kind: game.TargetCard, ID: theirs, Slot: 0}, {Kind: game.TargetCard, ID: mine, Slot: 1}}); err == nil {
		t.Error("fought with the wrong sides")
	}
}

func TestSncB02BuyYourSilenceExilesAndGivesItsControllerATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Buy Your Silence", "Sorcery", sncB02Silence, effectstest.B16TargetCard(land)); err == nil {
		t.Error("targeted a land")
	}
	effectstest.CastCatalogSpell(t, g, "Buy Your Silence", "Sorcery", sncB02Silence, effectstest.B16TargetCard(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the permanent should be exiled")
	}
	if got := effectstest.CountTreasuresControlledBy(g, opp.ID); got != 1 {
		t.Errorf("its controller has %d Treasures, want 1", got)
	}
	if got := effectstest.CountTreasuresControlledBy(g, me.ID); got != 0 {
		t.Errorf("the caster has %d Treasures, want 0", got)
	}
}

func TestSncB02RunOutOfTownLetsTheOwnerChoose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", sncB02Bear, 2, 2)
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Run Out of Town", "Instant", sncB02RunOut, effectstest.B16TargetCard(land)); err == nil {
		t.Error("targeted a land")
	}
	effectstest.CastCatalogSpell(t, g, "Run Out of Town", "Instant", sncB02RunOut, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
		t.Fatal("the caster was asked; the owner chooses")
	}
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if c == nil {
		t.Fatalf("the owner was not asked: %+v", g.PendingChoices)
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, []uuid.UUID{bear}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear should have left the battlefield")
	}
}

func TestSncB02IncandescentAriaSparesTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", sncB02Bear, 2, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", sncB02Bear, 2, 4)
	token := effectstest.B12Creature(g, opp.ID, "Soldier", sncB02TokenLine, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Incandescent Aria", "Sorcery", sncB02Aria, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("our nontoken 2/3 should die")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("a 4-toughness creature survives 3 damage")
	}
	if !g.Battlefield.Contains(token) {
		t.Error("a token takes no damage from Incandescent Aria")
	}
}
