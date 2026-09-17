package game

import (
	"testing"

	"github.com/google/uuid"
)

// TestManaValueForEffectCountsXOnlyOnTheStack pins the single read
// card files use for a mana value that might belong to a spell (#788).
// CR 202.3e: {X} is the value chosen for it while the spell is on the
// stack and zero everywhere else. A copy on the stack has the X of
// the spell it copies (CR 707.10). Each row is one place a card can
// be when something asks for its mana value.
func TestManaValueForEffectCountsXOnlyOnTheStack(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]

	cast := func(name, typeLine, cost string, x int) uuid.UUID {
		t.Helper()
		id := pushTypedCardToHandWithCost(p, name, typeLine, cost)
		if err := g.CastSpell(p.ID, id, CastSpellParams{XValue: x, HoldPriority: true}); err != nil {
			t.Fatalf("CastSpell %s: %v", name, err)
		}
		return id
	}

	// A permanent spell cast with X=5, resolved: on the battlefield
	// its X is zero again. A noncreature artifact, so no 0/0 dies.
	endless := cast("Astral Cornucopia", "Artifact", "{X}{X}{X}", 5)
	resolveTop(t, g)
	if !g.Battlefield.Contains(endless) {
		t.Fatal("the artifact did not resolve onto the battlefield")
	}

	// Two spells left on the stack, and a copy of the top one.
	ballista := cast("Walking Ballista", "Artifact Creature — Construct", "{X}{X}", 4)
	blaze := cast("Blaze", "Instant", "{X}{R}", 3)
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(blaze, p.ID, false); err != nil {
			t.Fatalf("CopySpellForEffect: %v", err)
		}
	})
	copyID := g.Stack.Cards[len(g.Stack.Cards)-1].InstanceID
	if copyID == blaze {
		t.Fatal("the copy is not on top of the stack")
	}

	// A card in hand with an X cost, and one in the graveyard that
	// still has a spell entry in StackMeta. That entry is left over
	// and must not make X count.
	inHand := pushTypedCardToHandWithCost(p, "Fireball", "Sorcery", "{X}{R}")
	stale := NewCard("Stale Hydra", p.ID)
	stale.TypeLine = "Creature — Hydra"
	stale.ManaCost = "{X}{G}{G}"
	p.Graveyard.PushTop(stale)
	g.StackMeta[stale.InstanceID] = &StackItem{ID: stale.InstanceID, Kind: StackItemSpell, XValue: 7}
	unreadable := pushTypedCardToHandWithCost(p, "Fire // Ice", "Instant // Instant", "{1}{R} // {1}{U}")

	rows := []struct {
		name   string
		id     uuid.UUID
		wantMV int
		wantOK bool
	}{
		{"on the stack, {X}{X} with X=4", ballista, 8, true},
		{"on the stack, {X}{R} with X=3", blaze, 4, true},
		{"a copy on the stack keeps the copied X (CR 707.10)", copyID, 4, true},
		{"on the battlefield, cast with X=5", endless, 0, true},
		{"in hand, X is zero", inHand, 1, true},
		{"in the graveyard with a leftover stack entry, X is zero", stale.InstanceID, 2, true},
		{"an unreadable cost is not ok", unreadable, 0, false},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			c, ok := g.LookupCardForEffect(row.id)
			if !ok {
				t.Fatalf("card %s is in no zone", row.id)
			}
			mv, ok := g.ManaValueForEffect(c)
			if mv != row.wantMV || ok != row.wantOK {
				t.Errorf("ManaValueForEffect = %d, %v; want %d, %v", mv, ok, row.wantMV, row.wantOK)
			}
			// Off the stack it agrees with the printed read.
			if row.id != ballista && row.id != blaze && row.id != copyID {
				pmv, pok := c.ParsedManaValue()
				if pmv != mv || pok != ok {
					t.Errorf("off the stack: ParsedManaValue = %d, %v, ManaValueForEffect = %d, %v", pmv, pok, mv, ok)
				}
			}
		})
	}

	// A nil game has no stack, so X is zero.
	var none *Game
	if mv, ok := none.ManaValueForEffect(Card{ManaCost: "{X}{X}"}); mv != 0 || !ok {
		t.Errorf("nil game: ManaValueForEffect = %d, %v; want 0, true", mv, ok)
	}
}
