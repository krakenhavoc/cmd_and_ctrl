package game

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// unattach_cost_test.go — #4799: the engine rules of
// AbilityCost.UnattachEquipment. Captain America, First Avenger's whole
// activation is pinned in cards/effects/captain_america_first_avenger_test.go.

// unattachTable is a creature with two Equipment attached (one its
// controller's, one an opponent's), an Equipment attached to another
// creature, a loose one, and a non-Equipment artifact attached to it.
type unattachTable struct {
	g                                   *Game
	host, other                         uuid.UUID
	mine, theirs, elsewhere, loose, rod uuid.UUID
}

func newUnattachTable(t *testing.T) unattachTable {
	t.Helper()
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	put := func(name, typeLine string, owner uuid.UUID, host uuid.UUID) uuid.UUID {
		c := NewCard(name, owner)
		c.TypeLine = typeLine
		if host != uuid.Nil {
			c.AttachedTo = TargetRef{Kind: TargetCard, ID: host}
		}
		g.Battlefield.PushTop(c)
		return c.InstanceID
	}
	tb := unattachTable{g: g}
	tb.host = put("Host", "Creature — Human", me.ID, uuid.Nil)
	tb.other = put("Other", "Creature — Bear", me.ID, uuid.Nil)
	tb.mine = put("Mine", "Artifact — Equipment", me.ID, tb.host)
	tb.theirs = put("Theirs", "Artifact — Equipment", opp.ID, tb.host)
	tb.elsewhere = put("Elsewhere", "Artifact — Equipment", me.ID, tb.other)
	tb.loose = put("Loose", "Artifact — Equipment", me.ID, uuid.Nil)
	tb.rod = put("Rod", "Artifact", me.ID, tb.host)
	return tb
}

// The candidates are the Equipment attached to the source, whoever
// controls them (CR 301.5d), and nothing else.
func TestUnattachCostOptionsAreTheEquipmentAttachedToTheSource(t *testing.T) {
	tb := newUnattachTable(t)
	var got []uuid.UUID
	tb.g.ReadSnapshot(func() { got = tb.g.UnattachCostOptionsLocked(tb.host) })
	if len(got) != 2 || !slices.Contains(got, tb.mine) || !slices.Contains(got, tb.theirs) {
		t.Errorf("options = %v, want the two Equipment attached to the host (%v, %v)", got, tb.mine, tb.theirs)
	}
	tb.g.ReadSnapshot(func() { got = tb.g.UnattachCostOptionsLocked(tb.loose) })
	if len(got) != 0 {
		t.Errorf("an Equipment's own options = %v, want none", got)
	}
}

func TestUnattachCostValidation(t *testing.T) {
	tb := newUnattachTable(t)
	cases := []struct {
		name    string
		n       int
		zone    ZoneKind
		ids     []uuid.UUID
		wantErr error
	}{
		{"one attached Equipment", 1, ZoneBattlefield, []uuid.UUID{tb.mine}, nil},
		{"an opponent's Equipment attached to it", 1, ZoneBattlefield, []uuid.UUID{tb.theirs}, nil},
		{"no component, nothing named", 0, ZoneBattlefield, nil, nil},
		{"no component, an id sent", 0, ZoneBattlefield, []uuid.UUID{tb.mine}, ErrInvalidParam},
		{"nothing named (CR 118.3)", 1, ZoneBattlefield, nil, ErrInvalidParam},
		{"two named for one", 1, ZoneBattlefield, []uuid.UUID{tb.mine, tb.theirs}, ErrInvalidParam},
		{"the same one twice", 2, ZoneBattlefield, []uuid.UUID{tb.mine, tb.mine}, ErrInvalidParam},
		{"another creature's Equipment", 1, ZoneBattlefield, []uuid.UUID{tb.elsewhere}, ErrIllegalTarget},
		{"an unattached Equipment", 1, ZoneBattlefield, []uuid.UUID{tb.loose}, ErrIllegalTarget},
		{"an attached non-Equipment", 1, ZoneBattlefield, []uuid.UUID{tb.rod}, ErrIllegalTarget},
		{"a source off the battlefield", 1, ZoneGraveyard, []uuid.UUID{tb.mine}, ErrInvalidParam},
	}
	for _, tc := range cases {
		var err error
		tb.g.ReadSnapshot(func() { err = tb.g.validateUnattachCostLocked(tb.host, tc.zone, tc.n, tc.ids) })
		if tc.wantErr == nil && err != nil || tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
			t.Errorf("%s: %v, want %v", tc.name, err, tc.wantErr)
		}
	}
}

// Paying unattaches (CR 701.3d): the Equipment stays on the battlefield
// under its controller, attached to nothing, and EventUnattach fires.
func TestUnattachCostPaymentLeavesTheEquipmentOnTheBattlefield(t *testing.T) {
	tb := newUnattachTable(t)
	g := tb.g
	before := len(g.Events)
	g.WithWriteLock(func() { g.payUnattachCostLocked([]uuid.UUID{tb.theirs}) })
	c := findBattlefieldCard(g, tb.theirs)
	if c == nil {
		t.Fatal("the Equipment left the battlefield")
	}
	if c.IsAttached() || c.Controller != g.Seats[1].ID {
		t.Errorf("after paying: attached %v, controller %v", c.AttachedTo, c.Controller)
	}
	if m := findBattlefieldCard(g, tb.mine); m == nil || !m.IsAttachedTo(tb.host) {
		t.Error("the Equipment not named was unattached too")
	}
	fired := false
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventUnattach && ev.CardID == tb.theirs && ev.Target == tb.host {
			fired = true
		}
	}
	if !fired {
		t.Error("no EventUnattach for the paid Equipment")
	}
}

// A divided amount sees the pick (DivideAmountArgs.Unattached), at bind
// time, exactly as the activation binds it.
var testUnattachedCount = RegisterDivideAmount("test-4799/unattached-count",
	func(_ *Game, a DivideAmountArgs) DivideSpec { return DivideSpec{Total: 3 * len(a.Unattached)} })

func TestADividedAmountReadsTheUnattachPick(t *testing.T) {
	g := newActiveGame(t)
	steps := []AnnouncedClause{{Clause: TargetSpec{Min: 1, Max: 3, Divide: &DivideSpec{AmountKey: testUnattachedCount}}}}
	g.WithWriteLock(func() {
		g.bindDivideAmountsLocked(steps, DivideAmountArgs{Unattached: []uuid.UUID{uuid.New()}})
	})
	if got := steps[0].Clause.Divide.TotalFor(0); got != 3 {
		t.Errorf("bound amount = %d, want 3 (one Equipment named)", got)
	}
}

// The pick is part of the announcement a parked activation writes into
// the snapshot (pendingChoices[].parked.activate.params.UnattachIDs), so
// a re-run after a restore pays the same Equipment.
func TestAParkedActivationKeepsItsUnattachPick(t *testing.T) {
	id := uuid.New()
	raw, err := json.Marshal(newActivateAnnounce(ActivateAbilityParams{UnattachIDs: []uuid.UUID{id}}))
	if err != nil {
		t.Fatal(err)
	}
	var back ActivateAnnounce
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	if got := back.params().UnattachIDs; len(got) != 1 || got[0] != id {
		t.Errorf("restored pick = %v, want [%v]", got, id)
	}
}
