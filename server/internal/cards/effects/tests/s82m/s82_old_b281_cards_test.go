package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b281_cards_test.go — Eternal Masters, batch 281 (#5542). Every test
// plays the real catalog entry through the engine.

const (
	b281AutocratOracle   = "0561043a-3896-4bb1-8ff5-ee674ddfea2a"
	b281FanaticOracle    = "7efd5d62-6da2-428a-98e3-5b842f668656"
	b281InvigorateOracle = "9264c198-76c7-4ac0-9334-37c691de1b54"
	b281GuardOracle      = "c95cebf7-c28c-4453-b5e4-e895d621f290"
	b281VoidOracle       = "fdb28fd3-70fb-43c3-b367-9cf0d2c46e06"
)

func TestB281EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		b281AutocratOracle: "Sengir Autocrat", b281FanaticOracle: "Mogg Fanatic", b281InvigorateOracle: "Invigorate",
		b281GuardOracle: "Coalition Honor Guard", b281VoidOracle: "Void",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func b281Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func b281Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle, pool string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, p, pool)
	id := effectstest.HandCardOf(g, p, name, typeLine, cost, oracle)
	params.Strict = true
	return id, g.CastSpell(p.ID, id, params)
}

// --- Sengir Autocrat ---

func TestB281AutocratMakesThreeSerfsAndExilesEverySerfWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	_, err := b281Cast(t, g, me, "Sengir Autocrat", "Creature — Human", "{3}{B}", b281AutocratOracle, "CCCB", game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	serfs := effectstest.TokensNamed(g, "Serf")
	if len(serfs) != 3 {
		t.Fatalf("%d Serf tokens, want 3", len(serfs))
	}
	if p, tough := effectstest.PtOf(t, g, serfs[0].InstanceID); p != 0 || tough != 1 {
		t.Errorf("a Serf is %d/%d, want 0/1", p, tough)
	}
	other := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Serf", TypeLine: "Token Creature — Serf",
		Toughness: 1, Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID})
	keep := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	autocrat := effectstest.BattlefieldIDNamed(g, me.ID, "Sengir Autocrat")
	effectstest.KillCreature(t, g, me.ID, autocrat)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.TokensNamed(g, "Serf")); got != 0 {
		t.Errorf("%d Serf tokens left, want 0 (including an opponent's)", got)
	}
	if g.Battlefield.Contains(other) {
		t.Error("the opponent's Serf survived")
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("a non-Serf was exiled")
	}
}

// --- Mogg Fanatic ---

func TestB281MoggFanaticSacrificesItselfForOneDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fan := effectstest.B12Push(g, me.ID, "Mogg Fanatic", "Creature — Goblin", b281FanaticOracle, 1, 1)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, fan, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fan) {
		t.Error("the Fanatic was not sacrificed")
	}
	if opp.Life != life-1 {
		t.Errorf("life %d -> %d, want 1 damage", life, opp.Life)
	}
}

// --- Invigorate ---

