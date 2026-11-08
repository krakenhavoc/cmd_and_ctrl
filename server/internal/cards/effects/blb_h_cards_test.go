package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

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

// blbHAnswerPlayer answers chooser's open target pick with a player.
func blbHAnswerPlayer(t *testing.T, g *game.Game, chooser, player uuid.UUID) error {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatal("no target prompt is open")
	}
	return g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player})
}

// blbHCastWithMana puts a card with a printed mana cost in the active
// player's hand, fills the pool with `pool` and casts it under strict
// mana, so a pool that can't pay the effective cost is refused.
func blbHCastWithMana(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, power, toughness int, pool ...string) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	b06AddMana(me, pool...)
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
}

// --- Harnesser of Storms ------------------------------------------------------

func TestBlbHHarnesserImpulsesOnANoncreatureSpellOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Harnesser of Storms", "Creature — Otter Wizard", blbHHarnesserOracle, 1, 4)
	top := seedLibrary(me, "Top", "Next")
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top[0]) {
		t.Fatal("the top card was not exiled")
	}
	if !permissionLive(g, exiledPermission(g, top[0]), me.ID) {
		t.Error("the exiled card is not playable this turn")
	}
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil || g.Exile.Contains(top[1]) {
		t.Error("the ability triggered a second time this turn")
	}
}

func TestBlbHHarnesserTriggersOnAnOtterCreatureSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Harnesser of Storms", "Creature — Otter Wizard", blbHHarnesserOracle, 1, 4)
	if _, err := blbHCastWithMana(t, g, "Bear", "Creature — Bear", "", "", 2, 2); err != nil {
		t.Fatalf("cast the Bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a non-Otter creature spell triggered the Harnesser")
	}
	if _, err := blbHCastWithMana(t, g, "Otter", "Creature — Otter", "", "", 2, 2); err != nil {
		t.Fatalf("cast the Otter: %v", err)
	}
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
}

// --- Harvestrite Host ---------------------------------------------------------

func TestBlbHHarvestriteHostDrawsOnTheSecondResolutionOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Harvestrite Host", "Creature — Rabbit Citizen", blbHHostOracle, 3, 3)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, bear); p != 3 {
		t.Errorf("the bear has power %d, want 3", p)
	}
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("the first resolution drew: hand %d, want %d", got, hand)
	}
	castHoldCreature(t, g, "Rabbit", "Creature — Rabbit", "", 1, 1)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("the second resolution: hand %d, want %d", got, hand+1)
	}
	if p, _ := ptOf(t, g, bear); p != 4 {
		t.Errorf("the bear has power %d, want 4", p)
	}
	castHoldCreature(t, g, "Rabbit", "Creature — Rabbit", "", 1, 1)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("the third resolution drew: hand %d, want %d", got, hand+1)
	}
}

func TestBlbHHarvestriteHostIgnoresANonRabbit(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	blbB1Push(g, me.ID, "Harvestrite Host", "Creature — Rabbit Citizen", blbHHostOracle, 3, 3)
	castHoldCreature(t, g, "Other Bear", "Creature — Bear", "", 2, 2)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a Bear entering triggered the Host")
	}
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Errorf("the bear has power %d, want 2", p)
	}
}

// --- Hazardroot Herbalist -----------------------------------------------------

func TestBlbHHazardrootHerbalistGivesATokenDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	herbalist := blbB1Push(g, me.ID, "Hazardroot Herbalist", "Creature — Rabbit Druid", blbHHerbalistOracle, 1, 4)
	token := blbB1Push(g, me.ID, "Rabbit", "Token Creature — Rabbit", "", 1, 1)
	declareAttack(t, g, opp.ID, herbalist)
	answerPickTarget(t, g, token)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, token); p != 2 || tough != 1 {
		t.Errorf("the token is %d/%d, want 2/1", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, token, "deathtouch") {
		t.Error("the token should have deathtouch")
	}
}

func TestBlbHHazardrootHerbalistGivesANontokenNoDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	herbalist := blbB1Push(g, me.ID, "Hazardroot Herbalist", "Creature — Rabbit Druid", blbHHerbalistOracle, 1, 4)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	declareAttack(t, g, opp.ID, herbalist)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the bear is %d/%d, want 3/2", p, tough)
	}
	if hasKeywordOnBattlefield(t, g, bear, "deathtouch") {
		t.Error("a nontoken creature gained deathtouch")
	}
}

