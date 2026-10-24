package aiseat_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// mother_of_runes_save_bot_test.go — #4312, whole-engine: the real
// catalog Mother of Runes and Murder, the real enumerator, the real
// heuristic. An opponent casts Murder at the bot's creature; the bot
// answers by giving it protection from black, and the Murder is countered
// on resolution for want of a legal target (CR 608.2b, 702.16b).
//
// Before #4312 the grant was priced by the attack it adds on the
// opponent's turn, which is nothing, and the bot let the creature die
// with Mother of Runes untapped beside it.
//
// Not behind AISEAT_GAME_TESTS: one priority exchange, milliseconds.

const (
	botMotherOfRunesOracle = "60433b48-d27f-413c-905c-43839b1943f1"
	botMurderOracle        = "938b4e2c-88d9-4637-bc00-e228920c9a78"
)

// murderTable parks a two-seat game on the opponent's first main phase:
// the bot (the seat after the active one) has Mother of Runes, able to
// tap, and a 4/4; the opponent has three Swamps and Murder in hand.
func murderTable(t *testing.T, seed uint64) (g *game.Game, me, them *game.Player, mother, ogre, murder uuid.UUID) {
	t.Helper()
	g = newSettledTable(t, seed)
	advanceToStep(t, g, game.StepPrecombatMain)
	them = g.Seats[g.Turn.ActiveSeat]
	me = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	everyone := seatIDs(g)
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		them.Hand.Cards = nil
		m := game.Card{InstanceID: uuid.New(), Name: "Mother of Runes", OracleID: botMotherOfRunesOracle,
			TypeLine: "Creature — Human Cleric", ManaCost: "{W}", Colors: []string{"W"}, Power: 1, Toughness: 1,
			Owner: me.ID, Controller: me.ID}
		m.AddKnowersAll(everyone)
		g.Battlefield.PushTop(m)
		mother = m.InstanceID
		o := game.Card{InstanceID: uuid.New(), Name: "Test Ogre", TypeLine: "Creature — Ogre",
			ManaCost: "{3}{G}", Colors: []string{"G"}, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID}
		o.AddKnowersAll(everyone)
		g.Battlefield.PushTop(o)
		ogre = o.InstanceID
		for i := 0; i < 3; i++ {
			s := game.Card{InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp",
				Owner: them.ID, Controller: them.ID}
			s.AddKnowersAll(everyone)
			g.Battlefield.PushTop(s)
		}
		k := game.Card{InstanceID: uuid.New(), Name: "Murder", OracleID: botMurderOracle, TypeLine: "Instant",
			ManaCost: "{1}{B}{B}", Colors: []string{"B"}, Owner: them.ID, Controller: them.ID,
			KnownBy: map[uuid.UUID]bool{them.ID: true}}
		them.Hand.PushTop(k)
		murder = k.InstanceID
	})
	g.BumpLayerVersionForTest()
	return g, me, them, mother, ogre, murder
}

// castMurderAt has the opponent cast Murder at the target.
func castMurderAt(t *testing.T, g *game.Game, them *game.Player, murder, target uuid.UUID) {
	t.Helper()
	for _, m := range legal.EnumerateFor(g, them.ID) {
		if m.Kind != legal.KindCast || m.Source != murder {
			continue
		}
		if ids := targetIDs(t, m); len(ids) == 1 && ids[0] == target {
			dispatchMove(t, g, m)
			return
		}
	}
	t.Fatalf("the opponent was offered no Murder at %s", target)
}