func TestB281InvigorateForFreeNeedsAForestAndGivesAnOpponentThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	params := game.CastSpellParams{AlternativeCost: "invigorate", AltCostOpponent: opp.ID, Targets: b281Target(bear)}
	if _, err := b281Cast(t, g, me, "Invigorate", "Instant", "{2}{G}", b281InvigorateOracle, "", params); err == nil {
		t.Fatal("the free cast was accepted with no Forest")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	life := opp.Life
	if _, err := b281Cast(t, g, me, "Invigorate", "Instant", "{2}{G}", b281InvigorateOracle, "", params); err != nil {
		t.Fatalf("free cast with a Forest: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life+3 {
		t.Errorf("the named opponent's life %d -> %d, want +3", life, opp.Life)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("the target is %d/%d, want 6/6", p, tough)
	}
}

func TestB281InvigoratePaidWithManaNeedsNoForestAndGivesNoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	if _, err := b281Cast(t, g, me, "Invigorate", "Instant", "{2}{G}", b281InvigorateOracle, "CCG", game.CastSpellParams{Targets: b281Target(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("opponent life changed to %d", opp.Life)
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 6 {
		t.Errorf("power %d, want 6", p)
	}
}

// --- Coalition Honor Guard ---

func TestB281HonorGuardForcesAnOpponentsTargetsOntoItButNotYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	guard := effectstest.B12Push(g, opp.ID, "Coalition Honor Guard", "Creature — Human Flagbearer", b281GuardOracle, 2, 4)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	if _, err := b281Cast(t, g, me, "Slaying Fire", "Instant", "{2}{R}", effectstest.SlayingFireOracle, "CCR",
		game.CastSpellParams{Targets: b281Target(other)}); err == nil {
		t.Fatal("an opponent's spell targeted another creature while a Flagbearer was available")
	}
	if _, err := b281Cast(t, g, me, "Slaying Fire", "Instant", "{2}{R}", effectstest.SlayingFireOracle, "CCR",
		game.CastSpellParams{Targets: b281Target(guard)}); err != nil {
		t.Fatalf("targeting the Flagbearer: %v", err)
	}
}

func TestB281HonorGuardDoesNotBindItsOwnController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Coalition Honor Guard", "Creature — Human Flagbearer", b281GuardOracle, 2, 4)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	if _, err := b281Cast(t, g, me, "Slaying Fire", "Instant", "{2}{R}", effectstest.SlayingFireOracle, "CCR",
		game.CastSpellParams{Targets: b281Target(other)}); err != nil {
		t.Fatalf("the controller's own spell was bound: %v", err)
	}
}

// --- Void ---

func b281Void(t *testing.T, number int) (g *game.Game, me, opp *game.Player, killed, spared, atNumber, wrongCost, land uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mk := func(name, typ, cost string, owner *game.Player) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typ, ManaCost: cost,
			Power: 1, Toughness: 1, Owner: owner.ID, Controller: owner.ID})
	}
	killed = mk("Two Drop", "Creature — Bear", "{1}{G}", opp)
	spared = mk("Three Drop", "Creature — Bear", "{2}{G}", opp)
	atNumber = mk("Mine Two", "Artifact", "{2}", me)
	_ = atNumber
	opp.Hand.Cards = nil
	wrongCost = effectstest.HandCardOf(g, opp, "Hand Three", "Sorcery", "{2}{R}", "")
	atNumber = effectstest.HandCardOf(g, opp, "Hand Two", "Sorcery", "{1}{R}", "")
	land = effectstest.HandCardOf(g, opp, "Island", "Basic Land — Island", "", "")
	if _, err := b281Cast(t, g, me, "Void", "Sorcery", "{3}{B}{R}", b281VoidOracle, "CCCBR",
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast Void: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, me.ID, number)
	effectstest.PassPriorityAroundTable(t, g)
	return
}

func TestB281VoidDestroysArtifactsAndCreaturesOfTheNumberAndDiscardsNonlandCardsOfIt(t *testing.T) {
	g, _, opp, killed, spared, handTwo, handThree, land := b281Void(t, 2)
	if g.Battlefield.Contains(killed) {
		t.Error("the mana value 2 creature survived")
	}
	if !g.Battlefield.Contains(spared) {
		t.Error("the mana value 3 creature was destroyed")
	}
	if !opp.Graveyard.Contains(handTwo) {
		t.Error("the nonland mana value 2 card was not discarded")
	}
	if !opp.Hand.Contains(handThree) || !opp.Hand.Contains(land) {
		t.Error("a card that should have stayed in hand was discarded")
	}
}

func TestB281VoidForZeroDestroysTokensAndSparesTheRest(t *testing.T) {
	g, _, opp, killed, spared, _, _, land := b281Void(t, 0)
	if !g.Battlefield.Contains(killed) || !g.Battlefield.Contains(spared) {
		t.Error("a creature with a mana value above 0 was destroyed")
	}
	if !opp.Hand.Contains(land) {
		t.Error("a land with mana value 0 was discarded: only nonland cards go")
	}
}
