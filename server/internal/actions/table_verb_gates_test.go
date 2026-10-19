package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4131 — the table-wide sandbox verbs, per role. Three seats at one
// table: the ACTIVE seat (who also holds priority at the start), the
// AFFECTED seat (the player or the card's controller the verb acts on)
// and an OTHER seat that is neither. Every row says what each role gets;
// admin (Caller uuid.Nil) is covered once at the end.

type gateRoles struct {
	g        *game.Game
	active   *game.Player
	affected *game.Player
	other    *game.Player
	// affectedCard is a creature on the battlefield controlled by the
	// affected seat; activeCard one controlled by the active seat.
	affectedCard string
	activeCard   string
	// affectedCmdr is the affected seat's commander card.
	affectedCmdr string
}

func newGateRoles(t *testing.T) gateRoles {
	t.Helper()
	g := game.NewGame()
	for i := range 3 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Cmdr %d", i+1), uuid.Nil)}
		for j := range 20 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	active := g.Seats[g.Turn.ActiveSeat]
	if g.Seats[g.Turn.PriorityHolder] != active {
		t.Fatalf("the test table expects the active seat to hold priority")
	}
	r := gateRoles{g: g, active: active}
	for _, p := range g.Seats {
		switch {
		case p == active:
		case r.affected == nil:
			r.affected = p
		default:
			r.other = p
		}
	}
	r.activeCard = pushCreature(t, g, r.active)
	r.affectedCard = pushCreature(t, g, r.affected)
	r.affectedCmdr = r.affected.Command.Cards[0].InstanceID.String()
	return r
}

func (r gateRoles) caller(role string) uuid.UUID {
	switch role {
	case "active":
		return r.active.ID
	case "affected":
		return r.affected.ID
	case "other":
		return r.other.ID
	case "admin":
		return uuid.Nil
	}
	panic("unknown role " + role)
}

