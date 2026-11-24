package testss82v

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// seeker_of_slaanesh_test.go — Allure of Slaanesh: "Each opponent must
// attack with at least one creature each combat if able" (#6087), a
// player-level CR 508.1d requirement obeyed by any one attacker.

const seekerOfSlaaneshOracle = "47af21d3-b43f-4101-81cf-ca44db1a2706"

func seekerHasPass(g *game.Game, seat *game.Player) bool {
	for _, m := range legal.EnumerateFor(g, seat.ID) {
		if m.Kind == legal.KindPass {
			return true
		}
	}
	return false
}

// TestSeekerOfSlaaneshMakesAnOpponentAttackWithOneCreature: an
// opponent's empty declaration is refused naming the Seeker, and one
// attacker, at any player, answers it while the other creature stays
// home.
func TestSeekerOfSlaaneshMakesAnOpponentAttackWithOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, seekerSide, other := g.Seats[0], g.Seats[1], g.Seats[2]
	seeker := effectstest.B12Push(g, seekerSide.ID, "Seeker of Slaanesh", "Creature — Demon", seekerOfSlaaneshOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)

	var re *game.AttackRequirementError
	err := g.PassPriority()
	if !errors.As(err, &re) {
		t.Fatalf("empty declaration under the Seeker = %v, want an attack-requirement refusal", err)
	}
	if re.Requirement.Source != seeker || !re.Requirement.AtLeastOneCreature {
		t.Errorf("refusal = %+v, want the Seeker's player-level requirement", re.Requirement)
	}
	const want = "At least one creature must attack this combat if able (Seeker of Slaanesh)."
	if got := re.Sentence(me.ID); got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	if seekerHasPass(g, me) {
		t.Error("the enumerator offers the pass the engine refuses")
	}

	if err := g.DeclareAttacker(bear, other.ID); err != nil {
		t.Fatalf("one attacker at another opponent: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with one creature attacking: %v", err)
	}
}

// TestSeekerOfSlaaneshAsksNothingWhenNoCreatureCanAttack: tapped
// creatures owe nothing, and the Seeker's controller is not its own
// opponent. The Seeker itself has haste.
func TestSeekerOfSlaaneshAsksNothingWhenNoCreatureCanAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Seeker of Slaanesh", "Creature — Demon", seekerOfSlaaneshOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with every creature tapped: %v", err)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	seeker := effectstest.B12Push(g2, me2.ID, "Seeker of Slaanesh", "Creature — Demon", seekerOfSlaaneshOracle, 3, 3)
	g2.WithWriteLock(func() {
		for i := range g2.Battlefield.Cards {
			if g2.Battlefield.Cards[i].InstanceID == seeker {
				g2.Battlefield.Cards[i].SummonedThisTurn = true
			}
		}
	})
	effectstest.AdvanceToDeclareAttackersOf(t, g2, 0)
	if err := g2.PassPriority(); err != nil {
		t.Fatalf("your own Seeker asked you to attack: %v", err)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g2, seeker, "haste") {
		t.Error("Seeker of Slaanesh lacks haste")
	}
}
