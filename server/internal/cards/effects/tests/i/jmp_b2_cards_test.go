package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jmp_b2_cards_test.go — the Jumpstart Dimir deck request, second batch
// (#4892). Each test plays the real catalog entry through the engine and
// asserts what the printed text says, and what it refuses.

const (
	jmpB2LaunchPartyOracle = "f713a6b4-bd9c-4bc3-9e00-26ecf16e4a92"
	jmpB2StewardOracle     = "cc1203cb-6758-4454-b78b-341930ff4da1"
	jmpB2WaltzOracle       = "cf75dbd8-b65a-48d9-b96b-4afb43f336d1"
	jmpB2ScytheOracle      = "0b4925d7-e2bb-4b74-a538-80fe759799ad"
	jmpB2OneirophageOracle = "48c97aeb-6591-4175-a24b-d3488bf3ab0b"
	jmpB2BroodlingsOracle  = "7d5e4033-bd58-4854-b9d3-a440240995ee"
	jmpB2RusalkaOracle     = "d6912f52-e478-4cc6-a34c-67e372abc4fa"
	jmpB2IndulgenceOracle  = "010ec11a-0484-486d-b7c0-de65593bf457"
	jmpB2SkeletonOracle    = "d09b0648-e324-4a16-b78c-d1849177c04f"
	jmpB2ScourgeOracle     = "32972d7a-c7c6-437c-8e7c-6c7d56b130e4"
	jmpB2OccultistOracle   = "d72031e1-c8cf-4e57-b15b-6a5eaacff1e2"
)

func TestJmpB2EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		jmpB2LaunchPartyOracle: "Launch Party",
		jmpB2StewardOracle:     "Liliana's Steward",
		jmpB2WaltzOracle:       "Macabre Waltz",
		jmpB2ScytheOracle:      "Malefic Scythe",
		jmpB2OneirophageOracle: "Oneirophage",
		jmpB2BroodlingsOracle:  "Phyrexian Broodlings",
		jmpB2RusalkaOracle:     "Plagued Rusalka",
		jmpB2IndulgenceOracle:  "Sanguine Indulgence",
		jmpB2SkeletonOracle:    "Sanitarium Skeleton",
		jmpB2ScourgeOracle:     "Scourge of Nel Toth",
		jmpB2OccultistOracle:   "Selhoff Occultist",
	}
	for oracle, name := range want {
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

// jmpB2Cast puts a card with a printed cost in seat 0's hand on its main
// phase and casts it strictly (the mana must be in the pool).
func jmpB2Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	params.Strict = true
	return effectstest.Msc17Cast(t, g, p, c, params)
}

// --- Launch Party ------------------------------------------------------------

func jmpB2LaunchParty() game.Card {
	return game.Card{Name: "Launch Party", TypeLine: "Instant", ManaCost: "{3}{B}", OracleID: jmpB2LaunchPartyOracle}
}

func TestJmpB2LaunchPartySacrificesDestroysAndDrainsTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	before := opp.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.PayMana(t, g, me, "{C}{C}{C}{B}")
	if _, err := jmpB2Cast(t, g, me, jmpB2LaunchParty(), game.CastSpellParams{
		Targets: effectstest.CardRefs(victim), SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice was not paid as the spell was cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the target was not destroyed")
	}
	if opp.Life != before-2 {
		t.Errorf("the target's controller is at %d, want %d", opp.Life, before-2)
	}
}

func TestJmpB2LaunchPartyIsRefusedWithoutACreatureToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.PayMana(t, g, me, "{C}{C}{C}{B}")
	if _, err := jmpB2Cast(t, g, me, jmpB2LaunchParty(), game.CastSpellParams{Targets: effectstest.CardRefs(victim)}); err == nil {
		t.Fatal("cast with no creature sacrificed")
	}
	if !g.Battlefield.Contains(victim) {
		t.Error("the refused cast destroyed the target")
	}
}

func TestJmpB2LaunchPartyDoesNothingWhenItsTargetIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	before := me.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.PayMana(t, g, me, "{C}{C}{C}{B}")
	if _, err := jmpB2Cast(t, g, me, jmpB2LaunchParty(), game.CastSpellParams{
		Targets: effectstest.CardRefs(mine), SacrificeIDs: []uuid.UUID{mine},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("life %d -> %d: a spell whose only target is gone lost life anyway", before, me.Life)
	}
}

// --- Liliana's Steward -------------------------------------------------------

