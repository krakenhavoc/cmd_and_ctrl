package testsav

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_h_cards_test.go — Bloomburrow cards starting with H (#4467). Each
// test plays the real catalog entry and asserts what its text says.

const (
	blbHHarnesserOracle  = "fb8f7047-280b-464f-aca2-6add3e21df8f"
	blbHHostOracle       = "56208bdc-2eed-4ccc-8c34-c0eba9db0aa9"
	blbHHerbalistOracle  = "f40cb1af-3db6-4e90-830f-86ff4b60e239"
	blbHNocturneOracle   = "2a74a055-ff7f-434b-8dfb-d9fb72579cee"
	blbHHomesteadOracle  = "467d6acb-9a3c-433f-836d-54e0e670b7d0"
	blbHBattlerOracle    = "beb02f49-95a6-49bf-b176-4e4b2ab5b651"
	blbHHelgaOracle      = "0a0d25a8-4f28-453b-9270-4d449fe0ca65"
	blbHGrottoOracle     = "382d18a2-438e-4ae7-a83f-1658ef1f9b07"
	blbHHighStrideOracle = "ab0aea13-1dbc-424a-ab7e-cdb303036a56"
	blbHHiredClawOracle  = "031cfc7d-8bda-454b-b88a-15d3f5fe47e2"
	blbHWolverineOracle  = "312be209-dba5-4e6c-938e-1287642f67d4"
	blbHHuntersOracle    = "a8a9b103-0cb8-4810-a2d7-42f6a10f80bd"
	blbHSwarmOracle      = "faefb352-575c-49f7-9f7a-96f2279a077b"
)

func TestBlbHEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		blbHHarnesserOracle:  "Harnesser of Storms",
		blbHHostOracle:       "Harvestrite Host",
		blbHHerbalistOracle:  "Hazardroot Herbalist",
		blbHNocturneOracle:   "Hazel's Nocturne",
		blbHHomesteadOracle:  "Head of the Homestead",
		blbHBattlerOracle:    "Hearthborn Battler",
		blbHHelgaOracle:      "Helga, Skittish Seer",
		blbHGrottoOracle:     "Hidden Grotto",
		blbHHighStrideOracle: "High Stride",
		blbHHiredClawOracle:  "Hired Claw",
		blbHWolverineOracle:  "Hivespine Wolverine",
		blbHHuntersOracle:    "Hunter's Talent",
		blbHSwarmOracle:      "Huskburster Swarm",
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

// blbHAnswerPlayer answers chooser's open target pick with a player.
func blbHAnswerPlayer(t *testing.T, g *game.Game, chooser, player uuid.UUID) error {
	t.Helper()
	p := effectstest.LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatal("no target prompt is open")
	}
	return g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player})
}

// --- Harnesser of Storms ------------------------------------------------------

func TestBlbHHarnesserImpulsesOnANoncreatureSpellOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Harnesser of Storms", "Creature — Otter Wizard", blbHHarnesserOracle, 1, 4)
	top := effectstest.SeedLibrary(me, "Top", "Next")
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top[0]) {
		t.Fatal("the top card was not exiled")
	}
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, top[0]), me.ID) {
		t.Error("the exiled card is not playable this turn")
	}
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil || g.Exile.Contains(top[1]) {
		t.Error("the ability triggered a second time this turn")
	}
}

func TestBlbHHarnesserTriggersOnAnOtterCreatureSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Harnesser of Storms", "Creature — Otter Wizard", blbHHarnesserOracle, 1, 4)
	if _, err := effectstest.BlbHCastWithMana(t, g, "Bear", "Creature — Bear", "", "", 2, 2); err != nil {
		t.Fatalf("cast the Bear: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a non-Otter creature spell triggered the Harnesser")
	}
	if _, err := effectstest.BlbHCastWithMana(t, g, "Otter", "Creature — Otter", "", "", 2, 2); err != nil {
		t.Fatalf("cast the Otter: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Harvestrite Host ---------------------------------------------------------

func TestBlbHHarvestriteHostDrawsOnTheSecondResolutionOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastHoldCreature(t, g, "Harvestrite Host", "Creature — Rabbit Citizen", blbHHostOracle, 3, 3)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 3 {
		t.Errorf("the bear has power %d, want 3", p)
	}
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("the first resolution drew: hand %d, want %d", got, hand)
	}
	effectstest.CastHoldCreature(t, g, "Rabbit", "Creature — Rabbit", "", 1, 1)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("the second resolution: hand %d, want %d", got, hand+1)
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 4 {
		t.Errorf("the bear has power %d, want 4", p)
	}
	effectstest.CastHoldCreature(t, g, "Rabbit", "Creature — Rabbit", "", 1, 1)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("the third resolution drew: hand %d, want %d", got, hand+1)
	}
}