// --- Hazel's Nocturne ---------------------------------------------------------

func TestBlbHHazelsNocturneReturnsTwoAndDrainsEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := graveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	b := graveyardCardOf(g, me, "Dead Elf", "Creature — Elf", "{G}", "")
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	if err := castCatalogSpellErr(t, g, "Hazel's Nocturne", "Instant", blbHNocturneOracle, cardRefs(a, b)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := graveyardCardOf(g, me, "Dead Spell", "Sorcery", "{1}", "")
	if err := castCatalogSpellErr(t, g, "Hazel's Nocturne", "Instant", blbHNocturneOracle, cardRefs(spell)); err == nil {
		t.Error("a sorcery card was accepted as a creature card")
	}
}

// --- Head of the Homestead ----------------------------------------------------

func TestBlbHHeadOfTheHomesteadMakesTwoRabbits(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	castHoldCreature(t, g, "Head of the Homestead", "Creature — Rabbit Citizen", blbHHomesteadOracle, 3, 2)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Rabbit"); n != 2 {
		t.Errorf("%d Rabbit tokens, want 2", n)
	}
}

// --- Hearthborn Battler -------------------------------------------------------

func TestBlbHHearthbornBattlerHitsAnOpponentOnTheSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	battler := blbB1Push(g, me.ID, "Hearthborn Battler", "Creature — Lizard Warlock", blbHBattlerOracle, 2, 3)
	if !hasKeywordOnBattlefield(t, g, battler, "haste") {
		t.Error("the Battler should have haste")
	}
	life := opp.Life
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || opp.Life != life {
		t.Fatal("the first spell of the turn triggered the Battler")
	}
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	if err := blbHAnswerPlayer(t, g, me.ID, me.ID); err == nil {
		t.Fatal("the Battler's controller was accepted as an opponent")
	}
	if err := blbHAnswerPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || opp.Life != life-2 {
		t.Error("the third spell of the turn triggered the Battler")
	}
}

// --- Helga, Skittish Seer -----------------------------------------------------

func TestBlbHHelgaDrawsGainsAndGrowsOnABigCreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	helga := blbB1Push(g, me.ID, "Helga, Skittish Seer", "Legendary Creature — Frog Druid", blbHHelgaOracle, 1, 3)
	hand, life := me.Hand.Size(), me.Life
	if _, err := blbHCastWithMana(t, g, "Small", "Creature — Bear", "", "{3}", 2, 2, "C", "C", "C"); err != nil {
		t.Fatalf("cast the small creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, helga) != 0 || me.Life != life {
		t.Fatal("a mana value 3 creature spell triggered Helga")
	}
	if _, err := blbHCastWithMana(t, g, "Big", "Creature — Beast", "", "{4}", 4, 4, "C", "C", "C", "C"); err != nil {
		t.Fatalf("cast the big creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, helga); got != 1 {
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
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, _, _ := blbB1Seats(g)
	helga := blbB1Push(g, me.ID, "Helga, Skittish Seer", "Legendary Creature — Frog Druid", blbHHelgaOracle, 1, 3)
	g.WithWriteLock(func() {
		_ = AddCounter{Target: helga, Kind: game.CounterPlusOne, N: 3}.Apply(keyedContext(g, me.ID, helga))
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
	if _, err := blbHCastWithMana(t, g, "Small", "Creature — Bear", "", "{2}", 2, 2); !errors.Is(err, game.ErrInsufficientMana) {
		t.Errorf("a mana value 2 creature spell with only Helga's mana: %v, want insufficient mana", err)
	}
	if _, err := blbHCastWithMana(t, g, "Spell", "Sorcery", "", "{4}", 0, 0); !errors.Is(err, game.ErrInsufficientMana) {
		t.Errorf("a mana value 4 sorcery with only Helga's mana: %v, want insufficient mana", err)
	}
	if _, err := blbHCastWithMana(t, g, "Big", "Creature — Beast", "", "{4}", 4, 4); err != nil {
		t.Errorf("Helga's mana should pay for a mana value 4 creature spell: %v", err)
	}
}

// --- Hidden Grotto ------------------------------------------------------------

func TestBlbHHiddenGrottoEntersUntappedSurveilsAndTapsForColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top", "Next")
	id := playLandFromHand(t, g, "Hidden Grotto", blbHGrottoOracle)
	if c, _ := battlefieldCard(g, id); c.Tapped {
		t.Error("Hidden Grotto entered tapped")
	}
	if triggerOnStack(g, id) == nil {
		t.Fatal("the surveil trigger did not go on the stack")
	}
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Fatal("resolving the trigger queued no surveil prompt")
	}
}

func TestBlbHHiddenGrottoFiltersOneIntoAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	grotto := seedPermanentWithOracle(g, me.ID, "Hidden Grotto", "Land", blbHGrottoOracle)
	b06AddMana(me, "C")
	if err := g.ActivateManaAbility(me.ID, grotto, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want {G}", got)
	}
}

// --- High Stride --------------------------------------------------------------

func TestBlbHHighStridePumpsGrantsReachAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Tapped: true, Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "High Stride", "Instant", blbHHighStrideOracle, cardTarget(theirs))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, theirs); p != 3 || tough != 5 {
		t.Errorf("the target is %d/%d, want 3/5", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, theirs, "reach") {
		t.Error("the target should have reach")
	}
	if b16Tapped(t, g, theirs) {
		t.Error("the target should be untapped")
	}
}

