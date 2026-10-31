package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sneak_copy_test.go — #4646: a copy of a permanent spell copies the
// alternative cost paid for the original (CR 707.10), so a copy of a
// sneaked creature spell becomes a token that enters tapped and attacking
// the same defender (CR 702.190b, 707.10f).

const (
	sneakCopyOracle = "test-sneak-copy-4646"
	sneakCopyName   = "Sneak Copy Rat"
)

// registerSneakCopyCreature registers a vanilla 2/2 with sneak {B}, so the
// token entering needs no target or trigger of its own.
func registerSneakCopyCreature(t *testing.T) {
	t.Helper()
	registerForTest(t, Spec{
		OracleID: sneakCopyOracle, Name: sneakCopyName, Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{B}")},
	})
}

func pushSneakCopyCreatureToHand(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: sneakCopyName, TypeLine: "Creature — Rat Ninja", ManaCost: "{2}{B}",
			OracleID: sneakCopyOracle, Power: 2, Toughness: 2,
			Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true},
		})
	})
	return id
}

// copyTheSpell puts a copy of the spell `spell` on the stack under `me`.
func copyTheSpell(t *testing.T, g *game.Game, me *game.Player, spell uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(spell, me.ID, false, nil); err != nil {
			t.Fatalf("copy: %v", err)
		}
	})
}

// copyTokens are the tokens named sneakCopyName on the battlefield.
func copyTokens(g *game.Game) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == sneakCopyName {
			out = append(out, c)
		}
	}
	return out
}

// The headline: copying a sneaked creature spell makes a token that
// enters tapped and attacking the player the returned creature attacked.
func TestACopyOfASneakedCreatureSpellEntersTappedAndAttacking(t *testing.T) {
	registerSneakCopyCreature(t)
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, attacker, _, _ := sneakBoard(t, g, uuid.Nil)
	rat := pushSneakCopyCreatureToHand(g, me)

	if err := sneakCast(g, me, rat, attacker); err != nil {
		t.Fatalf("sneak cast: %v", err)
	}
	if !me.Hand.Contains(attacker) {
		t.Fatal("the unblocked attacker was not returned as the cost")
	}
	copyTheSpell(t, g, me, rat)
	assistCardSettle(t, g, me.ID, nil, true)

	toks := copyTokens(g)
	if len(toks) != 1 {
		t.Fatalf("%d copy tokens on the battlefield, want 1", len(toks))
	}
	if !toks[0].Tapped {
		t.Error("the copy's token entered untapped (CR 702.190b)")
	}
	if toks[0].AttackingTarget != opp.ID {
		t.Errorf("the copy's token attacks %s, want %s, the player the returned creature attacked", toks[0].AttackingTarget, opp.ID)
	}
	if n := attackEventsFor(g, toks[0].InstanceID); n != 0 {
		t.Errorf("%d EventAttack for the token, want 0: it was never declared (CR 506.3c)", n)
	}
	// And the original still enters the same way.
	c, ok := g.LookupCardForEffect(rat)
	if !ok || !g.Battlefield.Contains(rat) || !c.Tapped || c.AttackingTarget != opp.ID {
		t.Errorf("the original: on battlefield=%v tapped=%v attacking %s, want tapped and attacking %s", g.Battlefield.Contains(rat), c.Tapped, c.AttackingTarget, opp.ID)
	}
}

// The same defender when it is a planeswalker.
func TestACopyOfASneakedCreatureSpellAttacksTheSamePlaneswalker(t *testing.T) {
	registerSneakCopyCreature(t)
	g := newCatalogGame(t)
	walker := pushWalkerForTest(g, g.Seats[1].ID, "Test Walker", "", 5)
	_, attacker, me, _ := sneakBoard(t, g, walker)
	rat := pushSneakCopyCreatureToHand(g, me)

	if err := sneakCast(g, me, rat, attacker); err != nil {
		t.Fatalf("sneak cast: %v", err)
	}
	copyTheSpell(t, g, me, rat)
	assistCardSettle(t, g, me.ID, nil, true)

	toks := copyTokens(g)
	if len(toks) != 1 || !toks[0].Tapped || toks[0].AttackingTarget != walker {
		t.Fatalf("%d copy tokens (first tapped=%v), want one tapped and attacking the walker %s", len(toks), len(toks) > 0 && toks[0].Tapped, walker)
	}
}

