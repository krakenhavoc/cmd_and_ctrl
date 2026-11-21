package testsso

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spiritual_focus_test.go — ADR 0153 PR 2 (#3168).

const spiritualFocusOracle = "da350f46-6d6e-4257-8f4f-77769c442d0c"

// Each card an opponent's spell makes its controller discard gains 2
// life and offers a draw: a two-card Mind Rot is 4 life and two offers.
func TestSpiritualFocusGainsLifeAndMayDrawPerDiscardedCard(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	d.Permanent(d.Victim, "Spiritual Focus", "Enchantment", spiritualFocusOracle)
	a := d.HandCard(d.Victim, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.HandCard(d.Victim, "Card B", "Sorcery", "{1}", "", 0, 0)
	life, hand := d.Victim.Life, d.Victim.Hand.Size()

	effectstest.CastCatalogSpell(t, d.G, "Mind Rot", "Sorcery", effectstest.MindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.Victim.ID}})
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.AnswerDiscard(t, d.G, d.Victim.ID, a, b)
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.AnswerMayChoice(t, d.G, d.Victim.ID, true)
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.AnswerMayChoice(t, d.G, d.Victim.ID, false)
	effectstest.PassPriorityAroundTable(t, d.G)

	if d.Victim.Life != life+4 {
		t.Errorf("life %d -> %d, want +4", life, d.Victim.Life)
	}
	if got := d.Victim.Hand.Size(); got != hand-2+1 {
		t.Errorf("hand %d -> %d, want one drawn after two discarded", hand, got)
	}
}

// Your own discard is not an opponent's: nothing happens.
func TestSpiritualFocusIgnoresYourOwnDiscard(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	d.Permanent(d.Caster, "Spiritual Focus", "Enchantment", spiritualFocusOracle)
	a := d.HandCard(d.Caster, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.HandCard(d.Caster, "Card B", "Sorcery", "{1}", "", 0, 0)
	life := d.Caster.Life

	d.IDiscard(t, a, b)

	if d.Caster.Life != life {
		t.Errorf("life %d -> %d, want unchanged", life, d.Caster.Life)
	}
	if effectstest.LatestConfirmFor(d.G, d.Caster.ID) != nil {
		t.Error("a draw was offered for your own discard")
	}
}

// It is its controller's discards only: an opponent's Spiritual Focus
// does nothing when you are made to discard.
func TestSpiritualFocusIgnoresAnotherPlayersDiscard(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	d.Permanent(d.Caster, "Spiritual Focus", "Enchantment", spiritualFocusOracle)
	a := d.HandCard(d.Victim, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.HandCard(d.Victim, "Card B", "Sorcery", "{1}", "", 0, 0)
	life := d.Caster.Life

	d.TheyDiscard(t, a, b)

	if d.Caster.Life != life {
		t.Errorf("life %d -> %d, want unchanged", life, d.Caster.Life)
	}
}

// A discard paid as a cost is not caused by an opponent. A cleanup-step
// discard is not caused by anyone's spell either.
func TestSpiritualFocusIgnoresCostAndCleanupDiscards(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	me := d.Caster.ID
	for _, cause := range []game.DiscardCause{game.DiscardCauseCost, game.DiscardCauseCleanup} {
		ev := game.Event{Kind: game.EventDiscardCard, Actor: me, Cause: causeForTest(cause), CauseController: me}
		if game.DiscardCausedByOpponentOf(ev, me) {
			t.Errorf("a %v discard reads as an opponent's", cause)
		}
	}
}

func causeForTest(c game.DiscardCause) game.MoveCauseKind {
	if c == game.DiscardCauseCost {
		return game.MoveCauseCost
	}
	return game.MoveCauseRule
}
