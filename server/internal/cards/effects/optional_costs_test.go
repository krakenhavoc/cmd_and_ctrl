package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// optional_costs_test.go — ADR 0073 (#664), the card half.
//
// Every assertion here fails with the engine change backed out, and
// each one fails for a different reason, which is the point of having
// six cards rather than one:
//
//   - Burst Lightning: the resolution reads the paid record.
//   - Rite of Replication: the same read feeding a COUNT.
//   - Wolfbriar Elemental: a trigger on the permanent reads the count
//     AFTER entry, which is the half the stack item cannot answer.
//   - Gatekeeper of Malakir: the same read as a CR 603.4 intervening
//     if, so an unkicked one triggers nothing at all.
//   - Capsize: buyback's mana half, and the graveyard when declined.
//   - Constant Mists: buyback's NON-mana half, paid at CR 601.2h.

const (
	burstLightningOracle  = "ac2086fe-98ee-4280-9c7c-c5c2d6548a8b"
	riteOfReplicationOra  = "fb60739e-1dc3-481d-a056-ad72e665c680"
	wolfbriarOracle       = "2f8872fe-84dc-4cda-a253-e7503a5c96a3"
	gatekeeperOracle      = "6781f8ae-2a86-4e3d-bc43-48809c9d6c26"
	capsizeOracle         = "77637eff-2963-4402-88f3-ca346f762fc8"
	constantMistsOracle   = "c850a29e-dc40-4ab6-89f1-d501a1a350d1"
	urzasRuinousOracle    = "978e0d87-3ff2-4a73-916c-ff0dc0ab2797"
	rakdosLordOracle      = "143a269a-b9ee-48ba-bd7b-4aa46eb36778"
	grafdiggersCageOracle = "753cb2b2-24ce-484f-a2d0-be6fd2c67ebd"
	ruleOfLawOracle       = "53e88e64-6f82-4154-a66e-6aeb0154b368"
)

// castWithOptionalCosts is castCatalogSpell plus the CR 601.2b
// announcement: the optional costs claimed, and the permanents paid
// to the non-mana half of one.
func castWithOptionalCosts(t *testing.T, g *game.Game, name, typeLine, oracleID string,
	targets []game.TargetRef, optional []int, sacrifices []uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{
		Targets:       targets,
		OptionalCosts: optional,
		SacrificeIDs:  sacrifices,
	})
}

func TestBurstLightningKickedAndUnkicked(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		want     int
	}{
		{"unkicked", nil, 2},
		{"kicked", []int{0}, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			victim := g.Seats[1]
			before := victim.Life

			if _, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant",
				burstLightningOracle,
				[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
				tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)

			if got := before - victim.Life; got != tc.want {
				t.Errorf("damage dealt: got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestRiteOfReplicationKickedMakesFiveCopies(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		want     int
	}{
		{"unkicked", nil, 1},
		{"kicked", []int{0}, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			bear := seedCreature(g, "Grizzly Bears", caster.ID)

			if _, err := castWithOptionalCosts(t, g, "Rite of Replication", "Sorcery",
				riteOfReplicationOra,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bear}},
				tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)

			tokens := 0
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Grizzly Bears" && c.InstanceID != bear {
					tokens++
				}
			}
			if tokens != tc.want {
				t.Errorf("token copies: got %d, want %d", tokens, tc.want)
			}
		})
	}
}

// TestWolfbriarElementalMultikickerCount is the multikicker assertion
// AND the "the ETB reads the right value after entry" one: the count
// is read by a trigger on the permanent, off Card.PaidOptionalCosts,
// long after the stack item is gone.
func TestWolfbriarElementalMultikickerCount(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		want     int
	}{
		{"unkicked", nil, 0},
		{"kicked once", []int{0}, 1},
		{"kicked three times", []int{0, 0, 0}, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)

			if _, err := castWithOptionalCosts(t, g, "Wolfbriar Elemental",
				"Creature — Elemental", wolfbriarOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)

			wolves := 0
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Wolf" {
					wolves++
				}
			}
			if wolves != tc.want {
				t.Errorf("Wolves created: got %d, want %d", wolves, tc.want)
			}
		})
	}
}

