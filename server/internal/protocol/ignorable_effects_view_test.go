package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// ignorable_effects_view_test.go — #1986, CR 116.2d: the seat a Damping
// Engine binds carries the ignore as `ignorable_effects`, from the same
// accessor the engine and the enumerator ask, for its own eyes only; the
// land it refuses carries the clause on `cant_cast`; and the sacrifice the
// view names is one the engine accepts.

const oracleDampingEngineView1986 = "3584a055-5220-4721-89f8-de33f779122f"

// engineTable puts a Damping Engine under `me` and enough tokens that me
// controls strictly more permanents than any other seat.
func engineTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := stripTable(t)
	most := 0
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		n := 0
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Controller == p.ID {
				n++
			}
		}
		most = max(most, n)
	}
	have := 0
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Controller == me.ID {
			have++
		}
	}
	for ; have <= most+1; have++ {
		g.Battlefield.PushTop(publicCard(g, me.ID, "Spirit", "Token Creature — Spirit", "", ""))
	}
	engine := publicCard(g, me.ID, "Damping Engine", "Artifact", "{4}", oracleDampingEngineView1986)
	g.Battlefield.PushTop(engine)
	return g, me, opp, engine.InstanceID
}

func TestTheBoundSeatCarriesTheIgnoreAndOnlyItsOwnFrameDoes(t *testing.T) {
	g, me, opp, engine := engineTable(t)

	own := seatView(t, g, me, me)
	if len(own.IgnorableEffects) != 1 {
		t.Fatalf("own frame: %d ignorable effects, want 1", len(own.IgnorableEffects))
	}
	row := own.IgnorableEffects[0]
	if row.SourceID != engine.String() || row.SourceName != "Damping Engine" || !strings.Contains(row.Label, "ignore Damping Engine's effect until end of turn") {
		t.Errorf("row = %+v", row)
	}
	if len(row.SacrificeIDs) < 2 {
		t.Fatalf("sacrifice choices = %v, want the seat's permanents", row.SacrificeIDs)
	}
	// Every choice is the seat's own permanent, and the enumerator's pick
	// (the first) is among them.
	for _, id := range row.SacrificeIDs {
		found := false
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID.String() == id && c.Controller == me.ID {
				found = true
			}
		}
		if !found {
			t.Errorf("sacrifice choice %s is not one of the seat's permanents", id)
		}
	}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Kind == legal.KindSpecialAction && strings.Contains(string(m.Params), `"ignore_effect"`) &&
			!strings.Contains(string(m.Params), row.SacrificeIDs[0]) {
			t.Errorf("the bot's move %s does not pay with the view's first choice %s", m.Params, row.SacrificeIDs[0])
		}
	}

	// Nobody else's frame, not an opponent's, not a spectator's, carries it.
	if got := seatView(t, g, opp, me).IgnorableEffects; len(got) != 0 {
		t.Errorf("an opponent's frame carries %v", got)
	}
	for _, s := range ViewOfGameFor(g, SpectatorViewerID).Seats {
		if len(s.IgnorableEffects) != 0 {
			t.Errorf("a spectator's frame carries %v", s.IgnorableEffects)
		}
	}
	// The seat the Engine does not bind carries nothing, whoever looks.
	if got := seatView(t, g, opp, opp).IgnorableEffects; len(got) != 0 {
		t.Errorf("an unbound seat carries %v", got)
	}
}

// The banner says why the seat can't play lands, the hand land carries the
// clause, and after the ignore all three go quiet.
func TestTheIgnoreClearsTheBannerTheClauseAndTheRow(t *testing.T) {
	g, me, _, engine := engineTable(t)
	forest := publicCard(g, me.ID, "Forest", "Basic Land — Forest", "", "")
	me.Hand.PushTop(forest)

	if got := seatView(t, g, me, me).CantPlayLands; !strings.Contains(got, "Damping Engine") {
		t.Errorf("banner = %q, want the Engine's clause", got)
	}
	if got := ownHandCard(t, g, me, forest.InstanceID).CantCast; !strings.Contains(got, "Damping Engine") {
		t.Errorf("the land's cant_cast = %q, want the Engine's clause", got)
	}
	cost := seatView(t, g, me, me).IgnorableEffects[0].SacrificeIDs[0]
	id, err := uuid.Parse(cost)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.PerformSpecialAction(me.ID, engine, game.SpecialActionIgnoreEffect, game.SpecialActionParams{SacrificeID: id}); err != nil {
		t.Fatalf("the view's own choice was refused: %v", err)
	}
	v := seatView(t, g, me, me)
	if v.CantPlayLands != "" || len(v.IgnorableEffects) != 0 {
		t.Errorf("after ignoring: banner %q, rows %v", v.CantPlayLands, v.IgnorableEffects)
	}
	if got := ownHandCard(t, g, me, forest.InstanceID).CantCast; got != "" {
		t.Errorf("after ignoring the land's cant_cast = %q", got)
	}
}
