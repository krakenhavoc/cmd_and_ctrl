package aiseat_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// heliod_grant_bot_test.go — #4302, whole-engine: the real catalog
// Heliod, Sun-Crowned, the real enumerator and the real heuristic.
//
// Ian's report from real games: with Heliod out, a heuristic seat
// ignored the creature in its hand and its castable commander, spent
// its mana on "{1}{W}: Another target creature gains lifelink until end
// of turn" aimed at an OPPONENT's creature, and did nothing else that
// turn. Two faults: the target (lifelink is a gift to whoever controls
// the creature) and the ranking (a grant that ends with the turn beat
// developing the board, and was taken again for as long as the mana
// lasted).
//
// Not behind AISEAT_GAME_TESTS: one main phase, milliseconds.

const botHeliodOracle = "63e596a2-9126-4af6-8782-c38687d664ad"

// heliodTable parks a two-seat game on the active seat's first main
// phase with Heliod, Sun-Crowned (devotion two, so not a creature) and
// `plains` untapped Plains, an empty hand and an empty command zone,
// facing an opponent's 3/3. The caller adds what the bot may cast.
func heliodTable(t *testing.T, seed uint64, plains int) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newSettledTable(t, seed)
	advanceToStep(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	everyone := seatIDs(g)
	var heliod uuid.UUID
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		me.Command.Cards = nil
		them.Command.Cards = nil
		bear := game.Card{InstanceID: uuid.New(), Name: "Opposing Bear", TypeLine: "Creature — Bear",
			ManaCost: "{2}{G}", Power: 3, Toughness: 3, Owner: them.ID, Controller: them.ID}
		bear.AddKnowersAll(everyone)
		g.Battlefield.PushTop(bear)
		h := game.Card{InstanceID: uuid.New(), Name: "Heliod, Sun-Crowned", OracleID: botHeliodOracle,
			TypeLine: "Legendary Enchantment Creature — God", ManaCost: "{2}{W}", Power: 5, Toughness: 5,
			Owner: me.ID, Controller: me.ID}
		h.AddKnowersAll(everyone)
		g.Battlefield.PushTop(h)
		heliod = h.InstanceID
		for i := 0; i < plains; i++ {
			p := game.Card{InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains",
				Owner: me.ID, Controller: me.ID}
			p.AddKnowersAll(everyone)
			g.Battlefield.PushTop(p)
		}
	})
	// Heliod's "isn't a creature" static reads the board the cards
	// were pushed onto by hand.
	g.BumpLayerVersionForTest()
	return g, me, them, heliod
}

// whiteTwoDrop is a vanilla {1}{W} 2/2: a card with no catalog entry
// and nothing to implement.
func whiteTwoDrop(owner uuid.UUID, name string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Cat",
		ManaCost: "{1}{W}", Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		KnownBy: map[uuid.UUID]bool{owner: true}}
}

// heliodMove is one move the bot made, with the controller of every
// card it targeted.
type heliodMove struct {
	move      legal.Move
	targetsOf []uuid.UUID
	step      game.Step
}

// playHeliodMainPhase plays the bot's first main phase out with the
// heuristic, the opponent passing back whenever it has priority. It
// returns every non-pass move the bot made.
func playHeliodMainPhase(t *testing.T, g *game.Game, me *game.Player, pol aiseat.Policy) []heliodMove {
	t.Helper()
	return playHeliodUntil(t, g, me, pol, func() bool { return g.Turn.Step != game.StepPrecombatMain })
}

// playHeliodTurn plays the bot's turn from here to its end step: its
// main phases, its attack and the priority windows of combat. The
// opponent, whose creature is tapped and cannot block, only passes.
func playHeliodTurn(t *testing.T, g *game.Game, me *game.Player, pol aiseat.Policy) []heliodMove {
	t.Helper()
	turn := g.Turn.Seq
	return playHeliodUntil(t, g, me, pol, func() bool { return g.Turn.Seq != turn || g.Turn.Step == game.StepEnd })
}

