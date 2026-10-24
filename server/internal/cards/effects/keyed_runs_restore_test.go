package effects

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_runs_restore_test.go — #4290: the multi-seat prompted runs are
// keyed, so a table waiting on any leg is a restore point. Each test
// answers the first leg, takes a STRICT restore through JSON with the
// next still open (restoreRoundTrip with strict), and finishes on the
// restored table.

// Syphon Mind: one opponent has discarded, one is still choosing; the
// caster draws once, for both, after the last answer on the restored
// table.
func TestSyphonMindRunIsARestorePointWithALegAnswered(t *testing.T) {
	g := newCatalogGame(t)
	emptyHandToLibrary(g, g.Seats[3])
	before := g.Seats[0].Hand.Size()
	castCatalogSpell(t, g, "Syphon Mind", "Sorcery", b09SyphonMindOracle, nil)
	passPriorityAroundTable(t, g)
	if c := discardChoiceFor(g, g.Seats[1].ID); c == nil || c.PickThen != "choose-cards/discard-leg" {
		t.Fatalf("the first opponent's discard is %+v, want a keyed leg", c)
	}
	discardFromHand(t, g, g.Seats[1].ID)
	g = restoreRoundTrip(t, g, true)
	if got := g.Seats[0].Hand.Size(); got != before {
		t.Fatalf("the caster drew before the last opponent answered: hand %d → %d", before, got)
	}
	discardFromHand(t, g, g.Seats[2].ID)
	if got := g.Seats[0].Hand.Size(); got != before+2 {
		t.Errorf("hand %d → %d on the restored table, want +2", before, got)
	}
}

// Vicious Rumors: each opponent's mill is the run's per-seat tail and
// the life gain its continuation; both hold across a restore taken
// between two opponents' answers.
func TestViciousRumorsRunIsARestorePointWithALegAnswered(t *testing.T) {
	g := newCatalogGame(t)
	life := g.Seats[0].Life
	castCatalogSpell(t, g, "Vicious Rumors", "Sorcery", b42ViciousRumorsOracle, nil)
	passPriorityAroundTable(t, g)
	first := g.Seats[1]
	grave := len(first.Graveyard.Cards)
	discardFromHand(t, g, first.ID)
	if got := len(first.Graveyard.Cards); got != grave+2 {
		t.Fatalf("the first opponent discards then mills: graveyard %d → %d, want +2", grave, got)
	}
	g = restoreRoundTrip(t, g, true)
	for i, p := range g.Seats[2:] {
		grave := len(p.Graveyard.Cards)
		if g.Seats[0].Life != life {
			t.Fatalf("you gained life with %s still choosing", p.Name)
		}
		discardFromHand(t, g, p.ID)
		if got := len(p.Graveyard.Cards); got != grave+2 {
			t.Errorf("seat %d discards then mills on the restored table: graveyard %d → %d, want +2", i+2, grave, got)
		}
	}
	if g.Seats[0].Life != life+1 {
		t.Errorf("your life is %d after every opponent answered, want %d", g.Seats[0].Life, life+1)
	}
}

// Tasigur's Cruelty: the fire-and-forget fan-out's legs are keyed with no
// run, so the table restores with one opponent answered.
func TestTasigursCrueltyFanOutIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Tasigur's Cruelty", "Sorcery", tasigursCrueltyOracle, nil)
	passPriorityAroundTable(t, g)
	first := g.Seats[1]
	hand := first.Hand.Size()
	discardFromHand(t, g, first.ID)
	if first.Hand.Size() != hand-2 {
		t.Fatalf("the first opponent's hand %d → %d, want -2", hand, first.Hand.Size())
	}
	g = restoreRoundTrip(t, g, true)
	for _, p := range g.Seats[2:] {
		hand := p.Hand.Size()
		discardFromHand(t, g, p.ID)
		if p.Hand.Size() != hand-2 {
			t.Errorf("%s's hand %d → %d on the restored table, want -2", p.Name, hand, p.Hand.Size())
		}
	}
}

// The run's key is refused by a binary that lacks it.
func TestSyphonMindRunWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Syphon Mind", "Sorcery", b09SyphonMindOracle, nil)
	passPriorityAroundTable(t, g)
	snap := g.CaptureSnapshot()
	if len(snap.PromptRuns) != 1 || snap.PromptRuns[0].Key != syphonMindDraw.Key() {
		t.Fatalf("the snapshot's runs are %+v, want Syphon Mind's", snap.PromptRuns)
	}
	snap.PromptRuns[0].Key = "prompt-run/never-registered"
	if _, err := snap.Restore(); !errors.Is(err, game.ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown run key: err = %v, want ErrUnknownEffectKey", err)
	}
	refuseUnknownKey(t, g, "choose-cards/discard-leg")
}

// Consuming Vapors: the victim's sacrifice is a keyed run's leg, so the
// table restores with it open and the caster gains the creature's
// toughness on the restored table.
func TestConsumingVaporsSacrificeIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, victim.ID, "Bear", 2, 3)
	pushVanillaCreature(g, victim.ID, "Other Bear", 2, 2)
	life := me.Life
	castCatalogSpell(t, g, "Consuming Vapors", "Sorcery", consumingVaporsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	answerSacrifice(t, g, victim.ID, bear)
	if got := g.Seats[0].Life; got != life+3 {
		t.Errorf("your life is %d after the sacrifice on the restored table, want %d", got, life+3)
	}
}

// The sacrifice leg names its run's key, so an older binary refuses it.
func TestConsumingVaporsSacrificeWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	pushVanillaCreature(g, victim.ID, "Bear", 2, 3)
	castCatalogSpell(t, g, "Consuming Vapors", "Sorcery", consumingVaporsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	passPriorityAroundTable(t, g)
	refuseUnknownKey(t, g, consumingVaporsGain.Key())
}
