package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// target_purpose_view_test.go — ADR 0126's amendment of 2026-10-08: a
// purpose's target entries on the wire, on the card and on a mode, and
// hidden with the rest of the purpose.

const (
	pvSignInBlood     = "c6207f6a-a624-4754-88f5-dbe700c841ff"
	pvPrismariCommand = "fa3e28b1-131c-4223-81e0-18dfbab22c26"
)

func TestViewOfPurposeProjectsTargetEntries(t *testing.T) {
	p := game.Purpose{Draws: 1, Targets: game.ForTargets(
		game.TargetPurpose{Slot: 0, Damage: 2},
		game.TargetPurpose{Slot: 1, Draws: 2, Discards: 1, Tokens: 3, LifeGain: 4, LifeLoss: 5},
	)}
	v := viewOfPurpose(p)
	if v == nil || v.Targets == nil {
		t.Fatalf("viewOfPurpose = %+v, want target entries", v)
	}
	want := []TargetPurposeView{
		{Slot: 0, Damage: 2},
		{Slot: 1, Draws: 2, Discards: 1, Tokens: 3, LifeGain: 4, LifeLoss: 5},
	}
	if got := *v.Targets; len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("targets = %+v, want %+v", got, want)
	}
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	const wantJSON = `{"draws":1,"targets":[{"slot":0,"damage":2},{"slot":1,"draws":2,"discards":1,"tokens":3,"life_gain":4,"life_loss":5}]}`
	if string(raw) != wantJSON {
		t.Errorf("wire = %s, want %s", raw, wantJSON)
	}
	// A purpose of target entries alone is still a purpose.
	if v := viewOfPurpose(game.Purpose{Targets: game.ForTargets(game.TargetPurpose{Damage: 3})}); v == nil {
		t.Error("a purpose with only target entries projected to nil")
	}
	// And one without any sends no key.
	raw, _ = json.Marshal(viewOfPurpose(game.Purpose{Draws: 2}))
	if strings.Contains(string(raw), "targets") {
		t.Errorf("a purpose with no target entries sends %s", raw)
	}
}

func TestTargetEntriesRideTheView(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sign := rowsFixtureCard("Sign in Blood", "Sorcery", pvSignInBlood, me.ID)
	prismari := rowsFixtureCard("Prismari Command", "Instant", pvPrismariCommand, me.ID)
	g.WithWriteLock(func() {
		for _, c := range []game.Card{sign, prismari} {
			c.KnownBy = map[uuid.UUID]bool{me.ID: true}
			me.Hand.PushTop(c)
		}
	})
	g.BumpLayerVersionForTest()
	v := ViewOfGame(g)

	hand := seatZones(FilterViewFor(v, me.ID.String()), me.ID.String())
	c := findCardView(t, hand, sign.InstanceID)
	if c == nil || c.Purpose == nil || c.Purpose.Targets == nil {
		t.Fatalf("Sign in Blood = %+v, want a target entry", c)
	}
	if got := *c.Purpose.Targets; len(got) != 1 || got[0] != (TargetPurposeView{Slot: 0, Draws: 2, LifeLoss: 2}) {
		t.Errorf("Sign in Blood's target entries = %+v", got)
	}
	if c.Purpose.Draws != 0 {
		t.Errorf("Sign in Blood declares %d draws for its controller; the draws are its target's", c.Purpose.Draws)
	}

	c = findCardView(t, hand, prismari.InstanceID)
	if c == nil || c.Modes == nil || len(c.Modes.Options) != 4 {
		t.Fatalf("Prismari Command = %+v", c)
	}
	wantModes := []*TargetPurposeView{
		{Slot: 0, Damage: 2},
		{Slot: 0, Draws: 2, Discards: 2},
		{Slot: 0, Tokens: 1},
		nil,
	}
	for i, want := range wantModes {
		p := c.Modes.Options[i].Purpose
		switch {
		case want == nil && p != nil:
			t.Errorf("Prismari mode %d declares %+v, want nothing", i, p)
		case want != nil && (p == nil || p.Targets == nil || len(*p.Targets) != 1 || (*p.Targets)[0] != *want):
			t.Errorf("Prismari mode %d purpose = %+v, want the target entry %+v", i, p, *want)
		}
	}

	t.Run("an opponent never sees them", func(t *testing.T) {
		hand := seatZones(FilterViewFor(v, opp.ID.String()), me.ID.String())
		for _, c := range hand.Cards {
			if c.Purpose != nil {
				t.Errorf("a hidden hand card carries purpose %+v to an opponent", c.Purpose)
			}
			if c.Modes != nil {
				for i, o := range c.Modes.Options {
					if o.Purpose != nil {
						t.Errorf("a hidden hand card's mode %d carries purpose %+v to an opponent", i, o.Purpose)
					}
				}
			}
		}
	})
}
