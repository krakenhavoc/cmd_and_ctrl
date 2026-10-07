package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// combat_relative_targets_test.go — #1863. "Target creature it's
// blocking" is one set everywhere: the enumerator the bot picks from,
// the legal_targets the client's picker highlights, and the engine's
// announce check all read TargetSpec.CombatWithSource through the same
// walk.

const oracleGoblinSnowman = "38e9ad30-4bbf-4b58-8bb2-47520ce351a3"

func TestGoblinSnowmanEnumeratorAndViewOfferOnlyTheCreatureItBlocks(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	snowman := battlefieldCard(g, active, game.Card{
		Name: "Goblin Snowman", TypeLine: "Creature — Goblin", OracleID: oracleGoblinSnowman,
		Power: 1, Toughness: 1,
	})
	blocked := battlefieldCard(g, opp, creature("Blocked", "{G}", 2, 2))
	free := battlefieldCard(g, opp, creature("Free", "{1}{G}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)
	for i := range g.Battlefield.Cards {
		switch g.Battlefield.Cards[i].InstanceID {
		case blocked, free:
			g.Battlefield.Cards[i].AttackingTarget = active.ID
		case snowman:
			g.Battlefield.Cards[i].BlockingTarget = blocked
		}
	}

	moves := legal.EnumerateFor(g, active.ID)
	seen := boundedTargetsOffered(t, moves, snowman, map[string]bool{blocked.String(): true})
	if !seen[blocked.String()] {
		t.Errorf("the creature the Snowman blocks was never offered: %v", labels(moves))
	}
	if seen[free.String()] || seen[snowman.String()] {
		t.Errorf("a creature the Snowman is not blocking was offered: %v", seen)
	}
	dispatchAll(t, g, active.ID, moves)

	v := protocol.ViewOfGameFor(g, active.ID.String())
	var cards []string
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID != snowman.String() {
			continue
		}
		for _, ab := range c.ActivatedAbilities {
			if ab.LegalTargets != nil && len(ab.LegalTargets.Cards) > 0 {
				cards = ab.LegalTargets.Cards
			}
		}
	}
	if len(cards) != 1 || cards[0] != blocked.String() {
		t.Errorf("view legal_targets = %v, want only %s", cards, blocked)
	}

	// Out of combat, the same walk offers nothing.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == snowman {
			g.Battlefield.Cards[i].BlockingTarget = uuid.Nil
		}
	}
	for _, m := range activationsOf(legal.EnumerateFor(g, active.ID), snowman) {
		if len(boundTargetIDs(t, m)) > 0 {
			t.Errorf("%q still offers targets once the Snowman blocks nothing", m.Label)
		}
	}
}