// playHeliodUntil is the loop behind both: the seat holding priority
// moves (the active seat while nobody does, in a declaration), the bot
// by the policy and the opponent by passing, until done says stop.
func playHeliodUntil(t *testing.T, g *game.Game, me *game.Player, pol aiseat.Policy, done func() bool) []heliodMove {
	t.Helper()
	var made []heliodMove
	for step := 0; step < 200; step++ {
		if done() || owesAnyChoice(g) {
			return made
		}
		holder := g.Turn.PriorityHolder
		if holder < 0 || holder >= len(g.Seats) {
			holder = g.Turn.ActiveSeat
		}
		seat := g.Seats[holder].ID
		moves := legal.EnumerateFor(g, seat)
		if len(moves) == 0 {
			t.Fatalf("step %d (%s): %s was offered nothing", step, g.Turn.Step, seat)
		}
		var mv legal.Move
		if seat != me.ID {
			pi := aiseat.PassIndex(moves)
			if pi < 0 {
				t.Fatalf("step %d (%s): the opponent has no pass on offer", step, g.Turn.Step)
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
				// The runner turns a decline into a pass when one exists.
				if d.Index = aiseat.PassIndex(moves); d.Index < 0 {
					t.Fatalf("step %d (%s): the bot declined with no pass on offer", step, g.Turn.Step)
				}
			}
			if d.Index < 0 || d.Index >= len(moves) {
				t.Fatalf("step %d: policy picked index %d of %d", step, d.Index, len(moves))
			}
			mv = moves[d.Index]
			if mv.Kind != legal.KindPass {
				t.Logf("%s: %s — %s", g.Turn.Step, mv.Label, d.Reason)
				made = append(made, heliodMove{move: mv, targetsOf: targetControllers(t, g, mv), step: g.Turn.Step})
			}
		}
		dispatchMove(t, g, mv)
	}
	t.Fatalf("the window never closed; stuck at %s after %v", g.Turn.Step, labelsOf(made))
	return nil
}

// targetIDs is every card a move targets.
func targetIDs(t *testing.T, m legal.Move) []uuid.UUID {
	t.Helper()
	var p struct {
		Targets []struct {
			Kind string    `json:"kind"`
			ID   uuid.UUID `json:"id"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", string(m.Params), err)
	}
	var out []uuid.UUID
	for _, tg := range p.Targets {
		if tg.Kind == "card" {
			out = append(out, tg.ID)
		}
	}
	return out
}

// targetControllers is the controller of each card a move targets.
func targetControllers(t *testing.T, g *game.Game, m legal.Move) []uuid.UUID {
	t.Helper()
	var p struct {
		Targets []struct {
			Kind string    `json:"kind"`
			ID   uuid.UUID `json:"id"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", string(m.Params), err)
	}
	var out []uuid.UUID
	for _, tg := range p.Targets {
		if tg.Kind != "card" {
			continue
		}
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == tg.ID {
				out = append(out, c.Controller)
			}
		}
	}
	return out
}

func heliodActivations(made []heliodMove, heliod uuid.UUID) []heliodMove {
	var out []heliodMove
	for _, m := range made {
		if m.move.Kind == legal.KindActivate && m.move.Source == heliod {
			out = append(out, m)
		}
	}
	return out
}

func heliodCastOf(made []heliodMove, id uuid.UUID) bool {
	for _, m := range made {
		if m.move.Kind == legal.KindCast && m.move.Source == id {
			return true
		}
	}
	return false
}

// TestHeuristicHeliodCastsTheCreatureOverTheGrant: two Plains, a {1}{W}
// creature in hand, Heliod's {1}{W} grant on offer and only an
// opponent's creature to aim it at. The bot casts the creature.
func TestHeuristicHeliodCastsTheCreatureOverTheGrant(t *testing.T) {
	g, me, _, heliod := heliodTable(t, 4302, 2)
	lion := whiteTwoDrop(me.ID, "Test Lion")
	g.WithWriteLock(func() { me.Hand.PushTop(lion) })

	made := playHeliodMainPhase(t, g, me, heuristic.New())
	for _, a := range heliodActivations(made, heliod) {
		t.Errorf("the bot activated Heliod (%q) instead of developing its board (#4302)", a.move.Label)
	}
	if !heliodCastOf(made, lion.InstanceID) {
		t.Errorf("the bot did not cast the {1}{W} creature it held with two Plains open; it made %v", labelsOf(made))
	}
}

// TestHeuristicHeliodCastsTheCommanderOverTheGrant: the same, with the
// {1}{W} commander in the command zone instead of a card in hand.
func TestHeuristicHeliodCastsTheCommanderOverTheGrant(t *testing.T) {
	g, me, _, heliod := heliodTable(t, 4303, 2)
	cmd := game.NewCommander("Test Commander", me.ID)
	cmd.TypeLine = "Legendary Creature — Human Soldier"
	cmd.ManaCost = "{1}{W}"
	cmd.Power, cmd.Toughness = 2, 2
	cmd.Controller = me.ID
	g.WithWriteLock(func() {
		cmd.AddKnowersAll(seatIDs(g))
		me.Command.PushTop(cmd)
	})

	made := playHeliodMainPhase(t, g, me, heuristic.New())
	for _, a := range heliodActivations(made, heliod) {
		t.Errorf("the bot activated Heliod (%q) instead of casting its commander (#4302)", a.move.Label)
	}
	if !heliodCastOf(made, cmd.InstanceID) {
		t.Errorf("the bot did not cast its {1}{W} commander with two Plains open; it made %v", labelsOf(made))
	}
}