func TestTableVerbGatesPerRole(t *testing.T) {
	allow := func(roles ...string) map[string]bool {
		m := map[string]bool{"admin": true}
		for _, r := range roles {
			m[r] = true
		}
		return m
	}

	cases := []struct {
		name string
		// build returns the action as the named role would send it,
		// Player and params included.
		build   func(t *testing.T, r gateRoles, role string) Action
		allowed map[string]bool
		refuse  error
		// check, when set, asserts the effect after an allowed run.
		check func(t *testing.T, r gateRoles)
	}{
		{
			// The source is the affected seat's; the seat announcing as
			// ITSELF with a source it controls is the only way through.
			name: "announce_trigger as the source's controller",
			build: func(t *testing.T, r gateRoles, role string) Action {
				a := mustAction(t, TypeAnnounceTrigger, params(t, map[string]any{
					"source_card_id": r.affectedCard, "label": "T",
				}))
				a.Player = r.affected.ID // the trigger is the affected seat's
				return a
			},
			allowed: allow("affected"),
			// active and other send it as someone else: the Player gate.
			refuse: ErrPlayerCallerMismatch,
			check: func(t *testing.T, r gateRoles) {
				placed := 0
				for _, item := range r.g.StackMeta {
					if item != nil && item.Kind == game.StackItemTriggered && item.Label == "T" {
						placed++
						if item.Controller != r.affected.ID {
							t.Errorf("the trigger's controller = %v, want the affected seat", item.Controller)
						}
					}
				}
				if placed != 1 {
					t.Errorf("announced triggers on the stack = %d, want 1", placed)
				}
			},
		},
		{
			// As yourself, but the source is another seat's card.
			name: "announce_trigger as yourself with a source you do not control",
			build: func(t *testing.T, r gateRoles, role string) Action {
				a := mustAction(t, TypeAnnounceTrigger, params(t, map[string]any{
					"source_card_id": r.affectedCard, "label": "T",
				}))
				a.Player = r.caller(role)
				if role == "admin" {
					a.Player = r.affected.ID // an admin names the seat it acts for
				}
				return a
			},
			allowed: allow("affected"),
			refuse:  game.ErrCardCallerMismatch,
		},
		{
			name: "set_commander_damage on yourself",
			build: func(t *testing.T, r gateRoles, role string) Action {
				// Damage dealt TO the affected seat by someone's commander.
				return mustAction(t, TypeSetCommanderDamage, params(t, map[string]any{
					"from": r.active.Command.Cards[0].InstanceID.String(), "to": r.affected.ID.String(), "amount": 7,
				}))
			},
			allowed: allow("affected"),
			refuse:  ErrPlayerCallerMismatch,
			check: func(t *testing.T, r gateRoles) {
				if got := r.affected.CommanderDamage[r.active.Command.Cards[0].InstanceID]; got != 7 {
					t.Errorf("commander damage = %d, want 7", got)
				}
			},
		},
		{
			name: "mark_damage positive",
			build: func(t *testing.T, r gateRoles, role string) Action {
				return mustAction(t, TypeMarkDamage, params(t, map[string]any{"instance_id": r.affectedCard, "delta": 1}))
			},
			allowed: allow("affected"),
			refuse:  game.ErrCardCallerMismatch,
		},
		{
			name: "mark_damage negative (healing)",
			build: func(t *testing.T, r gateRoles, role string) Action {
				if err := r.g.MarkDamage(uuid.MustParse(r.affectedCard), 1); err != nil {
					t.Fatalf("seed damage: %v", err)
				}
				return mustAction(t, TypeMarkDamage, params(t, map[string]any{"instance_id": r.affectedCard, "delta": -1}))
			},
			allowed: allow("affected"),
			refuse:  game.ErrCardCallerMismatch,
		},
		{
			name: "set_goaded by yourself on a creature you do not control",
			build: func(t *testing.T, r gateRoles, role string) Action {
				return mustAction(t, TypeSetGoaded, params(t, map[string]string{
					"instance_id": r.affectedCard, "by": r.caller(role).String(),
				}))
			},
			// The goading player acts as themself; the controller too.
			allowed: allow("active", "affected", "other"),
		},
		{
			name: "set_goaded by another player",
			build: func(t *testing.T, r gateRoles, role string) Action {
				return mustAction(t, TypeSetGoaded, params(t, map[string]string{
					"instance_id": r.affectedCard, "by": r.other.ID.String(),
				}))
			},
			// Only the creature's controller, or `other` goading as itself.
			allowed: allow("affected", "other"),
			refuse:  game.ErrCardCallerMismatch,
		},
		{
			name: "set_goaded clear",
			build: func(t *testing.T, r gateRoles, role string) Action {
				return mustAction(t, TypeSetGoaded, params(t, map[string]string{
					"instance_id": r.affectedCard, "by": "",
				}))
			},
			allowed: allow("affected"),
			refuse:  game.ErrCardCallerMismatch,
		},
		{
			name: "set_promise",
			build: func(t *testing.T, r gateRoles, role string) Action {
				return mustAction(t, TypeSetPromise, params(t, map[string]any{
					"from": r.affected.ID.String(), "to": r.other.ID.String(), "count": 1,
				}))
			},
			// Either end of the deal, not the seat outside it.
			allowed: allow("affected", "other"),
			refuse:  ErrPlayerCallerMismatch,
		},
		{
			name: "set_monarch to another seat",
			build: func(t *testing.T, r gateRoles, role string) Action {
				a := mustAction(t, TypeSetMonarch, nil)
				a.Player = r.affected.ID
				return a
			},
			allowed: allow("affected"),
			refuse:  ErrMarkerNotYours,
		},
		{
			name: "set_initiative to another seat",
			build: func(t *testing.T, r gateRoles, role string) Action {
				a := mustAction(t, TypeSetInitiative, nil)
				a.Player = r.affected.ID
				return a
			},
			allowed: allow("affected"),
			refuse:  ErrMarkerNotYours,
		},
		{
			name: "set_monarch clear while another seat holds it",
			build: func(t *testing.T, r gateRoles, role string) Action {
				if err := r.g.SetMonarch(r.affected.ID); err != nil {
					t.Fatalf("seed monarch: %v", err)
				}
				return mustAction(t, TypeSetMonarch, nil)
			},
			allowed: allow("affected"),
			refuse:  ErrMarkerNotYours,
		},
		{
			name: "set_initiative clear while another seat holds it",
			build: func(t *testing.T, r gateRoles, role string) Action {
				if err := r.g.SetInitiative(r.affected.ID); err != nil {
					t.Fatalf("seed initiative: %v", err)
				}
				return mustAction(t, TypeSetInitiative, nil)
			},
			allowed: allow("affected"),
			refuse:  ErrMarkerNotYours,
		},
		{
			name: "advance_step",
			build: func(t *testing.T, r gateRoles, role string) Action {
				return mustAction(t, TypeAdvanceStep, nil)
			},
			// The active seat is also the priority holder here.
			allowed: allow("active"),
			refuse:  ErrNotTurnOrPriority,
		},
		{
			name: "clear_combat",
			build: func(t *testing.T, r gateRoles, role string) Action {
				return mustAction(t, TypeClearCombat, nil)
			},
			allowed: allow("active"),
			refuse:  ErrNotTurnOrPriority,
		},
		{
			name: "start_vote as another seat",
			build: func(t *testing.T, r gateRoles, role string) Action {
				a := mustAction(t, TypeStartVote, params(t, map[string]any{
					"topic": "x", "options": []string{"a", "b"},
				}))
				a.Player = r.affected.ID
				return a
			},
			allowed: allow("affected"),
			refuse:  ErrPlayerCallerMismatch,
		},
		{
			name: "end_vote",
			build: func(t *testing.T, r gateRoles, role string) Action {
				if _, err := r.g.StartVote(r.affected.ID, "x", []string{"a", "b"}); err != nil {
					t.Fatalf("seed vote: %v", err)
				}
				return mustAction(t, TypeEndVote, nil)
			},
			allowed: allow("affected"),
			refuse:  ErrNotVoteInitiator,
		},
	}

	for _, tc := range cases {
		for _, role := range []string{"active", "affected", "other", "admin"} {
			t.Run(tc.name+"/"+role, func(t *testing.T) {
				r := newGateRoles(t)
				a := tc.build(t, r, role)
				a.Caller = r.caller(role)
				err := Dispatch(r.g, a)
				if tc.allowed[role] {
					if err != nil {
						t.Fatalf("%s: err = %v, want allowed", role, err)
					}
					if tc.check != nil {
						tc.check(t, r)
					}
					return
				}
				if !errors.Is(err, tc.refuse) {
					t.Fatalf("%s: err = %v, want %v", role, err, tc.refuse)
				}
			})
		}
	}
}