func TestBlbHHarvestriteHostIgnoresANonRabbit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.BlbB1Push(g, me.ID, "Harvestrite Host", "Creature — Rabbit Citizen", blbHHostOracle, 3, 3)
	effectstest.CastHoldCreature(t, g, "Other Bear", "Creature — Bear", "", 2, 2)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a Bear entering triggered the Host")
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("the bear has power %d, want 2", p)
	}
}

// --- Hazardroot Herbalist -----------------------------------------------------

func TestBlbHHazardrootHerbalistGivesATokenDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	herbalist := effectstest.BlbB1Push(g, me.ID, "Hazardroot Herbalist", "Creature — Rabbit Druid", blbHHerbalistOracle, 1, 4)
	token := effectstest.BlbB1Push(g, me.ID, "Rabbit", "Token Creature — Rabbit", "", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, herbalist)
	effectstest.AnswerPickTarget(t, g, token)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, token); p != 2 || tough != 1 {
		t.Errorf("the token is %d/%d, want 2/1", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, token, "deathtouch") {
		t.Error("the token should have deathtouch")
	}
}

func TestBlbHHazardrootHerbalistGivesANontokenNoDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	herbalist := effectstest.BlbB1Push(g, me.ID, "Hazardroot Herbalist", "Creature — Rabbit Druid", blbHHerbalistOracle, 1, 4)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, herbalist)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the bear is %d/%d, want 3/2", p, tough)
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "deathtouch") {
		t.Error("a nontoken creature gained deathtouch")
	}
}

// --- Hazel's Nocturne ---------------------------------------------------------

func TestBlbHHazelsNocturneReturnsTwoAndDrainsEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.GraveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	b := effectstest.GraveyardCardOf(g, me, "Dead Elf", "Creature — Elf", "{G}", "")
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Hazel's Nocturne", "Instant", blbHNocturneOracle, effectstest.CardRefs(a, b)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the two creature cards did not return to hand")
	}
	for _, p := range g.Seats {
		want := before[p.ID] - 2
		if p.ID == me.ID {
			want = before[p.ID] + 2
		}
		if p.Life != want {
			t.Errorf("seat %s at %d, want %d", p.ID, p.Life, want)
		}
	}
}

func TestBlbHHazelsNocturneRefusesANoncreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := effectstest.GraveyardCardOf(g, me, "Dead Spell", "Sorcery", "{1}", "")
	if err := effectstest.CastCatalogSpellErr(t, g, "Hazel's Nocturne", "Instant", blbHNocturneOracle, effectstest.CardRefs(spell)); err == nil {
		t.Error("a sorcery card was accepted as a creature card")
	}
}

// --- Head of the Homestead ----------------------------------------------------

func TestBlbHHeadOfTheHomesteadMakesTwoRabbits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.CastHoldCreature(t, g, "Head of the Homestead", "Creature — Rabbit Citizen", blbHHomesteadOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Rabbit"); n != 2 {
		t.Errorf("%d Rabbit tokens, want 2", n)
	}
}

// --- Hearthborn Battler -------------------------------------------------------