func TestJmpB2LilianasStewardSacrificesForATargetOpponentsDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	steward := effectstest.B12Push(g, me.ID, "Liliana's Steward", "Creature — Zombie", jmpB2StewardOracle, 1, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, steward, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if !effectstest.InGraveyard(me, steward) {
		t.Error("the Steward was not sacrificed")
	}
	if n := effectstest.DiscardOwed(g, opp.ID); n != 1 {
		t.Fatalf("the opponent owes %d discards, want 1", n)
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 0 {
		t.Errorf("the Steward's controller owes %d discards", n)
	}
}

func TestJmpB2LilianasStewardRefusesYouAndInstantSpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	steward := effectstest.B12Push(g, me.ID, "Liliana's Steward", "Creature — Zombie", jmpB2StewardOracle, 1, 2)
	// The harness stops on the draw step: not a main phase.
	if err := g.ActivateCatalogAbility(me.ID, steward, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, steward, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err == nil {
		t.Fatal("targeted its own controller, who is not an opponent")
	}
	if !g.Battlefield.Contains(steward) {
		t.Error("a refused activation sacrificed the Steward")
	}
}

// --- Macabre Waltz -----------------------------------------------------------

func jmpB2Waltz() game.Card {
	return game.Card{Name: "Macabre Waltz", TypeLine: "Sorcery", ManaCost: "{1}{B}", OracleID: jmpB2WaltzOracle}
}

func TestJmpB2MacabreWaltzReturnsTwoCreatureCardsThenDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.GraveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	b := effectstest.GraveyardCardOf(g, me, "Dead Elf", "Creature — Elf", "{G}", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.PayMana(t, g, me, "{C}{B}")
	if _, err := jmpB2Cast(t, g, me, jmpB2Waltz(), game.CastSpellParams{Targets: effectstest.CardRefs(a, b)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Fatal("the two creature cards did not return to hand")
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Fatalf("owes %d discards, want 1", n)
	}
	// A card just returned is a legal discard (CR 608.2c).
	effectstest.AnswerDiscard(t, g, me.ID, a)
	if !effectstest.InGraveyard(me, a) {
		t.Error("the discarded card is not in the graveyard")
	}
}

func TestJmpB2MacabreWaltzWithNoTargetStillDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PayMana(t, g, me, "{C}{B}")
	if _, err := jmpB2Cast(t, g, me, jmpB2Waltz(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast with no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Fatalf("owes %d discards, want 1", n)
	}
}

func TestJmpB2MacabreWaltzRefusesNoncreatureAndOpponentsCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := effectstest.GraveyardCardOf(g, me, "Dead Spell", "Sorcery", "{1}", "")
	theirs := effectstest.GraveyardCardOf(g, opp, "Their Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.AdvanceToMain(t, g)
	for _, bad := range []uuid.UUID{spell, theirs} {
		effectstest.PayMana(t, g, me, "{C}{B}")
		if _, err := jmpB2Cast(t, g, me, jmpB2Waltz(), game.CastSpellParams{Targets: effectstest.CardRefs(bad)}); err == nil {
			t.Fatalf("cast targeting %v", bad)
		}
		g.WithWriteLock(func() { me.ManaPool = nil })
	}
}

// --- Malefic Scythe ----------------------------------------------------------

func TestJmpB2MaleficScytheGrowsASoulCounterForEachEquippedDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PayMana(t, g, me, "{C}{B}")
	scythe, err := jmpB2Cast(t, g, me, game.Card{Name: "Malefic Scythe", TypeLine: "Artifact — Equipment", ManaCost: "{1}{B}", OracleID: jmpB2ScytheOracle}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, scythe, "soul"); n != 1 {
		t.Fatalf("entered with %d soul counters, want 1", n)
	}
	first := effectstest.B12Creature(g, me.ID, "First", "Creature — Bear", 2, 2)
	effectstest.PayMana(t, g, me, "{C}")
	effectstest.EquipTo(t, g, me.ID, scythe, first)
	if got := effectstest.EffectivePower(t, g, first); got != 3 {
		t.Fatalf("equipped power %d, want 2+1", got)
	}

	// An unequipped creature's death adds nothing.
	bystander := effectstest.B12Creature(g, me.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.P1Destroy(t, g, bystander)
	if !effectstest.StackFullyEmpty(g) {
		t.Fatal("an unequipped creature's death triggered the Scythe")
	}
	if n := effectstest.CounterCount(g, scythe, "soul"); n != 1 {
		t.Fatalf("an unequipped death made %d soul counters, want 1", n)
	}

	effectstest.P1Destroy(t, g, first)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, scythe, "soul"); n != 2 {
		t.Fatalf("after the equipped creature died: %d soul counters, want 2", n)
	}
	second := effectstest.B12Creature(g, me.ID, "Second", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g) // resolving the trigger passed into combat: equip again postcombat
	effectstest.PayMana(t, g, me, "{C}")
	effectstest.EquipTo(t, g, me.ID, scythe, second)
	if got := effectstest.EffectivePower(t, g, second); got != 4 {
		t.Errorf("re-equipped power %d, want 2+2", got)
	}
}

func TestJmpB2MaleficScytheEquipCostsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	scythe := effectstest.B12Push(g, me.ID, "Malefic Scythe", "Artifact — Equipment", jmpB2ScytheOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, scythe, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(bear), Strict: true,
	}); err == nil {
		t.Fatal("equipped with no mana")
	}
}

