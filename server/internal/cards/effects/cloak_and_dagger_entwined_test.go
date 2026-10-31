package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cloak_and_dagger_entwined_test.go — #4585: "choose target opponent and
// up to one target creature they control. They reveal their hand. You
// may exile a nonland card from their hand or the chosen creature until
// Cloak and Dagger leave the battlefield."

const cloakAndDaggerOracle = "64132f93-66ac-4794-8c07-a88f9ed0e22b"

// cdTable is the active player, an opponent with a revealed-able hand
// (a Forest, a Lightning Bolt and a Grizzly Bears card) and a creature,
// and a third player with a creature of their own.
type cdTable struct {
	g              *game.Game
	me, opp, other *game.Player
	hand           []uuid.UUID // Forest, Bolt, Bears card
	theirs, spare  uuid.UUID
}

func newCDTable(t *testing.T) *cdTable {
	t.Helper()
	g := newCatalogGame(t)
	x := &cdTable{g: g}
	x.me = g.Seats[g.Turn.ActiveSeat]
	x.opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	x.other = g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	x.hand = revealHand(x.opp, rhForest(), rhBolt(),
		game.Card{Name: "Hand Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	x.theirs = b12Creature(g, x.opp.ID, "Their Hill Giant", "Creature — Giant", 3, 3)
	x.spare = b12Creature(g, x.other.ID, "Elsewhere Elf", "Creature — Elf", 1, 1)
	return x
}

// cast puts Cloak and Dagger onto the battlefield and leaves its enters
// trigger asking for targets.
func (x *cdTable) cast(t *testing.T) uuid.UUID {
	t.Helper()
	id := castAndResolveCreature(t, x.g, "Cloak and Dagger, Entwined", "Legendary Creature — Human Hero", cloakAndDaggerOracle)
	if latestPickTarget(x.g, x.me.ID) == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	return id
}

// target answers the trigger's two clauses: the opponent, then the
// creature (or none).
func (x *cdTable) target(t *testing.T, creature uuid.UUID) {
	t.Helper()
	answerPickTargetPlayer(t, x.g, x.opp.ID)
	pick := latestPickTarget(x.g, x.me.ID)
	if pick == nil {
		t.Fatal("no prompt for the creature clause")
	}
	var refs []game.TargetRef
	if creature != uuid.Nil {
		refs = []game.TargetRef{{Kind: game.TargetCard, ID: creature, Slot: 1}}
	}
	if err := x.g.ResolvePickTargets(pick.ID, x.me.ID, refs); err != nil {
		t.Fatalf("creature clause: %v", err)
	}
}

func cdLabels(c *game.PendingChoice) []string {
	var out []string
	for _, o := range c.PickOptions {
		out = append(out, o.Label)
	}
	return out
}

// answer picks the option with this label.
func (x *cdTable) answer(t *testing.T, label string) {
	t.Helper()
	c := latestOptionPickFor(x.g, x.me.ID)
	if c == nil {
		t.Fatalf("no option pick is open: %+v", x.g.PendingChoices)
	}
	for i, o := range c.PickOptions {
		if o.Label == label {
			if err := x.g.ResolveOptionPick(c.ID, x.me.ID, i); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			return
		}
	}
	t.Fatalf("no option %q among %v", label, cdLabels(c))
}

func (x *cdTable) inExile(id uuid.UUID) bool { return exileHas(x.g, id) }

func (x *cdTable) onBattlefield(id uuid.UUID) bool {
	_, ok := battlefieldCard(x.g, id)
	return ok
}

// The hand half: the hand is revealed, the controller takes a nonland
// card out of it, and it goes back into the hand when Cloak and Dagger
// leaves the battlefield (CR 610.3: the previous zone).
func TestCloakAndDaggerExilesAHandCardUntilItLeaves(t *testing.T) {
	x := newCDTable(t)
	cloak := x.cast(t)
	x.target(t, x.theirs)
	passPriorityAroundTable(t, x.g)

	for _, id := range x.hand {
		card, _ := x.g.LookupCardForEffect(id)
		if !card.IsKnownTo(x.me.ID) || !card.IsKnownTo(x.other.ID) {
			t.Fatalf("%s was not revealed to the table", card.Name)
		}
	}
	c := latestOptionPickFor(x.g, x.me.ID)
	if c == nil {
		t.Fatal("no choice between the hand and the creature")
	}
	if got := cdLabels(c); len(got) != 3 || got[0] != "Exile nothing" || got[1] != "Exile a nonland card from their hand" || got[2] != "Exile Their Hill Giant" {
		t.Fatalf("options = %v", got)
	}
	x.answer(t, "Exile a nonland card from their hand")

	pick := openVariant(t, x.g)
	if !sameIDs(pick.DiscardOptions, x.hand[1:]) {
		t.Fatalf("offered %v, want the Bolt and the Bears card only", pick.DiscardOptions)
	}
	if err := x.g.ResolvePendingChoice(pick.ID, x.me.ID, []uuid.UUID{x.hand[1]}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, x.g)
	if !x.inExile(x.hand[1]) || x.opp.Hand.Contains(x.hand[1]) {
		t.Fatal("the Bolt was not exiled from their hand")
	}
	if rvDiscards(x.g, x.hand[1]) != 0 {
		t.Error("exiling a card from a hand is not a discard")
	}
	if !x.onBattlefield(x.theirs) {
		t.Error("the creature was exiled too; it is one or the other")
	}

	leaveBattlefield(t, x.g, x.me.ID, cloak)
	passPriorityAroundTable(t, x.g)
	if x.inExile(x.hand[1]) || !x.opp.Hand.Contains(x.hand[1]) {
		t.Fatalf("the Bolt did not return to its owner's hand (exile %v, hand %v)", x.inExile(x.hand[1]), x.opp.Hand.Contains(x.hand[1]))
	}
	if x.onBattlefield(x.hand[1]) {
		t.Error("the Bolt came back onto the battlefield")
	}
}

// The creature half: exiled until Cloak and Dagger leaves, then back on
// the battlefield under its owner's control (CR 610.3c).
func TestCloakAndDaggerExilesTheChosenCreatureUntilItLeaves(t *testing.T) {
	x := newCDTable(t)
	cloak := x.cast(t)
	x.target(t, x.theirs)
	passPriorityAroundTable(t, x.g)
	x.answer(t, "Exile Their Hill Giant")
	passPriorityAroundTable(t, x.g)

	if !x.inExile(x.theirs) || x.onBattlefield(x.theirs) {
		t.Fatal("the creature was not exiled")
	}
	if x.opp.Hand.Size() != 3 {
		t.Errorf("their hand has %d cards, want all 3 still there", x.opp.Hand.Size())
	}
	leaveBattlefield(t, x.g, x.me.ID, cloak)
	passPriorityAroundTable(t, x.g)
	// A returned object is a new one (CR 400.7), so find it by name.
	c, ok := e2BattlefieldNamed(x.g, "Their Hill Giant")
	if !ok || c.Controller != x.opp.ID {
		t.Fatalf("the creature did not return under its owner's control (on battlefield %v)", ok)
	}
}

// "You may": declining both exiles leaves everything where it was, with
// the hand revealed.
func TestCloakAndDaggerMayExileNothing(t *testing.T) {
	x := newCDTable(t)
	x.cast(t)
	x.target(t, x.theirs)
	passPriorityAroundTable(t, x.g)
	x.answer(t, "Exile nothing")
	passPriorityAroundTable(t, x.g)
	if x.opp.Hand.Size() != 3 || !x.onBattlefield(x.theirs) {
		t.Fatalf("something moved: hand %d, creature on battlefield %v", x.opp.Hand.Size(), x.onBattlefield(x.theirs))
	}
}

// "Up to one": with no creature chosen, and a hand with no nonland card,
// only the reveal happens and nothing is asked.
func TestCloakAndDaggerWithNothingToExileJustReveals(t *testing.T) {
	x := newCDTable(t)
	forest := revealHand(x.opp, rhForest())
	x.cast(t)
	x.target(t, uuid.Nil)
	passPriorityAroundTable(t, x.g)
	if c := latestOptionPickFor(x.g, x.me.ID); c != nil {
		t.Fatalf("asked %v with nothing to take", cdLabels(c))
	}
	card, _ := x.g.LookupCardForEffect(forest[0])
	if !card.IsKnownTo(x.other.ID) {
		t.Error("the hand was not revealed")
	}
}

// With no creature chosen, the choice is only the hand card or nothing.
func TestCloakAndDaggerWithoutACreatureOffersOnlyTheHand(t *testing.T) {
	x := newCDTable(t)
	x.cast(t)
	x.target(t, uuid.Nil)
	passPriorityAroundTable(t, x.g)
	c := latestOptionPickFor(x.g, x.me.ID)
	if c == nil {
		t.Fatal("no choice")
	}
	if got := cdLabels(c); len(got) != 2 || got[1] != "Exile a nonland card from their hand" {
		t.Fatalf("options = %v", got)
	}
}

// CR 601.2c: the creature must be controlled by the opponent chosen.
func TestCloakAndDaggerRefusesAnotherPlayersCreature(t *testing.T) {
	x := newCDTable(t)
	x.cast(t)
	answerPickTargetPlayer(t, x.g, x.opp.ID)
	pick := latestPickTarget(x.g, x.me.ID)
	if pick == nil {
		t.Fatal("no prompt for the creature clause")
	}
	if hasID(pick.PickTargetCards, x.spare) {
		t.Error("a creature another player controls is offered")
	}
	if !hasID(pick.PickTargetCards, x.theirs) {
		t.Error("the chosen opponent's creature is not offered")
	}
	if err := x.g.ResolvePickTargets(pick.ID, x.me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: x.spare, Slot: 1}}); err == nil {
		t.Fatal("a creature another player controls was accepted")
	}
}

