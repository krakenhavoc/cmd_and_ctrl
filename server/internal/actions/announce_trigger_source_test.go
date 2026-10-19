package actions

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stealCreature puts a creature owned by r.affected on the battlefield
// under r.other's control and returns its instance ID.
func stealCreature(t *testing.T, r gateRoles) (string, uuid.UUID) {
	t.Helper()
	id := pushCreature(t, r.g, r.affected)
	stolen := uuid.MustParse(id)
	for i := range r.g.Battlefield.Cards {
		if r.g.Battlefield.Cards[i].InstanceID == stolen {
			r.g.Battlefield.Cards[i].Controller = r.other.ID
		}
	}
	return id, stolen
}

// CR 603.3a, CR 603.10a: a dies trigger belongs to whoever controlled the
// creature as it last existed on the battlefield. A stolen creature that
// died sits in its OWNER's graveyard, so the thief must still be able to
// announce it by hand (the owner may too), and a third seat may not.
func TestAnnounceTriggerForAStolenCreatureThatDied(t *testing.T) {
	cases := []struct {
		name    string
		role    func(r gateRoles) *game.Player
		allowed bool
	}{
		{"last-known controller (the thief)", func(r gateRoles) *game.Player { return r.other }, true},
		{"owner", func(r gateRoles) *game.Player { return r.affected }, true},
		{"third seat", func(r gateRoles) *game.Player { return r.active }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newGateRoles(t)
			id, stolen := stealCreature(t, r)
			if err := r.g.SacrificePermanent(r.other.ID, stolen); err != nil {
				t.Fatalf("sacrifice the stolen creature: %v", err)
			}
			// Whatever Controller the dead card carries (a zone move may or
			// may not reset it), the gate must not depend on it: pin it to
			// the owner so only the owner and the last-known controller
			// recorded by the leaves-the-battlefield event can claim it.
			grave := r.affected.Graveyard.Cards
			found := false
			for i := range grave {
				if grave[i].InstanceID == stolen {
					grave[i].Controller = r.affected.ID
					found = true
				}
			}
			if !found {
				t.Fatalf("setup: the stolen creature is not in its owner's graveyard")
			}

			who := tc.role(r)
			a := mustAction(t, TypeAnnounceTrigger, params(t, map[string]any{
				"source_card_id": id, "label": "dies",
			}))
			a.Player, a.Caller = who.ID, who.ID
			err := Dispatch(r.g, a)
			if tc.allowed && err != nil {
				t.Fatalf("err = %v, want allowed", err)
			}
			if !tc.allowed && !errors.Is(err, game.ErrCardCallerMismatch) {
				t.Fatalf("err = %v, want ErrCardCallerMismatch", err)
			}
		})
	}
}

// A permanent still on the battlefield has one claimant: its controller,
// not its owner.
func TestAnnounceTriggerForAStolenCreatureOnTheBattlefieldIsTheControllers(t *testing.T) {
	r := newGateRoles(t)
	id, _ := stealCreature(t, r)
	for _, tc := range []struct {
		who     *game.Player
		allowed bool
	}{{r.other, true}, {r.affected, false}} {
		a := mustAction(t, TypeAnnounceTrigger, params(t, map[string]any{"source_card_id": id, "label": "x"}))
		a.Player, a.Caller = tc.who.ID, tc.who.ID
		err := Dispatch(r.g, a)
		if tc.allowed != (err == nil) {
			t.Errorf("%s: err = %v, allowed = %v", tc.who.Name, err, tc.allowed)
		}
	}
}
