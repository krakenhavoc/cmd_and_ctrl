package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// martyrdom_test.go — Martyrdom and ADR 0106 §1's 2026-10-09 amendment
// (#1947): a granted "Only you may activate this ability" is open to the
// player who cast the spell, whoever controls the creature.

const martyrdomOracle = "f6a6da20-52c8-4921-9884-29d3a3051b0d"

// mdSetup casts Martyrdom from the active seat on its own creature and
// returns the table, the caster, the opponent, the granted creature, a
// second creature of the caster's, and a creature of the opponent's.
func mdSetup(t *testing.T) (g *game.Game, me, opp *game.Player, granted, ally, foe uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	granted = pr7Creature(g, me.ID, "Martyr", 2, "W")
	ally = pr7Creature(g, me.ID, "Ally", 2, "W")
	foe = pr7Creature(g, opp.ID, "Pinger", 1, "R")
	castCatalogSpell(t, g, "Martyrdom", "Instant", martyrdomOracle, dgCardRef(granted))
	passPriorityAroundTable(t, g)
	rows := game.ActivatedAbilitiesForCard(*apaLive(g, granted))
	if len(rows) != 1 || !rows[0].GrantorOnly {
		t.Fatalf("the creature's activated rows = %+v, want Martyrdom's one grantor-only row", rows)
	}
	return g, me, opp, granted, ally, foe
}

// The redirection protects the TARGET and the creature that has the
// ability takes the damage: the reverse of Personal Incarnation. One
// damage only.
func TestMartyrdomRedirectsTheNextDamageToTheGrantedCreature(t *testing.T) {
	g, me, _, granted, ally, foe := mdSetup(t)
	pr7Activate(t, g, me.ID, granted, 0, game.ActivateAbilityParams{Targets: dgCardRef(ally)})
	pr6Damage(t, g, foe, ally, 1)
	if got := pr6Marked(g, ally); got != 0 {
		t.Errorf("the protected creature has %d damage, want 0", got)
	}
	if got := pr6Marked(g, granted); got != 1 {
		t.Errorf("the granted creature has %d damage, want 1", got)
	}
	pr6Damage(t, g, foe, ally, 1)
	if got := pr6Marked(g, ally); got != 1 {
		t.Errorf("the second damage: protected creature has %d, want 1 (the charge was one)", got)
	}
	if got := pr6Marked(g, granted); got != 1 {
		t.Errorf("the second damage reached the granted creature: %d, want 1", got)
	}
}

// A player is a legal thing to protect.
func TestMartyrdomCanProtectAPlayer(t *testing.T) {
	g, me, _, granted, _, foe := mdSetup(t)
	pr7Activate(t, g, me.ID, granted, 0, game.ActivateAbilityParams{Targets: pr6Player(me.ID)})
	life := me.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(foe, me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if me.Life != life {
		t.Errorf("life %d, want %d: the damage was not redirected", me.Life, life)
	}
	if got := pr6Marked(g, granted); got != 1 {
		t.Errorf("the granted creature has %d damage, want 1", got)
	}
}

// "Only you": the opponent who does not control the creature cannot use
// the row either.
func TestMartyrdomAnOpponentMayNotActivateIt(t *testing.T) {
	g, _, opp, granted, ally, _ := mdSetup(t)
	err := g.ActivateCatalogAbility(opp.ID, granted, 0, game.ActivateAbilityParams{Targets: dgCardRef(ally), Strict: true})
	if !errors.Is(err, game.ErrCardCallerMismatch) {
		t.Fatalf("an opponent activating the caster's creature's granted row: err = %v, want ErrCardCallerMismatch", err)
	}
}

// The creature changes hands after Martyrdom resolved. Its new controller
// has the creature but not the ability; the caster has the ability and
// not the creature, and the damage still goes to the creature.
func TestMartyrdomSurvivesAControlChange(t *testing.T) {
	g, me, opp, granted, ally, foe := mdSetup(t)
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), granted, opp.ID, game.IndefiniteDuration(), "test — steal") {
			t.Fatal("setup: control change refused")
		}
	})
	if c := apaLive(g, granted); c.Controller != opp.ID {
		t.Fatalf("setup: controller %v, want the opponent", c.Controller)
	}

	err := g.ActivateCatalogAbility(opp.ID, granted, 0, game.ActivateAbilityParams{Targets: pr6Player(opp.ID), Strict: true})
	if !errors.Is(err, game.ErrCardCallerMismatch) {
		t.Fatalf("the new controller activating the grantor-only row: err = %v, want ErrCardCallerMismatch", err)
	}

	pr7Activate(t, g, me.ID, granted, 0, game.ActivateAbilityParams{Targets: dgCardRef(ally)})
	pr6Damage(t, g, foe, ally, 1)
	if got := pr6Marked(g, ally); got != 0 {
		t.Errorf("the protected creature has %d damage, want 0", got)
	}
	if got := pr6Marked(g, granted); got != 1 {
		t.Errorf("the stolen creature has %d damage, want 1", got)
	}
}

// A grantor-only row is only ever a granted row's.
func TestGrantorOnlyIsRefusedOnACardsOwnRow(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a card's own grantor-only row was accepted")
		}
	}()
	Register(Spec{
		OracleID:  "test-own-grantor-only-row",
		Name:      "Own Grantor-Only Row",
		Activated: []ActivatedAbility{{Label: "x", Cost: ManaCost("{0}"), GrantorOnly: true}},
	})
}

func TestGrantorOnlyRegistrationGuards(t *testing.T) {
	for name, row := range map[string]ActivatedAbility{
		"grantor and opponents": {Label: "x", Cost: ManaCost("{1}"), GrantorOnly: true, OpponentsOnly: true},
		"grantor and owner":     {Label: "x", Cost: ManaCost("{1}"), GrantorOnly: true, OwnerOnly: true},
		"grantor and any":       {Label: "x", Cost: ManaCost("{1}"), GrantorOnly: true, AnyPlayer: true},
		"grantor with tap":      {Label: "x", Cost: TapCost(), GrantorOnly: true},
		"grantor purpose":       {Label: "x", Cost: ManaCost("{1}"), GrantorOnly: true, Purpose: game.Purpose{Draws: 1}},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: Register accepted it", name)
				}
			}()
			checkAnyPlayerAbility("Test", "ability 0", row)
		}()
	}
}
