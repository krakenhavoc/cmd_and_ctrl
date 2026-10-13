package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spiritual_focus_test.go — ADR 0153 PR 2 (#3168).

const spiritualFocusOracle = "da350f46-6d6e-4257-8f4f-77769c442d0c"

// Each card an opponent's spell makes its controller discard gains 2
// life and offers a draw: a two-card Mind Rot is 4 life and two offers.
func TestSpiritualFocusGainsLifeAndMayDrawPerDiscardedCard(t *testing.T) {
	d := newDBOTable(t)
	d.permanent(d.victim, "Spiritual Focus", "Enchantment", spiritualFocusOracle)
	a := d.handCard(d.victim, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.handCard(d.victim, "Card B", "Sorcery", "{1}", "", 0, 0)
	life, hand := d.victim.Life, d.victim.Hand.Size()

	castCatalogSpell(t, d.g, "Mind Rot", "Sorcery", mindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
	passPriorityAroundTable(t, d.g)
	answerDiscard(t, d.g, d.victim.ID, a, b)
	passPriorityAroundTable(t, d.g)
	answerMayChoice(t, d.g, d.victim.ID, true)
	passPriorityAroundTable(t, d.g)
	answerMayChoice(t, d.g, d.victim.ID, false)
	passPriorityAroundTable(t, d.g)

	if d.victim.Life != life+4 {
		t.Errorf("life %d -> %d, want +4", life, d.victim.Life)
	}
	if got := d.victim.Hand.Size(); got != hand-2+1 {
		t.Errorf("hand %d -> %d, want one drawn after two discarded", hand, got)
	}
}

// Your own discard is not an opponent's: nothing happens.
func TestSpiritualFocusIgnoresYourOwnDiscard(t *testing.T) {
	d := newDBOTable(t)
	d.permanent(d.caster, "Spiritual Focus", "Enchantment", spiritualFocusOracle)
	a := d.handCard(d.caster, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.handCard(d.caster, "Card B", "Sorcery", "{1}", "", 0, 0)
	life := d.caster.Life

	d.iDiscard(t, a, b)

	if d.caster.Life != life {
		t.Errorf("life %d -> %d, want unchanged", life, d.caster.Life)
	}
	if latestConfirmFor(d.g, d.caster.ID) != nil {
		t.Error("a draw was offered for your own discard")
	}
}

// It is its controller's discards only: an opponent's Spiritual Focus
// does nothing when you are made to discard.
func TestSpiritualFocusIgnoresAnotherPlayersDiscard(t *testing.T) {
	d := newDBOTable(t)
	d.permanent(d.caster, "Spiritual Focus", "Enchantment", spiritualFocusOracle)
	a := d.handCard(d.victim, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.handCard(d.victim, "Card B", "Sorcery", "{1}", "", 0, 0)
	life := d.caster.Life

	d.theyDiscard(t, a, b)

	if d.caster.Life != life {
		t.Errorf("life %d -> %d, want unchanged", life, d.caster.Life)
	}
}

// A discard paid as a cost is not caused by an opponent. A cleanup-step
// discard is not caused by anyone's spell either.
func TestSpiritualFocusIgnoresCostAndCleanupDiscards(t *testing.T) {
	d := newDBOTable(t)
	me := d.caster.ID
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