// --- Hired Claw ---------------------------------------------------------------

func TestBlbHHiredClawPingsWhenYouAttackWithALizard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	claw := blbB1Push(g, me.ID, "Hired Claw", "Creature — Lizard Mercenary", blbHHiredClawOracle, 1, 2)
	advanceToMain(t, g)
	b06AddMana(me, "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, claw, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the counter ability activated before an opponent lost life")
	}
	life := opp.Life
	declareAttack(t, g, opp.ID, claw)
	if err := blbHAnswerPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Fatalf("the opponent is at %d, want %d", opp.Life, life-1)
	}
	b06AddMana(me, "C", "R")
	b16Activate(t, g, me.ID, claw, 0, game.ActivateAbilityParams{})
	if got := plusOnes(g, claw); got != 1 {
		t.Fatalf("the Claw has %d +1/+1 counters, want 1", got)
	}
	b06AddMana(me, "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, claw, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the counter ability activated twice in one turn")
	}
}

func TestBlbHHiredClawIgnoresAnAttackWithoutALizard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Hired Claw", "Creature — Lizard Mercenary", blbHHiredClawOracle, 1, 2)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an attack with no Lizard triggered the Claw")
	}
}

// --- Hivespine Wolverine ------------------------------------------------------

// blbHWolverineMode casts the Wolverine and picks `mode` on its entry
// trigger, leaving the mode's target prompt open.
func blbHWolverineMode(t *testing.T, g *game.Game, me uuid.UUID, mode int) uuid.UUID {
	t.Helper()
	wolverine := castRingCreature(t, g, "Hivespine Wolverine", "Creature — Elemental Wolverine", blbHWolverineOracle, 5, 4)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	return wolverine
}

func TestBlbHHivespineWolverineFightsACreatureToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	token := blbB1Push(g, opp.ID, "Squirrel", "Token Creature — Squirrel", "", 2, 2)
	wolverine := blbHWolverineMode(t, g, me.ID, 1)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the fight")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a nontoken creature was accepted as a creature token")
	}
	answerPickTarget(t, g, token)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the token survived 5 damage")
	}
	if got := damageOn(t, g, wolverine); got != 2 {
		t.Errorf("the Wolverine has %d damage, want 2", got)
	}
}

func TestBlbHHivespineWolverinePutsACounterOrDestroys(t *testing.T) {
	t.Run("counter", func(t *testing.T) {
		g := newCatalogGame(t)
		me, _, _ := blbB1Seats(g)
		bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		blbHWolverineMode(t, g, me.ID, 0)
		answerPickTarget(t, g, bear)
		passPriorityAroundTable(t, g)
		if got := plusOnes(g, bear); got != 1 {
			t.Errorf("the bear has %d +1/+1 counters, want 1", got)
		}
	})
	t.Run("destroy", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
		blbHWolverineMode(t, g, me.ID, 2)
		answerPickTarget(t, g, rock)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) {
			t.Error("the artifact was not destroyed")
		}
	})
}

