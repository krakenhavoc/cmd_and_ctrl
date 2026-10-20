package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// exile_grantor_cost_test.go — #4000: the engine rules of
// AbilityCost.ExileGrantor. The Dominion Bracelet's whole activation is
// pinned in cards/effects/the_dominion_bracelet_test.go.

func TestExileGrantorReadsTheRowsGrantor(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bracelet := NewCard("Bracelet", me.ID)
	g.Battlefield.PushTop(bracelet)
	cost := AbilityCost{ExileGrantor: true}
	granted := AbilityOrigin{Ref: GrantedAbilityRef("x/y", 0, 0), Grant: "x/y", GrantedBy: bracelet.InstanceID}

	var got uuid.UUID
	var err error
	g.WithWriteLock(func() { got, err = g.exileGrantorOfLocked(cost, granted) })
	if err != nil || got != bracelet.InstanceID {
		t.Fatalf("grantor = %v, %v; want the Bracelet", got, err)
	}
	// No component, no grantor and no error.
	g.WithWriteLock(func() { got, err = g.exileGrantorOfLocked(AbilityCost{}, granted) })
	if err != nil || got != uuid.Nil {
		t.Errorf("no component: %v, %v", got, err)
	}
	// An own row has no grantor; a grantor that left can't pay.
	for name, o := range map[string]AbilityOrigin{
		"own row":         {Ref: OwnAbilityRef(0)},
		"grantor is gone": {Ref: granted.Ref, Grant: "x/y", GrantedBy: uuid.New()},
	} {
		g.WithWriteLock(func() { got, err = g.exileGrantorOfLocked(cost, o) })
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%s: %v, %v; want ErrInvalidParam", name, got, err)
		}
	}
}

// One permanent pays one component (CR 118.3).
func TestExileGrantorPaysNoOtherComponent(t *testing.T) {
	grantor, source := uuid.New(), uuid.New()
	cost := AbilityCost{ExileGrantor: true}
	if err := validateExileGrantorCostLocked(grantor, source, cost, nil, nil); err != nil {
		t.Errorf("a grantor no other component names: %v", err)
	}
	if err := validateExileGrantorCostLocked(grantor, source, cost, []uuid.UUID{grantor}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("a grantor also sacrificed: %v, want ErrInvalidParam", err)
	}
	selfMoving := AbilityCost{ExileGrantor: true, SacrificeSelf: true}
	if err := validateExileGrantorCostLocked(source, source, selfMoving); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("a grantor that is the source the cost also sacrifices: %v, want ErrInvalidParam", err)
	}
	if err := validateExileGrantorCostLocked(uuid.Nil, source, AbilityCost{}, []uuid.UUID{source}); err != nil {
		t.Errorf("no grantor: %v", err)
	}
}