// playOutTheStack plays priority and prompts until the stack is empty: the
// bot by the policy, the opponent by passing. It returns every non-pass
// move the bot made.
func playOutTheStack(t *testing.T, g *game.Game, me *game.Player, pol aiseat.Policy) []legal.Move {
	t.Helper()
	var made []legal.Move
	for step := 0; step < 100; step++ {
		var seat uuid.UUID
		empty := false
		g.ReadSnapshot(func() {
			for _, c := range g.PendingChoices {
				if c != nil {
					seat = c.Chooser
					break
				}
			}
			empty = len(g.Stack.Cards) == 0 && len(g.StackMeta) == 0
		})
		if seat == uuid.Nil {
			if empty {
				return made
			}
			holder := g.Turn.PriorityHolder
			if holder < 0 || holder >= len(g.Seats) {
				holder = g.Turn.ActiveSeat
			}
			seat = g.Seats[holder].ID
		}
		moves := legal.EnumerateFor(g, seat)
		if len(moves) == 0 {
			t.Fatalf("step %d: %s was offered nothing", step, seat)
		}
		var mv legal.Move
		if seat != me.ID {
			pi := aiseat.PassIndex(moves)
			if pi < 0 {
				t.Fatalf("step %d: the opponent owes something other than a pass: %v", step, moves[0].Label)
			}
			mv = moves[pi]
		} else {
			d, err := pol.Decide(context.Background(), aiseat.Input{
				View: protocol.ViewOfGameFor(g, seat.String()), Seat: seat, Moves: moves,
			})
			if err != nil {
				t.Fatalf("step %d: policy: %v", step, err)
			}
			if d.Index == aiseat.Decline {
				if d.Index = aiseat.PassIndex(moves); d.Index < 0 {
					t.Fatalf("step %d: the bot declined with no pass on offer", step)
				}
			}
			mv = moves[d.Index]
			if mv.Kind != legal.KindPass {
				t.Logf("%s: %s — %s", g.Turn.Step, mv.Label, d.Reason)
				made = append(made, mv)
			}
		}
		dispatchMove(t, g, mv)
	}
	t.Fatalf("the stack never emptied; the bot made %d moves", len(made))
	return nil
}

func ogreOnBattlefield(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				found = true
			}
		}
	})
	return found
}

// TestHeuristicSavesItsCreatureFromMurderWithMotherOfRunes: Murder at the
// bot's 4/4, Mother of Runes untapped. The bot activates Mother of Runes
// on the 4/4, names black, and the 4/4 survives.
func TestHeuristicSavesItsCreatureFromMurderWithMotherOfRunes(t *testing.T) {
	g, me, them, mother, ogre, murder := murderTable(t, 4312)
	castMurderAt(t, g, them, murder, ogre)

	made := playOutTheStack(t, g, me, heuristic.New())
	activated := false
	for _, m := range made {
		if m.Kind == legal.KindActivate && m.Source == mother {
			activated = true
			if ids := targetIDs(t, m); len(ids) != 1 || ids[0] != ogre {
				t.Errorf("Mother of Runes was aimed at %v, want the Murder's target %s", ids, ogre)
			}
		}
	}
	if !activated {
		t.Fatalf("the bot let Murder resolve with Mother of Runes untapped (#4312); it made %d moves", len(made))
	}
	if !ogreOnBattlefield(g, ogre) {
		t.Errorf("the 4/4 died: the protection did not stop the Murder (was black named?)")
	}
}

// TestHeuristicWithoutGrantSavesLetsMurderResolve: the same table with
// PriceGrantSaves off, the pricing #4302 left. The grant buys nothing on
// the opponent's turn and the creature dies: what the switch changes.
func TestHeuristicWithoutGrantSavesLetsMurderResolve(t *testing.T) {
	g, me, them, mother, ogre, murder := murderTable(t, 4313)
	castMurderAt(t, g, them, murder, ogre)

	cfg := heuristic.DefaultConfig()
	cfg.PriceGrantSaves = false
	made := playOutTheStack(t, g, me, heuristic.NewWithConfig(cfg))
	for _, m := range made {
		if m.Kind == legal.KindActivate && m.Source == mother {
			t.Errorf("with PriceGrantSaves off the bot still answered with Mother of Runes (%q)", m.Label)
		}
	}
	if ogreOnBattlefield(g, ogre) {
		t.Errorf("the 4/4 survived an unanswered Murder")
	}
}
