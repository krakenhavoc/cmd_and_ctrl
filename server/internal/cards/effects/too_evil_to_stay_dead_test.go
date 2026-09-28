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

// TestTooEvilToStayDeadStillCapsAtFourEvenWithTeamwork pins the
// declared caveat: teamwork does not widen the target, because
// TargetSpec.CardOK has no way to read the announced optional cost
// (see the doc comment). A creature with mana value 5 stays an
// illegal target even when teamwork is paid.
func TestTooEvilToStayDeadStillCapsAtFourEvenWithTeamwork(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushVanillaCreature(g, me.ID, "Elf A", 4, 4)
	b := pushVanillaCreature(g, me.ID, "Elf B", 4, 4)
	big := pushGraveyardCreature(g, me.ID, "Big Beater", "{3}{G}{G}")
	if _, err := castPaying(t, g, "Too Evil to Stay Dead", "Sorcery", tooEvilToStayDeadOracle, twTarget(big),
		[]int{0}, []uuid.UUID{a, b}, nil); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("err = %v, want ErrIllegalTarget — the caveat says teamwork does not widen the target", err)
	}
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
