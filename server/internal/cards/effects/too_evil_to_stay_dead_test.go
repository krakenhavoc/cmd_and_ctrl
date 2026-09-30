package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const tooEvilToStayDeadOracle = "e50fecff-8872-42f5-8882-41ad13d9d1ae"

// pushGraveyardCreature seeds a creature card with a given printed
// mana cost into a player's graveyard, and returns its ID.
func pushGraveyardCreature(g *game.Game, owner uuid.UUID, name, manaCost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			if p.ID != owner {
				continue
			}
			p.Graveyard.PushTop(game.Card{
				InstanceID: id,
				Name:       name,
				TypeLine:   "Creature — Bear",
				ManaCost:   manaCost,
				Owner:      owner,
				Controller: owner,
			})
		}
	})
	return id
}

func TestTooEvilToStayDeadReanimatesASmallCreatureWithoutTeamwork(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small := pushGraveyardCreature(g, me.ID, "Bear", "{2}{G}")
	if _, err := castPaying(t, g, "Too Evil to Stay Dead", "Sorcery", tooEvilToStayDeadOracle, twTarget(small),
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the small creature should have returned to the battlefield")
	}
}

// TestTooEvilToStayDeadCapsAtFourWithoutTeamwork: the printed clause
// stands on a plain cast — a creature card with mana value 5 is an
// illegal target at announce (CR 601.2c).
func TestTooEvilToStayDeadCapsAtFourWithoutTeamwork(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big := pushGraveyardCreature(g, me.ID, "Big Beater", "{3}{G}{G}")
	if _, err := castPaying(t, g, "Too Evil to Stay Dead", "Sorcery", tooEvilToStayDeadOracle, twTarget(big),
		nil, nil, nil); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("err = %v, want ErrIllegalTarget — without teamwork the ceiling holds", err)
	}
}

// castTooEvilWithTeamworkAtBig puts a teamwork-cast Too Evil to Stay
// Dead on the stack aimed at a mana-value-5 creature card, and returns
// that card. It was the gap-pinning test until #1716 flipped it: the
// announce accepts the wider clause.
func castTooEvilWithTeamworkAtBig(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushVanillaCreature(g, me.ID, "Elf A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Elf B", 2, 2)
	big := pushGraveyardCreature(g, me.ID, "Big Beater", "{3}{G}{G}")
	if _, err := castPaying(t, g, "Too Evil to Stay Dead", "Sorcery", tooEvilToStayDeadOracle, twTarget(big),
		[]int{0}, []uuid.UUID{a, b}, nil); err != nil {
		t.Fatalf("a teamwork cast may target any creature card in your graveyard: %v", err)
	}
	return big
}

// TestTooEvilToStayDeadTeamworkLiftsTheCeiling is the whole card: the
// teamwork cast is announced under "target creature card in your
// graveyard", and at resolution the CR 608.2b re-check reads the SAME
// clause off the stack item's paid record — judged under the printed
// clause the Big Beater would be illegal and the spell would fizzle.
func TestTooEvilToStayDeadTeamworkLiftsTheCeiling(t *testing.T) {
	g := newCatalogGame(t)
	big := castTooEvilWithTeamworkAtBig(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) {
		t.Error("the teamwork cast must return the mana-value-5 creature (the resolution re-check lost the paid record)")
	}
}

// TestTooEvilToStayDeadWidenedClauseSurvivesUndoAndRestore: the clause
// is catalog data re-derived from the paid record, so a clone (what
// undo restores) and a JSON snapshot round trip both still judge the
// teamwork cast under the wide clause at resolution.
func TestTooEvilToStayDeadWidenedClauseSurvivesUndoAndRestore(t *testing.T) {
	t.Run("clone", func(t *testing.T) {
		g := newCatalogGame(t)
		big := castTooEvilWithTeamworkAtBig(t, g)
		clone := g.Clone()
		passPriorityAroundTable(t, clone)
		if !clone.Battlefield.Contains(big) {
			t.Error("the cloned game lost the teamwork clause")
		}
	})
	t.Run("snapshot", func(t *testing.T) {
		g := newCatalogGame(t)
		big := castTooEvilWithTeamworkAtBig(t, g)
		restored := restoreRoundTrip(t, g, false)
		passPriorityAroundTable(t, restored)
		if !restored.Battlefield.Contains(big) {
			t.Error("the restored game lost the teamwork clause")
		}
	})
}

func TestTooEvilToStayDeadRefusesTeamworkWithTooLittlePower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small := pushGraveyardCreature(g, me.ID, "Bear", "{2}{G}")
	weak := pushVanillaCreature(g, me.ID, "Elf", 1, 1)
	if _, err := castPaying(t, g, "Too Evil to Stay Dead", "Sorcery", tooEvilToStayDeadOracle, twTarget(small),
		[]int{0}, []uuid.UUID{weak}, nil); !errors.Is(err, game.ErrInsufficientTeamwork) {
		t.Fatalf("err = %v, want ErrInsufficientTeamwork", err)
	}
}