// A vote whose initiator left the table would sit open for ever if only
// the initiator could end it, so any seat may close that one.
func TestEndVoteOpensToAnySeatOnceTheInitiatorIsGone(t *testing.T) {
	r := newGateRoles(t)
	if _, err := r.g.StartVote(r.affected.ID, "x", []string{"a", "b"}); err != nil {
		t.Fatalf("StartVote: %v", err)
	}
	end := mustAction(t, TypeEndVote, nil)
	end.Caller = r.other.ID
	if err := Dispatch(r.g, end); !errors.Is(err, ErrNotVoteInitiator) {
		t.Fatalf("while the initiator is seated: err = %v, want ErrNotVoteInitiator", err)
	}
	r.affected.Eliminated = true
	if err := Dispatch(r.g, end); err != nil {
		t.Fatalf("once the initiator is eliminated: err = %v, want nil", err)
	}
	if r.g.Vote != nil {
		t.Error("the vote stayed open")
	}
}

// A monarch or initiative holder may hand the marker on, and a seat
// takes it for itself; neither is the admin's alone.
func TestMarkerHolderMayHandOnAndAnySeatMayTake(t *testing.T) {
	for _, ty := range []Type{TypeSetMonarch, TypeSetInitiative} {
		t.Run(string(ty), func(t *testing.T) {
			r := newGateRoles(t)
			take := mustAction(t, ty, nil)
			take.Player, take.Caller = r.other.ID, r.other.ID
			if err := Dispatch(r.g, take); err != nil {
				t.Fatalf("take for yourself: %v", err)
			}
			hand := mustAction(t, ty, nil)
			hand.Player, hand.Caller = r.affected.ID, r.other.ID // holder gives it away
			if err := Dispatch(r.g, hand); err != nil {
				t.Fatalf("holder hands on: %v", err)
			}
			steal := mustAction(t, ty, nil)
			steal.Player, steal.Caller = r.active.ID, r.other.ID // no longer the holder
			if err := Dispatch(r.g, steal); !errors.Is(err, ErrMarkerNotYours) {
				t.Fatalf("a former holder crowning a third seat: err = %v, want ErrMarkerNotYours", err)
			}
		})
	}
}

// The priority holder, not only the active seat, may advance the step
// and clear combat, which is how passing works.
func TestAdvanceAndClearCombatOpenToThePriorityHolder(t *testing.T) {
	for _, ty := range []Type{TypeAdvanceStep, TypeClearCombat} {
		t.Run(string(ty), func(t *testing.T) {
			r := newGateRoles(t)
			// Hand priority to the next seat, so holder != active.
			pass := mustAction(t, TypePassPriority, nil)
			pass.Caller = r.active.ID
			if err := Dispatch(r.g, pass); err != nil {
				t.Fatalf("pass_priority: %v", err)
			}
			holder := r.g.Seats[r.g.Turn.PriorityHolder]
			if holder == r.active {
				t.Skip("priority did not move")
			}
			a := mustAction(t, ty, nil)
			a.Caller = holder.ID
			if err := Dispatch(r.g, a); err != nil {
				t.Fatalf("priority holder: err = %v, want allowed", err)
			}
			var bystander *game.Player
			for _, p := range r.g.Seats {
				if p != holder && p != r.active {
					bystander = p
				}
			}
			b := mustAction(t, ty, nil)
			b.Caller = bystander.ID
			if err := Dispatch(r.g, b); !errors.Is(err, ErrNotTurnOrPriority) {
				t.Fatalf("a bystander: err = %v, want ErrNotTurnOrPriority", err)
			}
		})
	}
}

// counter_spell and counter_ability stay with the priority holder: a
// counter effect may name any object on the stack, so the object's
// controller is the wrong test, and the sandbox hand-resolution is what
// the priority holder is doing when it casts the counter.
func TestCounterVerbsStayWithThePriorityHolder(t *testing.T) {
	for _, ty := range []Type{TypeCounterSpell, TypeCounterAbility} {
		t.Run(string(ty), func(t *testing.T) {
			r := newGateRoles(t)
			a := mustAction(t, ty, json.RawMessage(fmt.Sprintf(`{"instance_id":%q}`, uuid.NewString())))
			a.Caller = r.other.ID
			if err := Dispatch(r.g, a); !errors.Is(err, ErrNotPriorityHolder) {
				t.Fatalf("non-holder: err = %v, want ErrNotPriorityHolder", err)
			}
		})
	}
}
