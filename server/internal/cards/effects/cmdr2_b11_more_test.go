package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func b11CastQuickling(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	q := handCardOf(g, me, "Quickling", "Creature — Faerie Rogue", "{1}{U}", quicklingOracle)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == q {
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 2, 2
				me.Hand.Cards[i].Keywords = []string{"flash", "flying"}
			}
		}
		fillPoolColored(me, "U", 2)
	})
	if err := g.CastSpell(me.ID, q, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Quickling: %v", err)
	}
	passPriorityAroundTable(t, g)
	return q
}

// Quickling with no other creature is sacrificed.
func TestQuicklingSacrificedWithNothingToReturn(t *testing.T) {
	g, me, _ := tapTable(t)
	q := b11CastQuickling(t, g, me)
	if g.Battlefield.Contains(q) {
		t.Fatal("Quickling stayed with no creature returned")
	}
}

// Returning another creature keeps it; declining sacrifices it.
func TestQuicklingReturnsAnotherCreatureOrIsSacrificed(t *testing.T) {
	g, me, _ := tapTable(t)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	q := b11CastQuickling(t, g, me)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(q) || g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Fatal("the Bear should be in hand and the Quickling on the battlefield")
	}

	// Decline.
	g2, me2, _ := tapTable(t)
	bear2 := pr7Creature(g2, me2.ID, "Bear", 2, "G")
	q2 := b11CastQuickling(t, g2, me2)
	pick2 := latestChooseCardsFor(g2, me2.ID)
	if pick2 == nil {
		t.Fatal("no prompt")
	}
	if err := g2.ResolveChooseCards(pick2.ID, me2.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards(none): %v", err)
	}
	passPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(q2) || !g2.Battlefield.Contains(bear2) {
		t.Fatal("declining should sacrifice the Quickling and keep the Bear")
	}
}

// Stolen by the Fae: X is the target's mana value; X Faeries.
func TestStolenByTheFae(t *testing.T) {
	g, me, opp := tapTable(t)
	sf := handCardOf(g, me, "Stolen by the Fae", "Sorcery", "{X}{U}{U}", stolenByTheFaeOracle)
	three := b11PutOnBattlefield(g, opp.ID, "Three Drop", "Creature — Bear", "", 3, 3)
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, three).ManaCost = "{2}{G}"
		fillPoolColored(me, "U", 5)
	})
	// X = 2 does not match a mana value 3 creature.
	if err := g.CastSpell(me.ID, sf, game.CastSpellParams{Strict: true, XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: three}}}); err == nil {
		t.Fatal("X=2 targeted a mana value 3 creature")
	}
	if err := g.CastSpell(me.ID, sf, game.CastSpellParams{Strict: true, XValue: 3, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: three}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(three) || !opp.Hand.Contains(three) {
		t.Error("the target was not returned to its owner's hand")
	}
	if n := b11Count(g, "Faerie"); n != 3 {
		t.Errorf("%d Faerie tokens, want 3", n)
	}
}