// A copy of a copy of a sneaked spell is still a sneaked spell.
func TestACopyOfACopyOfASneakedSpellEntersAttackingToo(t *testing.T) {
	registerSneakCopyCreature(t)
	g := newCatalogGame(t)
	_, attacker, me, opp := sneakBoard(t, g, uuid.Nil)
	rat := pushSneakCopyCreatureToHand(g, me)
	if err := sneakCast(g, me, rat, attacker); err != nil {
		t.Fatalf("sneak cast: %v", err)
	}
	copyTheSpell(t, g, me, rat)
	var first uuid.UUID
	for id, item := range g.StackMeta {
		if item != nil && item.IsCopy {
			first = id
		}
	}
	if first == uuid.Nil {
		t.Fatal("no copy on the stack")
	}
	copyTheSpell(t, g, me, first)
	assistCardSettle(t, g, me.ID, nil, true)

	toks := copyTokens(g)
	if len(toks) != 2 {
		t.Fatalf("%d copy tokens, want 2", len(toks))
	}
	for _, tok := range toks {
		if !tok.Tapped || tok.AttackingTarget != opp.ID {
			t.Errorf("token tapped=%v attacking %s, want tapped and attacking %s", tok.Tapped, tok.AttackingTarget, opp.ID)
		}
	}
}

// The negative: a copy of the same card cast for its mana cost becomes an
// ordinary token, untapped and not attacking.
func TestACopyOfAHardCastSpellEntersUntappedAndNotAttacking(t *testing.T) {
	registerSneakCopyCreature(t)
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rat := pushSneakCopyCreatureToHand(g, me)
	floatMana(t, g, me, "{B}{B}{B}")

	if err := g.CastSpell(me.ID, rat, game.CastSpellParams{}); err != nil {
		t.Fatalf("hard cast: %v", err)
	}
	copyTheSpell(t, g, me, rat)
	assistCardSettle(t, g, me.ID, nil, true)

	toks := copyTokens(g)
	if len(toks) != 1 {
		t.Fatalf("%d copy tokens on the battlefield, want 1", len(toks))
	}
	if toks[0].Tapped || toks[0].AttackingTarget != uuid.Nil {
		t.Errorf("token tapped=%v attacking %s, want untapped and not attacking", toks[0].Tapped, toks[0].AttackingTarget)
	}
}

// The returned creature's defender rides the copy's item, so a copy on the
// stack survives a clone (undo) and a snapshot round trip.
func TestASneakedSpellCopyKeepsItsDefenderAcrossUndoAndRestore(t *testing.T) {
	for _, tc := range []struct {
		name    string
		reshape func(t *testing.T, g *game.Game) *game.Game
	}{
		{"clone", func(_ *testing.T, g *game.Game) *game.Game { return g.Clone() }},
		{"snapshot", func(t *testing.T, g *game.Game) *game.Game { return restoreRoundTrip(t, g, true) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			registerSneakCopyCreature(t)
			g := newCatalogGame(t)
			_, attacker, me, opp := sneakBoard(t, g, uuid.Nil)
			rat := pushSneakCopyCreatureToHand(g, me)
			if err := sneakCast(g, me, rat, attacker); err != nil {
				t.Fatalf("sneak cast: %v", err)
			}
			copyTheSpell(t, g, me, rat)
			restored := tc.reshape(t, g)
			assistCardSettle(t, restored, me.ID, nil, true)
			toks := copyTokens(restored)
			if len(toks) != 1 || !toks[0].Tapped || toks[0].AttackingTarget != opp.ID {
				t.Fatalf("%d restored copy tokens (first tapped=%v), want one tapped and attacking %s", len(toks), len(toks) > 0 && toks[0].Tapped, opp.ID)
			}
		})
	}
}
