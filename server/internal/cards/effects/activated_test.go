package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activated_test.go — S21 sub-PR 2: catalog activated abilities.

const (
	goblinBombardmentOracle = "edad60c6-80de-4033-af1b-a703ac332983"
	carrionFeederOracle     = "a1cc5e37-b09a-4b7f-afd5-77c1c35aa425"
	krenkoOracle            = "68418069-f615-40ef-ae0d-764192acae00"
)

// pushCatalogPermanent seeds a catalog card onto the battlefield,
// already free of summoning sickness unless `sick` says otherwise.
func pushCatalogPermanent(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, sick bool) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: 1, Toughness: 1, Owner: owner, Controller: owner,
		SummonedThisTurn: sick,
	})
	return id
}

func TestActivatedAbilitiesAreWired(t *testing.T) {
	for _, oracle := range []string{goblinBombardmentOracle, carrionFeederOracle, krenkoOracle} {
		if len(game.ActivatedAbilitiesForCard(game.Card{OracleID: oracle})) != 1 {
			t.Errorf("%s: expected exactly one activated ability", oracle)
		}
	}
	bomb := game.ActivatedAbilitiesForCard(game.Card{OracleID: goblinBombardmentOracle})[0]
	if bomb.Cost.SacrificeOther == nil || bomb.Targets == nil {
		t.Errorf("Bombardment should cost a sacrifice and take a target: %+v", bomb.Cost)
	}
	if bomb.Cost.Tap || bomb.Cost.Mana != "" {
		t.Errorf("Bombardment's cost is the creature alone: %+v", bomb.Cost)
	}
	kr := game.ActivatedAbilitiesForCard(game.Card{OracleID: krenkoOracle})[0]
	if !kr.Cost.Tap || kr.Targets != nil {
		t.Errorf("Krenko should be a tap ability with no target: %+v", kr)
	}
}

// --- Goblin Bombardment ------------------------------------------

func TestGoblinBombardmentSacrificesAndDealsDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bomb := pushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", goblinBombardmentOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
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
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("life %d -> %d, want -1", before, opp.Life)
	}
	if g.Battlefield.Contains(bomb) == false {
		t.Errorf("Bombardment itself is not sacrificed")
	}
}

func TestGoblinBombardmentRejectsBadCosts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bomb := pushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", goblinBombardmentOracle, false)
	mine := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Theirs", "Creature — Bear", "", false)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bomb := pushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", goblinBombardmentOracle, false)
	// Doomed Traveler dies → Spirit token. Sacrificing it to the
	// Bombardment should make the Spirit before the damage lands.
	traveler := pushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier",
		"a30907c0-fbde-4fd3-a8c7-f304305fcea7", false)

	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{traveler},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if triggerOnStack(g, traveler) == nil {
		t.Fatalf("the dies-trigger from the cost should be on the stack")
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Spirit") == uuid.Nil {
		t.Errorf("Doomed Traveler's Spirit should exist")
	}
}

// --- Carrion Feeder ----------------------------------------------

func TestCarrionFeederGrowsAndCanEatItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	feeder := pushCatalogPermanent(g, me.ID, "Carrion Feeder", "Creature — Zombie", carrionFeederOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)

	if err := g.ActivateCatalogAbility(me.ID, feeder, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	passPriorityAroundTable(t, g)
}

// --- Krenko ------------------------------------------------------

func TestKrenkoCountsGoblinsAtResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	krenko := pushCatalogPermanent(g, me.ID, "Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", krenkoOracle, false)
	pushCatalogPermanent(g, me.ID, "Goblin Piker", "Creature — Goblin", "", false)
	pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)

	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	sick := pushCatalogPermanent(g, me.ID, "Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", krenkoOracle, true)
	if err := g.ActivateCatalogAbility(me.ID, sick, 0, game.ActivateAbilityParams{}); err != game.ErrSummoningSick {
		t.Fatalf("freshly-played Krenko: got %v, want ErrSummoningSick", err)
	}
	ready := pushCatalogPermanent(g, me.ID, "Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", krenkoOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, ready, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("settled Krenko: %v", err)
	}
	// Now tapped — a second activation fails.
	if err := g.ActivateCatalogAbility(me.ID, ready, 0, game.ActivateAbilityParams{}); err != game.ErrAlreadyTapped {
		t.Errorf("tapped Krenko: got %v, want ErrAlreadyTapped", err)
	}
}

func TestActivateRejectsOpponentsPermanentAndBadIndex(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushCatalogPermanent(g, opp.ID, "Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", krenkoOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, theirs, 0, game.ActivateAbilityParams{}); err != game.ErrCardCallerMismatch {
		t.Errorf("opponent's permanent: got %v", err)
	}
	mine := pushCatalogPermanent(g, me.ID, "Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", krenkoOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, mine, 3, game.ActivateAbilityParams{}); err != game.ErrInvalidParam {
		t.Errorf("out-of-range index: got %v", err)
	}
}