// --- Hunter's Talent ----------------------------------------------------------

func TestBlbHHuntersTalentBitesAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 3, 3)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 3, 3)
	castCatalogSpell(t, g, "Hunter's Talent", "Enchantment — Class", blbHHuntersOracle, nil)
	passPriorityAroundTable(t, g)
	first := latestPickTarget(g, me.ID)
	if first == nil {
		t.Fatal("no prompt for the creature you control")
	}
	if err := g.ResolvePickTargets(first.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err != nil {
		t.Fatalf("pick mine: %v", err)
	}
	second := latestPickTarget(g, me.ID)
	if second == nil {
		t.Fatal("no prompt for the creature you don't control")
	}
	if err := g.ResolvePickTargets(second.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs, Slot: 1}}); err != nil {
		t.Fatalf("pick theirs: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("their 3/3 survived 3 damage")
	}
	if got := damageOn(t, g, mine); got != 0 {
		t.Errorf("my creature took %d damage; the bite is one-sided", got)
	}
}

func TestBlbHHuntersTalentLevelTwoPumpsAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	class := pushClass(g, me.ID, "Hunter's Talent", "Enchantment — Class", blbHHuntersOracle)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	if err := levelUpTo(t, g, me.ID, class, 0, "{1}{G}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if got := levelOf(g, class); got != 2 {
		t.Fatalf("level %d, want 2", got)
	}
	declareAttack(t, g, opp.ID, bear)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the bear is %d/%d, want 3/2", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("the bear should have trample")
	}
}

func TestBlbHHuntersTalentLevelOneHasNoAttackTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	pushClass(g, me.ID, "Hunter's Talent", "Enchantment — Class", blbHHuntersOracle)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	if latestPickTarget(g, me.ID) != nil {
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
			g := newCatalogGame(t)
			me, _, _ := blbB1Seats(g)
			class := pushClass(g, me.ID, "Hunter's Talent", "Enchantment — Class", blbHHuntersOracle)
			blbB1Push(g, me.ID, "Beast", "Creature — Beast", "", tc.power, 4)
			advanceToMain(t, g)
			if err := levelUpTo(t, g, me.ID, class, 0, "{1}{G}"); err != nil {
				t.Fatalf("level 2: %v", err)
			}
			if err := levelUpTo(t, g, me.ID, class, 1, "{3}{G}"); err != nil {
				t.Fatalf("level 3: %v", err)
			}
			hand := me.Hand.Size()
			advanceTo(t, g, game.StepEnd)
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size(); got != hand+tc.draws {
				t.Errorf("hand %d, want %d", got, hand+tc.draws)
			}
		})
	}
}

// --- Huskburster Swarm --------------------------------------------------------

func TestBlbHHuskburstSwarmCostsLessPerCreatureCardInExileAndGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	for _, name := range []string{"Dead One", "Dead Two", "Dead Three"} {
		graveyardCardOf(g, me, name, "Creature — Bear", "{1}", "")
	}
	graveyardCardOf(g, me, "Dead Spell", "Sorcery", "{1}", "")
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
	if _, err := blbHCastWithMana(t, g, "Huskburster Swarm", "Creature — Elemental Insect", blbHSwarmOracle, "{7}{B}", 6, 6, "B", "C"); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("cast for {1}{B}: %v, want insufficient mana (five creature cards count)", err)
	}
	me.ManaPool.EmptyPool()
	swarm, err := blbHCastWithMana(t, g, "Huskburster Swarm", "Creature — Elemental Insect", blbHSwarmOracle, "{7}{B}", 6, 6, "B", "C", "C")
	if err != nil {
		t.Fatalf("cast for {2}{B}: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"menace", "deathtouch"} {
		if !hasKeywordOnBattlefield(t, g, swarm, kw) {
			t.Errorf("the Swarm should have %s", kw)
		}
	}
}
