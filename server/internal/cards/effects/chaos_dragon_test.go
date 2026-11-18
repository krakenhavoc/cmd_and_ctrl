package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// chaos_dragon_test.go — #5897: Chaos Dragon's "each player rolls a d20;
// it can't attack the opponents with the highest result or their
// planeswalkers this combat".

const chaosDragonOracle = "a04f46b4-b0b0-4aef-907f-1df1a2e261de"

// chaosDragonTable is a four-seat table with Chaos Dragon under seat 0, each seat
// but seat 0 with a planeswalker, at seat 0's beginning of combat with the
// Dragon's trigger on the stack and not yet resolved.
type chaosDragonTable struct {
	g       *game.Game
	dragon  uuid.UUID
	walkers map[uuid.UUID]uuid.UUID // controller -> planeswalker
}

func chaosDragonSetup(t *testing.T, key [32]byte, eliminate ...int) chaosDragonTable {
	t.Helper()
	g := newCatalogGame(t)
	g.SetRNGKeyForTest(key)
	for _, i := range eliminate {
		g.Seats[i].Eliminated = true
	}
	me := g.Seats[0]
	dragon := pushTestPermanent(g, me.ID, game.Card{Name: "Chaos Dragon", OracleID: chaosDragonOracle,
		TypeLine: "Creature — Dragon", Power: 4, Toughness: 4})
	walkers := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats[1:] {
		w := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: w, Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
			Owner: p.ID, Controller: p.ID, Counters: map[string]int{game.CounterLoyalty: 3},
		})
		walkers[p.ID] = w
	}
	advanceTo(t, g, game.StepBeginCombat)
	if stackFullyEmpty(g) {
		t.Fatal("the beginning-of-combat trigger is not on the stack")
	}
	return chaosDragonTable{g: g, dragon: dragon, walkers: walkers}
}

// chaosDragonPeek is each live player's d20 for the Dragon's trigger this turn, read
// on a clone so the table's own draws do not move.
func chaosDragonPeek(tb chaosDragonTable) map[uuid.UUID]int {
	c := tb.g.Clone()
	out := map[uuid.UUID]int{}
	c.WithWriteLock(func() {
		for _, p := range c.Seats {
			if p.Eliminated {
				continue
			}
			r, _ := c.RollDiceForEffect(game.RandomDraw{Player: p.ID, Source: tb.dragon}, 20, 1)
			out[p.ID] = r[0]
		}
	})
	return out
}

// chaosDragonFind sets up tables under keys 0..255 until the players' rolls (seat
// order) satisfy want.
func chaosDragonFind(t *testing.T, want func(r []int) bool, eliminate ...int) (chaosDragonTable, map[uuid.UUID]int) {
	t.Helper()
	for k := 0; k < 256; k++ {
		tb := chaosDragonSetup(t, randomModesKey(byte(k)), eliminate...)
		rolls := chaosDragonPeek(tb)
		var seq []int
		for _, p := range tb.g.Seats {
			if !p.Eliminated {
				seq = append(seq, rolls[p.ID])
			}
		}
		if want(seq) {
			return tb, rolls
		}
	}
	t.Fatal("no key in 0..255 gives the wanted rolls")
	return chaosDragonTable{}, nil
}