// TestHeuristicHeliodNeverGrantsLifelinkToAnOpponent: six Plains and
// nothing to cast. The only legal target is the opponent's creature, so
// the grant is never worth making: lifelink is the creature's
// controller's gain (CR 702.15b). The bot neither aims it there nor
// spends its mana on it over and over.
func TestHeuristicHeliodNeverGrantsLifelinkToAnOpponent(t *testing.T) {
	g, me, them, heliod := heliodTable(t, 4304, 6)

	made := playHeliodMainPhase(t, g, me, heuristic.New())
	acts := heliodActivations(made, heliod)
	for _, a := range acts {
		for _, c := range a.targetsOf {
			if c == them.ID {
				t.Errorf("the bot gave an opponent's creature lifelink: %q (#4302)", a.move.Label)
			}
		}
	}
	if len(acts) > 0 {
		t.Errorf("the bot activated Heliod %d times with nothing of its own to target (#4302)", len(acts))
	}
}

func labelsOf(made []heliodMove) []string {
	out := make([]string, 0, len(made))
	for _, m := range made {
		out = append(out, m.move.Label)
	}
	return out
}

// heliodCombatTable is heliodTable with the opponent's creature tapped,
// so nothing blocks, and a 2/2 of the bot's own that can attack.
func heliodCombatTable(t *testing.T, seed uint64, plains int) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g, me, them, heliod := heliodTable(t, seed, plains)
	var mine uuid.UUID
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Controller == them.ID {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
		c := game.Card{InstanceID: uuid.New(), Name: "Veteran Soldier", TypeLine: "Creature — Human Soldier",
			ManaCost: "{1}{W}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
		c.AddKnowersAll(seatIDs(g))
		g.Battlefield.PushTop(c)
		mine = c.InstanceID
	})
	g.BumpLayerVersionForTest()
	return g, me, them, heliod, mine
}

// TestHeuristicHeliodGrantsLifelinkToItsOwnAttacker: four Plains, nothing
// in hand, the bot's 2/2 able to attack into an open board. Lifelink is
// worth having on a creature that is about to deal damage, and the mana
// has nothing else to buy, so the bot attacks and gives its attacker
// lifelink, once: a second instance adds nothing (CR 702.15f).
func TestHeuristicHeliodGrantsLifelinkToItsOwnAttacker(t *testing.T) {
	g, me, them, heliod, mine := heliodCombatTable(t, 4305, 4)
	life := me.Life

	made := playHeliodTurn(t, g, me, heuristic.New())
	acts := heliodActivations(made, heliod)
	if len(acts) != 1 {
		t.Fatalf("the bot activated Heliod %d times, want once on its attacker; it made %v", len(acts), labelsOf(made))
	}
	a := acts[0]
	if a.step != game.StepDeclareAttackers && a.step != game.StepDeclareBlockers {
		t.Errorf("the bot granted lifelink in %s, want once its creature was attacking", a.step)
	}
	if ids := targetIDs(t, a.move); len(ids) != 1 || ids[0] != mine {
		t.Errorf("the bot aimed the lifelink at %v, want its own attacker %s", ids, mine)
	}
	for _, c := range a.targetsOf {
		if c == them.ID {
			t.Errorf("the bot gave an opponent's creature lifelink: %q", a.move.Label)
		}
	}
	if me.Life <= life {
		t.Errorf("life %d after a 2/2 lifelink attack into an open board, want more than %d", me.Life, life)
	}
}

// TestHeuristicHeliodKeepsCombatManaForItsMainPhaseCast: the same attack
// with two Plains and a {1}{W} creature in hand that arrives after the
// first main phase. The grant would spend the mana the second main phase
// needs, so the bot attacks without it and casts the creature after
// combat.
func TestHeuristicHeliodKeepsCombatManaForItsMainPhaseCast(t *testing.T) {
	g, me, _, heliod, _ := heliodCombatTable(t, 4306, 2)
	advanceToStep(t, g, game.StepBeginCombat)
	lion := whiteTwoDrop(me.ID, "Test Lion")
	g.WithWriteLock(func() { me.Hand.PushTop(lion) })

	made := playHeliodTurn(t, g, me, heuristic.New())
	for _, a := range heliodActivations(made, heliod) {
		t.Errorf("the bot spent its main-phase mana on Heliod in %s (%q)", a.step, a.move.Label)
	}
	if !heliodCastOf(made, lion.InstanceID) {
		t.Errorf("the bot did not cast its creature after combat; it made %v", labelsOf(made))
	}
}