// CR 608.2b: the creature changed hands while the trigger was on the
// stack, so it is no longer "a creature they control" and cannot be
// exiled; the hand still can.
func TestCloakAndDaggerRechecksTheCreatureAtResolution(t *testing.T) {
	x := newCDTable(t)
	x.cast(t)
	x.target(t, x.theirs)
	x.g.WithWriteLock(func() {
		if !x.g.GainControlForEffect(x.theirs, x.theirs, x.other.ID, game.IndefiniteDuration(), "test") {
			t.Error("control change refused")
		}
		// What the state check before the next priority does.
		x.g.RecomputeLayersIfStaleLocked()
	})
	passPriorityAroundTable(t, x.g)
	c := latestOptionPickFor(x.g, x.me.ID)
	if c == nil {
		t.Fatal("no choice")
	}
	if got := cdLabels(c); len(got) != 2 {
		t.Fatalf("options = %v, want the creature left out", got)
	}
}

// CR 610.3b: Cloak and Dagger left before the trigger resolved, so
// nothing is exiled. The hand is still revealed.
func TestCloakAndDaggerThatLeftExilesNothing(t *testing.T) {
	x := newCDTable(t)
	cloak := x.cast(t)
	x.target(t, x.theirs)
	leaveBattlefield(t, x.g, x.me.ID, cloak)
	passPriorityAroundTable(t, x.g)
	if c := latestOptionPickFor(x.g, x.me.ID); c != nil {
		t.Fatalf("asked %v for a Cloak and Dagger that is gone", cdLabels(c))
	}
	if x.inExile(x.theirs) || x.opp.Hand.Size() != 3 {
		t.Fatal("something was exiled")
	}
	card, _ := x.g.LookupCardForEffect(x.hand[1])
	if !card.IsKnownTo(x.other.ID) {
		t.Error("the hand was not revealed")
	}
}

