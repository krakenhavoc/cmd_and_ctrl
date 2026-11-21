package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestActivatedAbilitiesAreWired(t *testing.T) {
	for _, oracle := range []string{effectstest.GoblinBombardmentOracle, effectstest.CarrionFeederOracle, effectstest.KrenkoOracle} {
		if len(game.ActivatedAbilitiesForCard(game.Card{OracleID: oracle})) != 1 {
			t.Errorf("%s: expected exactly one activated ability", oracle)
		}
	}
	bomb := game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.GoblinBombardmentOracle})[0]
	if bomb.Cost.SacrificeOther == nil || bomb.Targets == nil {
		t.Errorf("Bombardment should cost a sacrifice and take a target: %+v", bomb.Cost)
	}
	if bomb.Cost.Tap || bomb.Cost.Mana != "" {
		t.Errorf("Bombardment's cost is the creature alone: %+v", bomb.Cost)
	}
	kr := game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.KrenkoOracle})[0]
	if !kr.Cost.Tap || kr.Targets != nil {
		t.Errorf("Krenko should be a tap ability with no target: %+v", kr)
	}
}

// --- Goblin Bombardment ------------------------------------------

func TestGoblinBombardmentSacrificesAndDealsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bomb := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", effectstest.GoblinBombardmentOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	before := opp.Life

	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// The cost is paid on announce — the creature is already gone
	// while the ability waits on the stack.
	if g.Battlefield.Contains(fodder) {
		t.Errorf("sacrifice cost should be paid at announce, not resolution")
	}
	if opp.Life != before {
		t.Errorf("damage should wait for resolution")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("life %d -> %d, want -1", before, opp.Life)
	}
	if g.Battlefield.Contains(bomb) == false {
		t.Errorf("Bombardment itself is not sacrificed")
	}
}

func TestGoblinBombardmentRejectsBadCosts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bomb := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", effectstest.GoblinBombardmentOracle, false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Theirs", "Creature — Bear", "", false)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}

	cases := []struct {
		name string
		sac  []uuid.UUID
		want error
	}{
		{"no sacrifice named", nil, game.ErrInvalidParam},
		{"opponent's creature", []uuid.UUID{theirs}, game.ErrCardCallerMismatch},
		{"a non-creature", []uuid.UUID{rock}, game.ErrIllegalTarget},
		{"two creatures for a one-creature cost", []uuid.UUID{mine, mine}, game.ErrInvalidParam},
	}
	for _, tc := range cases {
		err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
			SacrificeIDs: tc.sac, Targets: target,
		})
		if err != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, err, tc.want)
		}
	}
	// Nothing was paid on any of the rejected attempts.
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(rock) || !g.Battlefield.Contains(theirs) {
		t.Errorf("a rejected activation must not pay any part of its cost")
	}
	// A missing target is rejected too.
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{mine},
	}); err != game.ErrInvalidParam {
		t.Errorf("no target: got %v, want ErrInvalidParam", err)
	}
	if !g.Battlefield.Contains(mine) {
		t.Errorf("target validation must run before the cost is paid")
	}
}

// The dies-trigger caused by paying the cost goes on the stack
// ABOVE the ability, so it resolves first (CR 601.2h / 603.3b).
func TestSacrificeCostTriggerResolvesBeforeTheAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bomb := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", effectstest.GoblinBombardmentOracle, false)
	// Doomed Traveler dies → Spirit token. Sacrificing it to the
	// Bombardment should make the Spirit before the damage lands.
	traveler := effectstest.PushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier",
		"a30907c0-fbde-4fd3-a8c7-f304305fcea7", false)

	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{traveler},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.TriggerOnStack(g, traveler) == nil {
		t.Fatalf("the dies-trigger from the cost should be on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Spirit") == uuid.Nil {
		t.Errorf("Doomed Traveler's Spirit should exist")
	}
}

// --- Carrion Feeder ----------------------------------------------

func TestCarrionFeederGrowsAndCanEatItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	feeder := effectstest.PushCatalogPermanent(g, me.ID, "Carrion Feeder", "Creature — Zombie", effectstest.CarrionFeederOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)

	if err := g.ActivateCatalogAbility(me.ID, feeder, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var counters int
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == feeder {
			counters = g.Battlefield.Cards[i].Counters["+1/+1"]
		}
	}
	if counters != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", counters)
	}
	// Eating itself is legal and self-defeating: the ability
	// resolves with its source in the graveyard, so nothing gets a
	// counter and nothing errors.
	if err := g.ActivateCatalogAbility(me.ID, feeder, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{feeder},
	}); err != nil {
		t.Fatalf("self-sacrifice should be legal: %v", err)
	}
	if g.Battlefield.Contains(feeder) {
		t.Errorf("the Feeder ate itself and should be gone")
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// #3667: an earlier activation still on the stack when the Feeder eats
// itself names a permanent that is gone (CR 400.7, CR 608.2b): no counter
// lands on the graveyard card.
func TestCarrionFeederSacrificedInResponseGetsNoGraveyardCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	feeder := effectstest.PushCatalogPermanent(g, me.ID, "Carrion Feeder", "Creature — Zombie", effectstest.CarrionFeederOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)

	if err := g.ActivateCatalogAbility(me.ID, feeder, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, feeder, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{feeder},
	}); err != nil {
		t.Fatalf("self-sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	found := false
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID != feeder {
			continue
		}
		found = true
		if n := c.Counters["+1/+1"]; n != 0 {
			t.Errorf("graveyard Feeder has %d +1/+1 counters, want 0", n)
		}
	}
	if !found {
		t.Fatalf("Feeder should be in the graveyard")
	}
}

// --- Krenko ------------------------------------------------------

func TestKrenkoCountsGoblinsAtResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	krenko := effectstest.PushCatalogPermanent(g, me.ID, "Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", effectstest.KrenkoOracle, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Goblin Piker", "Creature — Goblin", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)

	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Krenko + Piker = 2 Goblins → 2 tokens (the Bear doesn't count).
	var tokens int
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Goblin" {
			tokens++
		}
	}
	if tokens != 2 {
		t.Errorf("created %d Goblins, want 2", tokens)
	}
}

func TestKrenkoTapCostRespectsSummoningSicknessAndTapState(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sick := effectstest.PushCatalogPermanent(g, me.ID, "Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", effectstest.KrenkoOracle, true)
	if err := g.ActivateCatalogAbility(me.ID, sick, 0, game.ActivateAbilityParams{}); err != game.ErrSummoningSick {
		t.Fatalf("freshly-played Krenko: got %v, want ErrSummoningSick", err)
	}
	ready := effectstest.PushCatalogPermanent(g, me.ID, "Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", effectstest.KrenkoOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, ready, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("settled Krenko: %v", err)
	}
	// Now tapped — a second activation fails.
	if err := g.ActivateCatalogAbility(me.ID, ready, 0, game.ActivateAbilityParams{}); err != game.ErrAlreadyTapped {
		t.Errorf("tapped Krenko: got %v, want ErrAlreadyTapped", err)
	}
}

func TestActivateRejectsOpponentsPermanentAndBadIndex(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", effectstest.KrenkoOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, theirs, 0, game.ActivateAbilityParams{}); err != game.ErrCardCallerMismatch {
		t.Errorf("opponent's permanent: got %v", err)
	}
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", effectstest.KrenkoOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, mine, 3, game.ActivateAbilityParams{}); err != game.ErrInvalidParam {
		t.Errorf("out-of-range index: got %v", err)
	}
}