func TestBlbHHearthbornBattlerHitsAnOpponentOnTheSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	battler := effectstest.BlbB1Push(g, me.ID, "Hearthborn Battler", "Creature — Lizard Warlock", blbHBattlerOracle, 2, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, battler, "haste") {
		t.Error("the Battler should have haste")
	}
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || opp.Life != life {
		t.Fatal("the first spell of the turn triggered the Battler")
	}
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	if err := blbHAnswerPlayer(t, g, me.ID, me.ID); err == nil {
		t.Fatal("the Battler's controller was accepted as an opponent")
	}
	if err := blbHAnswerPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || opp.Life != life-2 {
		t.Error("the third spell of the turn triggered the Battler")
	}
}

// --- Helga, Skittish Seer -----------------------------------------------------

func TestBlbHHelgaDrawsGainsAndGrowsOnABigCreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	helga := effectstest.BlbB1Push(g, me.ID, "Helga, Skittish Seer", "Legendary Creature — Frog Druid", blbHHelgaOracle, 1, 3)
	hand, life := me.Hand.Size(), me.Life
	if _, err := effectstest.BlbHCastWithMana(t, g, "Small", "Creature — Bear", "", "{3}", 2, 2, "C", "C", "C"); err != nil {
		t.Fatalf("cast the small creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, helga) != 0 || me.Life != life {
		t.Fatal("a mana value 3 creature spell triggered Helga")
	}
	if _, err := effectstest.BlbHCastWithMana(t, g, "Big", "Creature — Beast", "", "{4}", 4, 4, "C", "C", "C", "C"); err != nil {
		t.Fatalf("cast the big creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, helga); got != 1 {
		t.Errorf("Helga has %d +1/+1 counters, want 1", got)
	}
	if me.Life != life+1 {
		t.Errorf("I am at %d, want %d", me.Life, life+1)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestBlbHHelgaTapsForHerPowerInOneRestrictedColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _, _ := effectstest.BlbB1Seats(g)
	helga := effectstest.BlbB1Push(g, me.ID, "Helga, Skittish Seer", "Legendary Creature — Frog Druid", blbHHelgaOracle, 1, 3)
	g.WithWriteLock(func() {
		_ = effects.AddCounter{Target: helga, Kind: game.CounterPlusOne, N: 3}.Apply(effects.KeyedContext(g, me.ID, helga))
	})
	if err := g.ActivateManaAbility(me.ID, helga, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	units := me.ManaPool.Units()
	if len(units) != 4 {
		t.Fatalf("pool has %d mana, want 4 (Helga's power)", len(units))
	}
	for _, u := range units {
		if u.Color != "G" || len(u.Restrictions) != 1 {
			t.Errorf("mana %+v, want one restricted {G}", u)
		}
	}
	if _, err := effectstest.BlbHCastWithMana(t, g, "Small", "Creature — Bear", "", "{2}", 2, 2); !errors.Is(err, game.ErrInsufficientMana) {
		t.Errorf("a mana value 2 creature spell with only Helga's mana: %v, want insufficient mana", err)
	}
	if _, err := effectstest.BlbHCastWithMana(t, g, "Spell", "Sorcery", "", "{4}", 0, 0); !errors.Is(err, game.ErrInsufficientMana) {
		t.Errorf("a mana value 4 sorcery with only Helga's mana: %v, want insufficient mana", err)
	}
	if _, err := effectstest.BlbHCastWithMana(t, g, "Big", "Creature — Beast", "", "{4}", 4, 4); err != nil {
		t.Errorf("Helga's mana should pay for a mana value 4 creature spell: %v", err)
	}
}

// --- Hidden Grotto ------------------------------------------------------------

func TestBlbHHiddenGrottoEntersUntappedSurveilsAndTapsForColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Top", "Next")
	id := effectstest.PlayLandFromHand(t, g, "Hidden Grotto", blbHGrottoOracle)
	if c, _ := effectstest.BattlefieldCard(g, id); c.Tapped {
		t.Error("Hidden Grotto entered tapped")
	}
	if effectstest.TriggerOnStack(g, id) == nil {
		t.Fatal("the surveil trigger did not go on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Fatal("resolving the trigger queued no surveil prompt")
	}
}

func TestBlbHHiddenGrottoFiltersOneIntoAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	grotto := effectstest.SeedPermanentWithOracle(g, me.ID, "Hidden Grotto", "Land", blbHGrottoOracle)
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateManaAbility(me.ID, grotto, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want {G}", got)
	}
}

// --- High Stride --------------------------------------------------------------

func TestBlbHHighStridePumpsGrantsReachAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Tapped: true, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "High Stride", "Instant", blbHHighStrideOracle, effectstest.CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, theirs); p != 3 || tough != 5 {
		t.Errorf("the target is %d/%d, want 3/5", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, theirs, "reach") {
		t.Error("the target should have reach")
	}
	if effectstest.B16Tapped(t, g, theirs) {
		t.Error("the target should be untapped")
	}
}

// --- Hired Claw ---------------------------------------------------------------

func TestBlbHHiredClawPingsWhenYouAttackWithALizard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	claw := effectstest.BlbB1Push(g, me.ID, "Hired Claw", "Creature — Lizard Mercenary", blbHHiredClawOracle, 1, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, claw, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the counter ability activated before an opponent lost life")
	}
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, claw)
	if err := blbHAnswerPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Fatalf("the opponent is at %d, want %d", opp.Life, life-1)
	}
	effectstest.B06AddMana(me, "C", "R")
	effectstest.B16Activate(t, g, me.ID, claw, 0, game.ActivateAbilityParams{})
	if got := effectstest.PlusOnes(g, claw); got != 1 {
		t.Fatalf("the Claw has %d +1/+1 counters, want 1", got)
	}
	effectstest.B06AddMana(me, "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, claw, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the counter ability activated twice in one turn")
	}
}