// TestGatekeeperOfMalakirKickedIsAnInterveningIf pins CR 603.4: the
// unkicked Gatekeeper does not put an ability on the stack at all,
// which is observably different from one that triggers and does
// nothing.
func TestGatekeeperOfMalakirKickedIsAnInterveningIf(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]

	if _, err := castWithOptionalCosts(t, g, "Gatekeeper of Malakir",
		"Creature — Vampire Warrior", gatekeeperOracle, nil, nil, nil); err != nil {
		t.Fatalf("unkicked CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(g.PendingTriggers); n != 0 {
		t.Fatalf("unkicked Gatekeeper queued %d triggers, want 0", n)
	}

	// Kicked: the trigger exists, targets a player, and that player
	// is asked to sacrifice.
	victim := g.Seats[1]
	prey := seedCreature(g, "Doomed Bear", victim.ID)
	_ = caster
	if _, err := castWithOptionalCosts(t, g, "Gatekeeper of Malakir",
		"Creature — Vampire Warrior", gatekeeperOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
		[]int{0}, nil); err != nil {
		t.Fatalf("kicked CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)

	// The victim's only creature is either gone (auto-picked) or a
	// prompt is open naming it. Either way the ability happened,
	// which the unkicked case above proves it otherwise does not.
	if len(g.PendingChoices) == 0 && findBattlefieldCardByID(g, prey) != nil {
		t.Errorf("kicked Gatekeeper neither sacrificed the creature nor opened a prompt")
	}
}

// TestCapsizeBuybackReturnsToHand is buyback's mana half, both ways
// round in one test because the difference IS the assertion.
func TestCapsizeBuybackReturnsToHand(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		wantHand bool
	}{
		{"not bought back goes to the graveyard", nil, false},
		{"bought back returns to hand", []int{0}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			bear := seedCreature(g, "Grizzly Bears", caster.ID)

			id, err := castWithOptionalCosts(t, g, "Capsize", "Instant", capsizeOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)

			inHand := caster.Hand.Contains(id)
			inGrave := caster.Graveyard.Contains(id)
			if inHand != tc.wantHand || inGrave == tc.wantHand {
				t.Errorf("Capsize after resolution: hand=%v graveyard=%v, want hand=%v",
					inHand, inGrave, tc.wantHand)
			}
		})
	}
}

// TestConstantMistsBuybackSacrificesALand is the NON-MANA buyback:
// the land leaves at CR 601.2h, with the spell already on the stack,
// and the spell comes back to hand.
func TestConstantMistsBuybackSacrificesALand(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	land := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Forest",
		TypeLine:   "Basic Land — Forest",
		Owner:      caster.ID,
		Controller: caster.ID,
	})

	id, err := castWithOptionalCosts(t, g, "Constant Mists", "Instant",
		constantMistsOracle, nil, []int{0}, []uuid.UUID{land})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// CR 601.2a before 601.2h: the land is already gone while the
	// spell is still on the stack.
	if findBattlefieldCardByID(g, land) != nil {
		t.Errorf("the sacrificed land is still on the battlefield after announce")
	}
	if !caster.Graveyard.Contains(land) {
		t.Errorf("the sacrificed land did not reach the graveyard")
	}
	passPriorityAroundTable(t, g)

	if !caster.Hand.Contains(id) {
		t.Errorf("a bought-back Constant Mists did not return to its owner's hand")
	}
}

// TestConstantMistsRefusesAnUnpayableBuyback is the announce-time
// half: a caster with no land cannot claim the offer, and the refusal
// leaves the card in hand.
func TestConstantMistsRefusesAnUnpayableBuyback(t *testing.T) {
	g := newCatalogGame(t)

	id, err := castWithOptionalCosts(t, g, "Constant Mists", "Instant",
		constantMistsOracle, nil, []int{0}, nil)
	if err == nil {
		t.Fatalf("claiming buyback with no land named was accepted")
	}
	active := g.Seats[g.Turn.ActiveSeat]
	if !active.Hand.Contains(id) {
		t.Errorf("a refused cast did not leave the card in hand")
	}
}

// findBattlefieldCardByID is a local read so these tests do not
// depend on the engine's own unexported lookup.
func findBattlefieldCardByID(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}