// chaosDragonAttackable reports, for each opponent and their planeswalker, whether
// the engine lets the Dragon attack it (on a clone) and whether the
// enumerator the bot plays from offers it.
func chaosDragonAttackable(t *testing.T, tb chaosDragonTable) (engine, offered map[uuid.UUID]bool) {
	t.Helper()
	g := tb.g
	me := g.Seats[0].ID
	engine, offered = map[uuid.UUID]bool{}, map[uuid.UUID]bool{}
	for _, m := range legal.EnumerateFor(g, me) {
		if m.Kind != legal.KindAttack || m.Source != tb.dragon {
			continue
		}
		var p struct {
			Target string `json:"target"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("attack params %s: %v", m.Params, err)
		}
		offered[uuid.MustParse(p.Target)] = true
	}
	for _, p := range g.Seats[1:] {
		if p.Eliminated {
			continue
		}
		for _, target := range []uuid.UUID{p.ID, tb.walkers[p.ID]} {
			err := declareResult(g, tb.dragon, target)
			if err != nil && !errors.Is(err, game.ErrIllegalAttackTarget) {
				t.Fatalf("declare at %s: %v", target, err)
			}
			engine[target] = err == nil
		}
	}
	return engine, offered
}

// chaosDragonCheck asserts that exactly the players in `protected` (and their
// planeswalkers) are closed to the Dragon, in the engine and the
// enumerator alike.
func chaosDragonCheck(t *testing.T, tb chaosDragonTable, protected ...uuid.UUID) {
	t.Helper()
	shut := map[uuid.UUID]bool{}
	for _, p := range protected {
		shut[p], shut[tb.walkers[p]] = true, true
	}
	engine, offered := chaosDragonAttackable(t, tb)
	for target, ok := range engine {
		if ok == shut[target] {
			t.Errorf("attack at %s accepted=%v, want %v", target, ok, !shut[target])
		}
		if offered[target] != ok {
			t.Errorf("enumerator offers %s=%v, engine says %v", target, offered[target], ok)
		}
	}
}

// chaosDragonRollsAnnounced is how many d20s the Dragon's trigger has announced.
func chaosDragonRollsAnnounced(tb chaosDragonTable) int {
	n := 0
	tb.g.WithWriteLock(func() {
		for _, ev := range tb.g.Events {
			if ev.Kind == game.EventRollDie && ev.Source == tb.dragon {
				n++
			}
		}
	})
	return n
}

func chaosDragonResolveAndAttack(t *testing.T, tb chaosDragonTable) {
	t.Helper()
	passPriorityAroundTable(t, tb.g)
	advanceTo(t, tb.g, game.StepDeclareAttackers)
}

// One opponent alone has the highest result: the Dragon can't attack that
// player or their planeswalker, can attack every other opponent and
// planeswalker, and "attacks each combat if able" still binds it.
func TestChaosDragonCantAttackTheOpponentWithTheHighestRoll(t *testing.T) {
	tb, rolls := chaosDragonFind(t, func(r []int) bool { return r[1] > r[0] && r[1] > r[2] && r[1] > r[3] })
	chaosDragonResolveAndAttack(t, tb)
	if n := chaosDragonRollsAnnounced(tb); n != 4 {
		t.Fatalf("%d rolls announced, want one per player (%v)", n, rolls)
	}
	g := tb.g
	chaosDragonCheck(t, tb, g.Seats[1].ID)
	g.WithWriteLock(func() {
		if g.AttackRequirementsUnmetForEffect() == nil {
			t.Error("with unprotected opponents left, the Dragon still owes an attack")
		}
	})
	if err := g.DeclareAttacker(tb.dragon, g.Seats[2].ID); err != nil {
		t.Fatalf("attack an unprotected opponent: %v", err)
	}
	lockInAttacks(t, g)
}

// Ties count: every opponent tied for the highest result is protected.
func TestChaosDragonProtectsEveryOpponentTiedForHighest(t *testing.T) {
	tb, _ := chaosDragonFind(t, func(r []int) bool { return r[1] == r[3] && r[1] > r[0] && r[1] > r[2] })
	chaosDragonResolveAndAttack(t, tb)
	chaosDragonCheck(t, tb, tb.g.Seats[1].ID, tb.g.Seats[3].ID)
}

// A tie with the Dragon's controller still protects the opponent; you
// alone at the top protects nobody.
func TestChaosDragonTieWithYouProtectsTheOpponentButYouAloneProtectsNobody(t *testing.T) {
	t.Run("tie", func(t *testing.T) {
		tb, _ := chaosDragonFind(t, func(r []int) bool { return r[0] == r[2] && r[0] > r[1] && r[0] > r[3] })
		chaosDragonResolveAndAttack(t, tb)
		chaosDragonCheck(t, tb, tb.g.Seats[2].ID)
	})
	t.Run("alone", func(t *testing.T) {
		tb, _ := chaosDragonFind(t, func(r []int) bool { return r[0] > r[1] && r[0] > r[2] && r[0] > r[3] })
		chaosDragonResolveAndAttack(t, tb)
		chaosDragonCheck(t, tb)
		tb.g.WithWriteLock(func() {
			for _, e := range tb.g.ScopedEffects {
				for _, m := range e.Mods {
					if m.Kind == game.ModCantAttackPlayer {
						t.Errorf("a restriction was registered: %+v", e)
					}
				}
			}
		})
	})
}

// Every opponent protected: the requirement gives way (CR 508.1d) and the
// Dragon may stay home.
func TestChaosDragonWithEveryOpponentProtectedNeedNotAttack(t *testing.T) {
	tb, _ := chaosDragonFind(t, func(r []int) bool { return r[1] >= r[0] }, 2, 3)
	chaosDragonResolveAndAttack(t, tb)
	g := tb.g
	chaosDragonCheck(t, tb, g.Seats[1].ID)
	g.WithWriteLock(func() {
		if err := g.AttackRequirementsUnmetForEffect(); err != nil {
			t.Errorf("with nothing it may attack, the Dragon owes no attack: %v", err)
		}
	})
	lockInAttacks(t, g)
}

// The restriction lasts this combat only: once combat ends the Dragon
// carries none, and the record is gone.
func TestChaosDragonRestrictionEndsWithTheCombat(t *testing.T) {
	tb, _ := chaosDragonFind(t, func(r []int) bool { return r[1] >= r[0] }, 2, 3)
	chaosDragonResolveAndAttack(t, tb)
	g := tb.g
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if rs := findBattlefieldCardForTest(g, tb.dragon).Effective().AttackTargetRestrictions; len(rs) != 1 {
			t.Fatalf("during combat the Dragon carries %+v, want one restriction", rs)
		}
	})
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if rs := findBattlefieldCardForTest(g, tb.dragon).Effective().AttackTargetRestrictions; len(rs) != 0 {
			t.Errorf("after combat the Dragon still carries %+v", rs)
		}
	})
}

// It is the resolved trigger's effect, not an ability of the Dragon: a
// later "loses all abilities" leaves it (CR 611.2a, 613.1f).
func TestChaosDragonRestrictionSurvivesLosingAllAbilities(t *testing.T) {
	tb, _ := chaosDragonFind(t, func(r []int) bool { return r[1] > r[0] && r[1] > r[2] && r[1] > r[3] })
	passPriorityAroundTable(t, tb.g)
	g := tb.g
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(tb.dragon),
			[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "loses all abilities")
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	chaosDragonCheck(t, tb, g.Seats[1].ID)
	g.WithWriteLock(func() {
		if err := g.AttackRequirementsUnmetForEffect(); err != nil {
			t.Errorf("with no abilities it no longer attacks each combat: %v", err)
		}
	})
}

// Each player's roll is its own event: Pixie Guide's "one more die and
// ignore the lowest" applies to its controller's roll alone, and the kept
// (higher) die is that player's result.
func TestChaosDragonRollReplacementsApplyPerPlayer(t *testing.T) {
	// Seat 1's two draws: the higher must beat everyone, the first alone
	// must not, so only the replacement makes seat 1 the highest.
	for k := 0; k < 256; k++ {
		tb := chaosDragonSetup(t, randomModesKey(byte(k)))
		g := tb.g
		var two []int
		c := g.Clone()
		c.WithWriteLock(func() {
			two, _ = c.RollDiceForEffect(game.RandomDraw{Player: g.Seats[1].ID, Source: tb.dragon}, 20, 2)
		})
		rolls := chaosDragonPeek(tb)
		hi := max(two[0], two[1])
		beatsAll := func(v int) bool {
			return v > rolls[g.Seats[0].ID] && v > rolls[g.Seats[2].ID] && v > rolls[g.Seats[3].ID]
		}
		if !beatsAll(hi) || beatsAll(two[0]) {
			continue
		}
		p2Push(g, g.Seats[1].ID, "Pixie Guide", "Creature — Faerie", pixieGuideOracle, 1, 3)
		chaosDragonResolveAndAttack(t, tb)
		if n := chaosDragonRollsAnnounced(tb); n != 4 {
			t.Fatalf("%d rolls announced, want one kept die per player", n)
		}
		chaosDragonCheck(t, tb, g.Seats[1].ID)
		return
	}
	t.Fatal("no key in 0..255 gives the wanted rolls")
}

// A replacement that asks a question pauses that player's roll; the
// players after them roll once it is answered, and the chosen roll is the
// one compared.
func TestChaosDragonPausesForADiceQuestionAndCarriesOn(t *testing.T) {
	tb := chaosDragonSetup(t, randomModesKey(7))
	g := tb.g
	me, roller := g.Seats[0], g.Seats[2]
	g.WithWriteLock(func() {
		if !g.RollExtraDieNextTimeForEffect(uuid.Nil, me.ID, roller.ID, "sticker") {
			t.Fatal("no shield registered")
		}
	})
	passPriorityAroundTable(t, g)
	c := sfkPrompt(g, "dice/ignore-roll")
	if c == nil || c.Chooser != me.ID {
		t.Fatalf("the question is %+v, want mine about seat 2's roll", c)
	}
	if n := chaosDragonRollsAnnounced(tb); n != 2 {
		t.Fatalf("%d rolls announced before the answer, want seats 0 and 1 only", n)
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
		t.Fatalf("answer: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := chaosDragonRollsAnnounced(tb); n != 4 {
		t.Fatalf("%d rolls announced, want four", n)
	}
	// The protected set follows the announced results.
	results := map[uuid.UUID]int{}
	g.WithWriteLock(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventRollDie && ev.Source == tb.dragon {
				results[ev.Actor] = ev.RollResult()
			}
		}
	})
	top := 0
	for _, r := range results {
		top = max(top, r)
	}
	var protected []uuid.UUID
	for _, p := range g.Seats[1:] {
		if results[p.ID] == top {
			protected = append(protected, p.ID)
		}
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	chaosDragonCheck(t, tb, protected...)
}