// The held card survives a snapshot, and the restored game still sends
// it back to the hand when Cloak and Dagger leaves.
func TestCloakAndDaggerHandReturnSurvivesARestore(t *testing.T) {
	x := newCDTable(t)
	cloak := x.cast(t)
	x.target(t, x.theirs)
	passPriorityAroundTable(t, x.g)
	x.answer(t, "Exile a nonland card from their hand")
	pick := openVariant(t, x.g)
	if err := x.g.ResolvePendingChoice(pick.ID, x.me.ID, []uuid.UUID{x.hand[1]}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, x.g)

	restored := restoreThroughJSON(t, x.g)
	opp := restored.PlayerByIDForEffect(x.opp.ID)
	if !exileHas(restored, x.hand[1]) {
		t.Fatal("the Bolt is not in exile after the restore")
	}
	leaveBattlefield(t, restored, x.me.ID, cloak)
	passPriorityAroundTable(t, restored)
	if !opp.Hand.Contains(x.hand[1]) || exileHas(restored, x.hand[1]) {
		t.Fatal("the restored game did not return the Bolt to the hand")
	}
}

// The open choice is a restore point: answered after a restore, the
// creature still goes.
func TestCloakAndDaggerChoiceSurvivesARestore(t *testing.T) {
	x := newCDTable(t)
	x.cast(t)
	x.target(t, x.theirs)
	passPriorityAroundTable(t, x.g)
	if latestOptionPickFor(x.g, x.me.ID) == nil {
		t.Fatal("no choice")
	}
	restored := restoreThroughJSON(t, x.g)
	c := latestOptionPickFor(restored, x.me.ID)
	if c == nil {
		t.Fatal("the choice did not survive")
	}
	for i, o := range c.PickOptions {
		if o.Label == "Exile Their Hill Giant" {
			if err := restored.ResolveOptionPick(c.ID, x.me.ID, i); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
		}
	}
	passPriorityAroundTable(t, restored)
	if !exileHas(restored, x.theirs) {
		t.Fatal("the creature was not exiled after the restore")
	}
}