func TestBlbHHiredClawIgnoresAnAttackWithoutALizard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Hired Claw", "Creature — Lizard Mercenary", blbHHiredClawOracle, 1, 2)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("an attack with no Lizard triggered the Claw")
	}
}

// --- Hivespine Wolverine ------------------------------------------------------

// blbHWolverineMode casts the Wolverine and picks `mode` on its entry
// trigger, leaving the mode's target prompt open.
func blbHWolverineMode(t *testing.T, g *game.Game, me uuid.UUID, mode int) uuid.UUID {
	t.Helper()
	wolverine := effectstest.CastRingCreature(t, g, "Hivespine Wolverine", "Creature — Elemental Wolverine", blbHWolverineOracle, 5, 4)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	return wolverine
}

func TestBlbHHivespineWolverineFightsACreatureToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	token := effectstest.BlbB1Push(g, opp.ID, "Squirrel", "Token Creature — Squirrel", "", 2, 2)
	wolverine := blbHWolverineMode(t, g, me.ID, 1)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the fight")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a nontoken creature was accepted as a creature token")
	}
	effectstest.AnswerPickTarget(t, g, token)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the token survived 5 damage")
	}
	if got := effectstest.DamageOn(t, g, wolverine); got != 2 {
		t.Errorf("the Wolverine has %d damage, want 2", got)
	}
}

func TestBlbHHivespineWolverinePutsACounterOrDestroys(t *testing.T) {
	t.Run("counter", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, _, _ := effectstest.BlbB1Seats(g)
		bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		blbHWolverineMode(t, g, me.ID, 0)
		effectstest.AnswerPickTarget(t, g, bear)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.PlusOnes(g, bear); got != 1 {
			t.Errorf("the bear has %d +1/+1 counters, want 1", got)
		}
	})
	t.Run("destroy", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := effectstest.BlbB1Seats(g)
		rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
		blbHWolverineMode(t, g, me.ID, 2)
		effectstest.AnswerPickTarget(t, g, rock)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) {
			t.Error("the artifact was not destroyed")
		}
	})
}

// --- Hunter's Talent ----------------------------------------------------------

func TestBlbHHuntersTalentBitesAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 3, 3)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Hunter's Talent", "Enchantment — Class", blbHHuntersOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	first := effectstest.LatestPickTarget(g, me.ID)
	if first == nil {
		t.Fatal("no prompt for the creature you control")
	}
	if err := g.ResolvePickTargets(first.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err != nil {
		t.Fatalf("pick mine: %v", err)
	}
	second := effectstest.LatestPickTarget(g, me.ID)
	if second == nil {
		t.Fatal("no prompt for the creature you don't control")
	}
	if err := g.ResolvePickTargets(second.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs, Slot: 1}}); err != nil {
		t.Fatalf("pick theirs: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("their 3/3 survived 3 damage")
	}
	if got := effectstest.DamageOn(t, g, mine); got != 0 {
		t.Errorf("my creature took %d damage; the bite is one-sided", got)
	}
}