// --- Oneirophage -------------------------------------------------------------

func TestJmpB2OneirophageGrowsOnYourDrawsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	squid := effectstest.B12Push(g, me.ID, "Oneirophage", "Creature — Squid Illusion", jmpB2OneirophageOracle, 1, 2)
	effectstest.AssertKeywords(t, g, squid, "flying")
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 2; i++ {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatalf("draw: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.CounterCount(g, squid, game.CounterPlusOne); n != 2 {
		t.Fatalf("%d counters after two draws, want 2", n)
	}
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("opponent draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, squid, game.CounterPlusOne); n != 2 {
		t.Errorf("an opponent's draw grew it to %d counters", n)
	}
}

// --- Phyrexian Broodlings ----------------------------------------------------

func TestJmpB2PhyrexianBroodlingsGrowsForOneAndASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	brood := effectstest.B12Push(g, me.ID, "Phyrexian Broodlings", "Creature — Phyrexian Minion", jmpB2BroodlingsOracle, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, brood, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err == nil {
		t.Fatal("activated without paying {1}")
	}
	if !g.Battlefield.Contains(fodder) {
		t.Fatal("the refused activation sacrificed the fodder")
	}
	effectstest.PayMana(t, g, me, "{C}")
	effectstest.B16Activate(t, g, me.ID, brood, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the fodder was not sacrificed")
	}
	if n := effectstest.CounterCount(g, brood, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
}

// --- Plagued Rusalka ---------------------------------------------------------

func TestJmpB2PlaguedRusalkaShrinksATargetUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rusalka := effectstest.B12Push(g, me.ID, "Plagued Rusalka", "Creature — Spirit", jmpB2RusalkaOracle, 1, 1)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, rusalka, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(bear), SacrificeIDs: []uuid.UUID{fodder}, Strict: true,
	}); err == nil {
		t.Fatal("activated without paying {B}")
	}
	effectstest.PayMana(t, g, me, "{B}")
	effectstest.B16Activate(t, g, me.ID, rusalka, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(bear), SacrificeIDs: []uuid.UUID{fodder},
	})
	if g.Battlefield.Contains(fodder) {
		t.Error("the fodder was not sacrificed")
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 1 {
		t.Errorf("the target's power is %d, want 2-1", got)
	}
}

// --- Sanguine Indulgence -----------------------------------------------------

func jmpB2Indulgence() game.Card {
	return game.Card{Name: "Sanguine Indulgence", TypeLine: "Sorcery", ManaCost: "{3}{B}", OracleID: jmpB2IndulgenceOracle}
}

func TestJmpB2SanguineIndulgenceCostsBAfterThreeLifeAndReturnsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.GraveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	b := effectstest.GraveyardCardOf(g, me, "Dead Elf", "Creature — Elf", "{G}", "")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	effectstest.PayMana(t, g, me, "{B}")
	if _, err := jmpB2Cast(t, g, me, jmpB2Indulgence(), game.CastSpellParams{Targets: effectstest.CardRefs(a, b)}); err != nil {
		t.Fatalf("cast for {B} after gaining 3 life: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the two creature cards did not return to hand")
	}
}

func TestJmpB2SanguineIndulgenceIsFullPriceUnderThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.GraveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	effectstest.PayMana(t, g, me, "{B}")
	if _, err := jmpB2Cast(t, g, me, jmpB2Indulgence(), game.CastSpellParams{Targets: effectstest.CardRefs(a)}); err == nil {
		t.Fatal("cast for {B} after gaining only 2 life")
	}
	g.WithWriteLock(func() { me.ManaPool = nil })
	effectstest.PayMana(t, g, me, "{C}{C}{C}{B}")
	if _, err := jmpB2Cast(t, g, me, jmpB2Indulgence(), game.CastSpellParams{Targets: effectstest.CardRefs(a)}); err != nil {
		t.Fatalf("cast at full price: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) {
		t.Error("the creature card did not return to hand")
	}
}

// --- Sanitarium Skeleton -----------------------------------------------------

func TestJmpB2SanitariumSkeletonReturnsFromTheGraveyardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	onBoard := effectstest.B12Push(g, me.ID, "Sanitarium Skeleton", "Creature — Skeleton", jmpB2SkeletonOracle, 1, 2)
	effectstest.PayMana(t, g, me, "{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, onBoard, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated from the battlefield")
	}
	skeleton := effectstest.GraveyardCardOf(g, me, "Sanitarium Skeleton", "Creature — Skeleton", "{B}", jmpB2SkeletonOracle)
	effectstest.B16Activate(t, g, me.ID, skeleton, 0, game.ActivateAbilityParams{})
	if !me.Hand.Contains(skeleton) {
		t.Fatal("the Skeleton did not return to hand")
	}
}

func TestJmpB2SanitariumSkeletonNeedsItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	skeleton := effectstest.GraveyardCardOf(g, me, "Sanitarium Skeleton", "Creature — Skeleton", "{B}", jmpB2SkeletonOracle)
	effectstest.PayMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, skeleton, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for {B}")
	}
	if !effectstest.InGraveyard(me, skeleton) {
		t.Error("the refused activation moved the card")
	}
}

// --- Scourge of Nel Toth -----------------------------------------------------

func TestJmpB2ScourgeOfNelTothIsCastFromTheGraveyardForBBAndTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	if !game.CardCastableFromZone(jmpB2ScourgeOracle, game.ZoneGraveyard) {
		t.Fatal("the graveyard is not open to the Scourge")
	}
	if got := game.AlternativeCostsOfferedFromZone(jmpB2ScourgeOracle, game.ZoneHand); len(got) != 0 {
		t.Fatalf("the graveyard cost is offered from hand: %+v", got)
	}
	offers := game.AlternativeCostsOfferedFromZone(jmpB2ScourgeOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].ManaCost != "{B}{B}" {
		t.Fatalf("graveyard offers = %+v, want {B}{B} and two creatures", offers)
	}
	scourge := effectstest.GraveyardCardOf(g, me, "Scourge of Nel Toth", "Creature — Zombie Dragon", "{5}{B}{B}", jmpB2ScourgeOracle)
	one := effectstest.B12Creature(g, me.ID, "One", "Creature — Goblin", 1, 1)
	two := effectstest.B12Creature(g, me.ID, "Two", "Creature — Goblin", 1, 1)
	params := game.CastSpellParams{FromZone: "graveyard", AlternativeCost: offers[0].Key, Strict: true}

	// One creature is not two, and no mana is not {B}{B}.
	effectstest.PayMana(t, g, me, "{B}{B}")
	params.AltCostIDs = []uuid.UUID{one}
	if err := g.CastSpell(me.ID, scourge, params); err == nil {
		t.Fatal("cast from the graveyard sacrificing one creature")
	}
	g.WithWriteLock(func() { me.ManaPool = nil })
	params.AltCostIDs = []uuid.UUID{one, two}
	if err := g.CastSpell(me.ID, scourge, params); err == nil {
		t.Fatal("cast from the graveyard without {B}{B}")
	}
	if !g.Battlefield.Contains(one) || !g.Battlefield.Contains(two) || !effectstest.InGraveyard(me, scourge) {
		t.Fatal("a refused cast paid part of its cost")
	}

	effectstest.PayMana(t, g, me, "{B}{B}")
	if err := g.CastSpell(me.ID, scourge, params); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	if g.Battlefield.Contains(one) || g.Battlefield.Contains(two) {
		t.Error("the two creatures were not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(scourge) {
		t.Fatal("the Scourge did not enter")
	}
	effectstest.AssertKeywords(t, g, scourge, "flying")
}

func TestJmpB2ScourgeOfNelTothCannotBeCastFromTheGraveyardForItsManaCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	scourge := effectstest.GraveyardCardOf(g, me, "Scourge of Nel Toth", "Creature — Zombie Dragon", "{5}{B}{B}", jmpB2ScourgeOracle)
	effectstest.PayMana(t, g, me, "{C}{C}{C}{C}{C}{B}{B}")
	if err := g.CastSpell(me.ID, scourge, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("cast from the graveyard for its mana cost")
	}
	if !effectstest.InGraveyard(me, scourge) {
		t.Error("the refused cast moved the card")
	}
}

// --- Selhoff Occultist -------------------------------------------------------

func TestJmpB2SelhoffOccultistMillsATargetPlayerForEachDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	occultist := effectstest.B12Push(g, me.ID, "Selhoff Occultist", "Creature — Human Rogue", jmpB2OccultistOracle, 2, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	mill := func(what uuid.UUID) {
		t.Helper()
		lib, gy := opp.Library.Size(), opp.Graveyard.Size()
		effectstest.P1Destroy(t, g, what)
		for i := 0; i < 6 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("pass: %v", err)
			}
		}
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Library.Size() != lib-1 {
			t.Errorf("library %d -> %d, want one card milled", lib, opp.Library.Size())
		}
		// The milled card plus, for an opponent's creature, the creature itself.
		if opp.Graveyard.Size() <= gy {
			t.Errorf("graveyard %d -> %d, want the milled card in it", gy, opp.Graveyard.Size())
		}
	}
	// Another player's creature dying triggers it.
	mill(theirs)
	// So does the Occultist's own death.
	mill(occultist)
}
