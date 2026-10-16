package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hand_revealed_cards_test.go — ADR 0199, #3953: the three cards that make
// a player play with their hand revealed.

const (
	hrStromgaldSpy = "b7aa0625-c691-4219-b099-3798b349d09c"
	hrTelepathy    = "d2e647ac-717f-46e6-baa8-6a793e7bdc32"
	hrRevelation   = "251ff4e4-3be4-424e-8dd7-eb4ede7b415c"
)

func hrRevealed(g *game.Game, player uuid.UUID) []string {
	var out []string
	g.ReadSnapshot(func() { out = g.HandRevealedLocked(player) })
	return out
}

// Stromgald Spy: yes reveals the defending player's hand and the Spy deals
// no combat damage; no does neither. Only the defender's hand is revealed.
func TestStromgaldSpyRevealsTheDefendersHandInsteadOfDealingDamage(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		spy := ncdAttacker(g, "Stromgald Spy", hrStromgaldSpy, 2, 4)
		life := g.Seats[1].Life
		ncdPlay(t, g, spy, yes)
		got, lost := hrRevealed(g, g.Seats[1].ID), ncdLifeLost(g, life)
		if yes && (!slices.Equal(got, []string{"Stromgald Spy"}) || lost != 0 || !ncdMarked(g, spy)) {
			t.Errorf("yes: revealed by %v (want [Stromgald Spy]), life lost %d (want 0), marked %v", got, lost, ncdMarked(g, spy))
		}
		if !yes && (len(got) != 0 || lost != 2 || ncdMarked(g, spy)) {
			t.Errorf("no: revealed by %v (want none), life lost %d (want 2), marked %v", got, lost, ncdMarked(g, spy))
		}
		for _, seat := range []int{0, 2, 3} {
			if r := hrRevealed(g, g.Seats[seat].ID); len(r) != 0 {
				t.Errorf("seat %d is not the defender and its hand is revealed by %v", seat, r)
			}
		}
	}
}

// The reveal lasts for as long as the Spy remains on the battlefield
// (CR 611.2b): past the turn, and not a moment after the Spy leaves.
func TestStromgaldSpyRevealLastsWhileItRemains(t *testing.T) {
	g := newCatalogGame(t)
	spy := ncdAttacker(g, "Stromgald Spy", hrStromgaldSpy, 2, 4)
	defender := g.Seats[1].ID
	ncdPlay(t, g, spy, true)
	endTurn(t, g)
	if got := hrRevealed(g, defender); len(got) == 0 {
		t.Fatal("the reveal ended with the turn; it lasts while the Spy remains")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(spy); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if got := hrRevealed(g, defender); len(got) != 0 {
		t.Errorf("the Spy left and the hand is still revealed by %v", got)
	}
}

// A Spy that has left before its trigger resolves reveals nothing: the
// "for as long as" duration never starts (CR 611.2b).
func TestStromgaldSpyThatLeftRevealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	spy := ncdAttacker(g, "Stromgald Spy", hrStromgaldSpy, 2, 4)
	defender := g.Seats[1].ID
	pr7bAttack(t, g, defender, spy)
	advanceTo(t, g, game.StepDeclareBlockers)
	finishBlockDeclarations(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(spy); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	asked := false
	for i := 0; i < 24; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				open = c
				break
			}
		}
		if open == nil {
			if len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0 {
				break
			}
			if err := g.PassPriority(); err != nil {
				break
			}
			continue
		}
		if open.Kind != game.PendingChoiceConfirm {
			t.Fatalf("unexpected prompt %s", open.Kind)
		}
		asked = true
		if err := g.ResolveConfirm(open.ID, g.Seats[0].ID, true); err != nil {
			t.Fatalf("confirm: %v", err)
		}
	}
	if !asked {
		t.Fatal("the trigger never asked; the test did not reach the case")
	}
	if got := hrRevealed(g, defender); len(got) != 0 {
		t.Errorf("a Spy that had left revealed the hand: %v", got)
	}
}

// A grant with no duration of its own is "until end of turn", read live.
func TestHandRevealedGrantEndsWithItsDuration(t *testing.T) {
	g := newCatalogGame(t)
	defender := g.Seats[1].ID
	g.WithWriteLock(func() {
		g.GrantHandRevealedForEffect(defender, "test", uuid.Nil, g.UntilEndOfTurnDuration())
	})
	if got := hrRevealed(g, defender); !slices.Equal(got, []string{"test"}) {
		t.Fatalf("the grant reads %v", got)
	}
	endTurn(t, g)
	if got := hrRevealed(g, defender); len(got) != 0 {
		t.Errorf("an until-end-of-turn grant outlived its turn: %v", got)
	}
}

// Telepathy: every opponent's hand, not its controller's.
func TestTelepathyRevealsEachOpponentsHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	apaPush(g, me.ID, me.ID, game.Card{Name: "Telepathy", TypeLine: "Enchantment", OracleID: hrTelepathy})
	if got := hrRevealed(g, me.ID); len(got) != 0 {
		t.Errorf("Telepathy revealed its controller's hand: %v", got)
	}
	for _, p := range g.Seats[1:] {
		if got := hrRevealed(g, p.ID); !slices.Equal(got, []string{"Telepathy"}) {
			t.Errorf("opponent %s: revealed by %v, want [Telepathy]", p.Name, got)
		}
	}
}

// Revelation: every player's hand, its controller's too; and it reveals
// nothing once it has lost its abilities (it is a static ability).
func TestRevelationRevealsEveryHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rev := apaPush(g, me.ID, me.ID, game.Card{Name: "Revelation", TypeLine: "World Enchantment", OracleID: hrRevelation})
	for _, p := range g.Seats {
		if got := hrRevealed(g, p.ID); !slices.Equal(got, []string{"Revelation"}) {
			t.Errorf("%s: revealed by %v, want [Revelation]", p.Name, got)
		}
	}
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(rev),
			[]game.Mod{game.LoseAllAbilitiesMod()}, game.IndefiniteDuration(), "test — loses all abilities")
		g.RecomputeLayersIfStaleLocked()
	})
	for _, p := range g.Seats {
		if got := hrRevealed(g, p.ID); len(got) != 0 {
			t.Errorf("%s: a Revelation with no abilities still reveals: %v", p.Name, got)
		}
	}
}
