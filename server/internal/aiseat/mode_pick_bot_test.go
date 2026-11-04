package aiseat_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// mode_pick_bot_test.go — #4674 through the real cards, the real wire,
// the real enumerator and the real heuristic. A "choose up to" mode_pick
// used to score every answer the same, so the bot took the first, none,
// and never chose a mode. The prompt now carries each offered bullet's
// purpose and legal targets (`mode_bullets`, the chooser's alone), and
// the bot takes the bullets that help it.
//
// Not behind AISEAT_GAME_TESTS: one cast, a few decisions.

const (
	botErtaiResurrectedOracle       = "3d038f7c-95fa-4b71-8f74-b9b4dd45cde0"
	botTheRuinousWreckingCrewOracle = "0ff9c0cc-4fb3-4117-9a8e-5b51f62306c4"
)

// enterOwingModePick casts `c` from the active seat's hand (the sandbox
// cast skips the mana gate) and passes priority until the seat owes its
// enters trigger's mode_pick.
func enterOwingModePick(t *testing.T, g *game.Game, me *game.Player, c game.Card, x int) *game.PendingChoice {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	c.AddKnowersAll(seatIDs(g))
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	for i := 0; i < 16 && !owesChoice(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceModePick && pc.Chooser == me.ID {
			return pc
		}
	}
	t.Fatalf("%s's enters trigger did not ask for its mode", c.Name)
	return nil
}

// putCreature puts a vanilla creature onto the battlefield under `p`.
func putCreature(g *game.Game, p *game.Player, name string, power, toughness int) uuid.UUID {
	id := uuid.New()
	c := game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Beast", ManaCost: "{3}{G}",
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID}
	c.AddKnowersAll(seatIDs(g))
	g.Battlefield.PushTop(c)
	return id
}

func otherSeat(g *game.Game, me *game.Player) *game.Player {
	for _, s := range g.Seats {
		if s.ID != me.ID {
			return s
		}
	}
	return nil
}

// bulletsFor is the mode_pick's mode_bullets as `viewer` sees them.
func bulletsFor(g *game.Game, viewer uuid.UUID, choice uuid.UUID) (protocol.PendingChoiceView, bool) {
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, c := range v.PendingChoices {
		if c.ID == choice.String() {
			return c, true
		}
	}
	return protocol.PendingChoiceView{}, false
}

// decidedModes is the bullets `pol` answers the open mode_pick with.
func decidedModes(t *testing.T, g *game.Game, me *game.Player, pol aiseat.Policy) []int {
	t.Helper()
	moves := legal.EnumerateFor(g, me.ID)
	d, err := pol.Decide(context.Background(), aiseat.Input{View: protocol.ViewOfGameFor(g, me.ID.String()), Seat: me.ID, Moves: moves})
	if err != nil {
		t.Fatal(err)
	}
	var p struct {
		Modes []int `json:"modes"`
	}
	if err := json.Unmarshal(moves[d.Index].Params, &p); err != nil {
		t.Fatalf("params %s: %v", moves[d.Index].Params, err)
	}
	return p.Modes
}

func noModePickPricing() aiseat.Policy {
	cfg := heuristic.DefaultConfig()
	cfg.PriceModePicks = false
	return heuristic.NewWithConfig(cfg)
}

func ertai() game.Card {
	return game.Card{Name: "Ertai Resurrected", OracleID: botErtaiResurrectedOracle,
		TypeLine: "Legendary Creature — Phyrexian Human Wizard", ManaCost: "{2}{U}{B}", Power: 3, Toughness: 2}
}

func TestHeuristicErtaiDestroysTheOpponentsCreature(t *testing.T) {
	for _, priced := range []bool{true, false} {
		g := newSettledTable(t, 4674)
		advanceToStep(t, g, game.StepPrecombatMain)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := otherSeat(g, me)
		me.Hand.Cards = nil
		hoof := putCreature(g, opp, "Craterhoof Behemoth", 5, 5)
		pc := enterOwingModePick(t, g, me, ertai(), 0)

		view, ok := bulletsFor(g, me.ID, pc.ID)
		if !ok || len(view.ModeBullets) != len(view.ModeIndexes) || len(view.ModeBullets) == 0 {
			t.Fatalf("the chooser's prompt carries %d bullets for %d indexes", len(view.ModeBullets), len(view.ModeIndexes))
		}
		destroy := view.ModeBullets[slices.Index(view.ModeIndexes, 1)]
		if destroy.LegalTargets == nil || !destroy.LegalTargets.Harmful || !slices.Contains(destroy.LegalTargets.Cards, hoof.String()) {
			t.Fatalf("the destroy bullet's targets = %+v, want a harmful clause that can take the opponent's creature", destroy.LegalTargets)
		}
		if theirs, _ := bulletsFor(g, opp.ID, pc.ID); theirs.ModeBullets != nil {
			t.Error("an opponent's copy of the prompt carries the chooser's bullets")
		}

		var pol aiseat.Policy = heuristic.New()
		if !priced {
			pol = noModePickPricing()
		}
		driveChoices(t, g, pol, me.ID, 4)
		for i := 0; i < 16 && (g.Stack.Size() > 0 || len(g.StackMeta) > 0 || len(g.PendingTriggers) > 0); i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		destroyed := opp.Graveyard.Contains(hoof)
		if priced && !destroyed {
			t.Error("the bot did not destroy the opponent's Craterhoof with Ertai's trigger")
		}
		if !priced && destroyed {
			t.Error("PriceModePicks off: the bot chose a mode (want #4674's old answer, none)")
		}
	}
}

func TestHeuristicErtaiSparesItsOwnCreature(t *testing.T) {
	g := newSettledTable(t, 4675)
	advanceToStep(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	mine := putCreature(g, me, "Grizzly Bears", 2, 2)
	enterOwingModePick(t, g, me, ertai(), 0)
	if got := decidedModes(t, g, me, heuristic.New()); len(got) != 0 {
		t.Errorf("with only its own creature to destroy, the bot chose modes %v, want none", got)
	}
	driveChoices(t, g, heuristic.New(), me.ID, 4)
	if !g.Battlefield.Contains(mine) {
		t.Error("the bot destroyed its own creature")
	}
}

func TestHeuristicWreckingCrewTakesTheRummageAndTheDrain(t *testing.T) {
	g := newSettledTable(t, 4676)
	advanceToStep(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	crew := game.Card{Name: "The Ruinous Wrecking Crew", OracleID: botTheRuinousWreckingCrewOracle,
		TypeLine: "Legendary Creature — Human Villain", ManaCost: "{X}{B}{R}", Power: 2, Toughness: 2}
	pc := enterOwingModePick(t, g, me, crew, 2)
	if pc.ModeMax != 2 {
		t.Fatalf("X = 2: mode_max %d", pc.ModeMax)
	}
	view, _ := bulletsFor(g, me.ID, pc.ID)
	rummage := view.ModeBullets[slices.Index(view.ModeIndexes, 0)]
	if rummage.Purpose == nil || rummage.Purpose.Draws != 1 || rummage.Purpose.Discards != 1 {
		t.Fatalf("the rummage bullet's purpose = %+v, want one draw and one discard", rummage.Purpose)
	}
	got := decidedModes(t, g, me, heuristic.New())
	slices.Sort(got)
	if !slices.Equal(got, []int{0, 1}) {
		t.Errorf("X = 2 with no token on the board: chose modes %v, want the rummage and the drain [0 1]", got)
	}
	if got := decidedModes(t, g, me, noModePickPricing()); len(got) != 0 {
		t.Errorf("PriceModePicks off: chose modes %v, want none", got)
	}
}