func TestBlbHHuntersTalentLevelTwoPumpsAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	class := effectstest.PushClass(g, me.ID, "Hunter's Talent", "Enchantment — Class", blbHHuntersOracle)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.LevelUpTo(t, g, me.ID, class, 0, "{1}{G}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if got := effectstest.LevelOf(g, class); got != 2 {
		t.Fatalf("level %d, want 2", got)
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the bear is %d/%d, want 3/2", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("the bear should have trample")
	}
}

func TestBlbHHuntersTalentLevelOneHasNoAttackTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.PushClass(g, me.ID, "Hunter's Talent", "Enchantment — Class", blbHHuntersOracle)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the level 2 trigger fired at level 1")
	}
}

func TestBlbHHuntersTalentLevelThreeDrawsWithABigCreature(t *testing.T) {
	for _, tc := range []struct {
		name  string
		power int
		draws int
	}{{"power 4", 4, 1}, {"power 3", 3, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _, _ := effectstest.BlbB1Seats(g)
			class := effectstest.PushClass(g, me.ID, "Hunter's Talent", "Enchantment — Class", blbHHuntersOracle)
			effectstest.BlbB1Push(g, me.ID, "Beast", "Creature — Beast", "", tc.power, 4)
			effectstest.AdvanceToMain(t, g)
			if err := effectstest.LevelUpTo(t, g, me.ID, class, 0, "{1}{G}"); err != nil {
				t.Fatalf("level 2: %v", err)
			}
			if err := effectstest.LevelUpTo(t, g, me.ID, class, 1, "{3}{G}"); err != nil {
				t.Fatalf("level 3: %v", err)
			}
			hand := me.Hand.Size()
			effectstest.AdvanceTo(t, g, game.StepEnd)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size(); got != hand+tc.draws {
				t.Errorf("hand %d, want %d", got, hand+tc.draws)
			}
		})
	}
}

// --- Huskburster Swarm --------------------------------------------------------

func TestBlbHHuskburstSwarmCostsLessPerCreatureCardInExileAndGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	for _, name := range []string{"Dead One", "Dead Two", "Dead Three"} {
		effectstest.GraveyardCardOf(g, me, name, "Creature — Bear", "{1}", "")
	}
	effectstest.GraveyardCardOf(g, me, "Dead Spell", "Sorcery", "{1}", "")
	exiled := func(owner uuid.UUID, name, typeLine string, faceDown bool) {
		c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: owner, Controller: owner}
		if faceDown {
			c.SetFaceDown(game.FaceDownExiled)
		}
		g.Exile.PushTop(c)
	}
	exiled(me.ID, "Exiled One", "Creature — Bear", false)
	exiled(me.ID, "Exiled Two", "Creature — Bear", false)
	exiled(me.ID, "Face Down", "Creature — Bear", true)
	exiled(opp.ID, "Theirs", "Creature — Bear", false)
	exiled(me.ID, "Exiled Spell", "Instant", false)
	// Five count: {7}{B} less {5} is {2}{B}.
	if _, err := effectstest.BlbHCastWithMana(t, g, "Huskburster Swarm", "Creature — Elemental Insect", blbHSwarmOracle, "{7}{B}", 6, 6, "B", "C"); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("cast for {1}{B}: %v, want insufficient mana (five creature cards count)", err)
	}
	me.ManaPool.EmptyPool()
	swarm, err := effectstest.BlbHCastWithMana(t, g, "Huskburster Swarm", "Creature — Elemental Insect", blbHSwarmOracle, "{7}{B}", 6, 6, "B", "C", "C")
	if err != nil {
		t.Fatalf("cast for {2}{B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"menace", "deathtouch"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, swarm, kw) {
			t.Errorf("the Swarm should have %s", kw)
		}
	}
}
