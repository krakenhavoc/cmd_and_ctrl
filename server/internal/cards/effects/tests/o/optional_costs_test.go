package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
			g := effectstest.NewCatalogGame(t)
			victim := g.Seats[1]
			before := victim.Life

			if _, err := effectstest.CastWithOptionalCosts(t, g, "Burst Lightning", "Instant",
				effectstest.BurstLightningOracle,
				[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
				tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)

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
			g := effectstest.NewCatalogGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			bear := effectstest.SeedCreature(g, "Grizzly Bears", caster.ID)

			if _, err := effectstest.CastWithOptionalCosts(t, g, "Rite of Replication", "Sorcery",
				effectstest.RiteOfReplicationOra,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bear}},
				tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)

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
// is read by a trigger on the permanent, off Card.Provenance,
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
			g := effectstest.NewCatalogGame(t)

			if _, err := effectstest.CastWithOptionalCosts(t, g, "Wolfbriar Elemental",
				"Creature — Elemental", effectstest.WolfbriarOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]

	if _, err := effectstest.CastWithOptionalCosts(t, g, "Gatekeeper of Malakir",
		"Creature — Vampire Warrior", effectstest.GatekeeperOracle, nil, nil, nil); err != nil {
		t.Fatalf("unkicked CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(g.PendingTriggers); n != 0 {
		t.Fatalf("unkicked Gatekeeper queued %d triggers, want 0", n)
	}

	// Kicked: the trigger exists, targets a player, and that player
	// is asked to sacrifice.
	victim := g.Seats[1]
	prey := effectstest.SeedCreature(g, "Doomed Bear", victim.ID)
	_ = caster
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Gatekeeper of Malakir",
		"Creature — Vampire Warrior", effectstest.GatekeeperOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
		[]int{0}, nil); err != nil {
		t.Fatalf("kicked CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// The victim's only creature is either gone (auto-picked) or a
	// prompt is open naming it. Either way the ability happened,
	// which the unkicked case above proves it otherwise does not.
	if len(g.PendingChoices) == 0 && effectstest.FindBattlefieldCardByID(g, prey) != nil {
		t.Errorf("kicked Gatekeeper neither sacrificed the creature nor opened a prompt")
	}
}

// TestCapsizeBuybackReturnsToHand is buyback's mana half, both ways
// round in one test because the difference IS the assertion.
func TestCapsizeBuybackReturnsToHand(t *testing.T) {
	for _, tc := range []struct {
		Name     string
		Optional []int
		WantHand bool
	}{
		{"not bought back goes to the graveyard", nil, false},
		{"bought back returns to hand", []int{0}, true},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			bear := effectstest.SeedCreature(g, "Grizzly Bears", caster.ID)

			id, err := effectstest.CastWithOptionalCosts(t, g, "Capsize", "Instant", effectstest.CapsizeOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}, tc.Optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)

			inHand := caster.Hand.Contains(id)
			inGrave := caster.Graveyard.Contains(id)
			if inHand != tc.WantHand || inGrave == tc.WantHand {
				t.Errorf("Capsize after resolution: hand=%v graveyard=%v, want hand=%v",
					inHand, inGrave, tc.WantHand)
			}
		})
	}
}

// TestConstantMistsBuybackSacrificesALand is the NON-MANA buyback:
// the land leaves at CR 601.2h, with the spell already on the stack,
// and the spell comes back to hand.
func TestConstantMistsBuybackSacrificesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Forest",
		TypeLine:   "Basic Land — Forest",
		Owner:      caster.ID,
		Controller: caster.ID,
	})

	id, err := effectstest.CastWithOptionalCosts(t, g, "Constant Mists", "Instant",
		effectstest.ConstantMistsOracle, nil, []int{0}, []uuid.UUID{land})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// CR 601.2a before 601.2h: the land is already gone while the
	// spell is still on the stack.
	if effectstest.FindBattlefieldCardByID(g, land) != nil {
		t.Errorf("the sacrificed land is still on the battlefield after announce")
	}
	if !caster.Graveyard.Contains(land) {
		t.Errorf("the sacrificed land did not reach the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !caster.Hand.Contains(id) {
		t.Errorf("a bought-back Constant Mists did not return to its owner's hand")
	}
}

// TestConstantMistsRefusesAnUnpayableBuyback is the announce-time
// half: a caster with no land cannot claim the offer, and the refusal
// leaves the card in hand.
func TestConstantMistsRefusesAnUnpayableBuyback(t *testing.T) {
	g := effectstest.NewCatalogGame(t)

	id, err := effectstest.CastWithOptionalCosts(t, g, "Constant Mists", "Instant",
		effectstest.ConstantMistsOracle, nil, []int{0}, nil)
	if err == nil {
		t.Fatalf("claiming buyback with no land named was accepted")
	}
	active := g.Seats[g.Turn.ActiveSeat]
	if !active.Hand.Contains(id) {
		t.Errorf("a refused cast did not leave the card in hand")
	}
}
