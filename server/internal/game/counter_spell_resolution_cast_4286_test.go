package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// counter_spell_resolution_cast_4286_test.go — #4286, CR 608.2g. Guile's
// "you may play that card without paying its mana cost" is made while the
// counterspell resolves, so the replaced counter asks the resolution_cast
// prompt over the exiled card (OfferResolutionCastForEffect) instead of
// leaving a pass-closed permission for the next priority. Stubbed
// replacements, as counter_spell_test.go does: the engine contract, for
// every counter verb. Guile itself, through the real counterspells, is
// pinned in cards/effects/guile_resolution_cast_4286_test.go.

// guileOffer is the open resolution_cast prompt chooser owes, or nil.
func guileOffer(g *Game, chooser uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceResolutionCast && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// counterVerbs are the three effect counter verbs a resolving spell or
// ability reaches (a hard counter, a counter that sends the spell to a
// zone, a counter that sends it to a library).
func counterVerbs() []struct {
	name    string
	counter func(g *Game, id uuid.UUID) error
} {
	return []struct {
		name    string
		counter func(g *Game, id uuid.UUID) error
	}{
		{"counter", func(g *Game, id uuid.UUID) error { return g.CounterTargetForEffect(id) }},
		{"counter to its owner's hand (Remand)", func(g *Game, id uuid.UUID) error {
			return g.CounterTargetToZoneForEffect(id, ZoneRef{Kind: ZoneHand, Owner: g.Seats[0].ID})
		}},
		{"counter to exile (Devious Cover-Up)", func(g *Game, id uuid.UUID) error {
			return g.CounterTargetToZoneForEffect(id, ZoneRef{Kind: ZoneExile})
		}},
		{"counter to the top of the library (Memory Lapse)", func(g *Game, id uuid.UUID) error {
			return g.CounterSpellToLibraryThenForEffect(id, TuckOptions{}, nil)
		}},
		{"counter to the bottom of the library (Spell Crumple)", func(g *Game, id uuid.UUID) error {
			return g.CounterSpellToLibraryThenForEffect(id, TuckOptions{ToBottom: true}, nil)
		}},
	}
}

// counteredByGuile is a table where seat 1's resolving spell counters
// seat 0's spell under a Guile-shaped replacement, through verb. The
// resolution is still open when it returns.
func counteredByGuile(t *testing.T, verb func(g *Game, id uuid.UUID) error) (g *Game, me, opp *Player, spell uuid.UUID) {
	t.Helper()
	g = newActiveGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	spell = pushSpellFor(t, g, "Creature — Bear", me, me)
	g.WithWriteLock(func() { g.RegisterReplacementForTest(counterWatcher(opp.ID, "exile", nil, exileInstead(opp.ID))) })
	counteringFor(g, opp)
	g.WithWriteLock(func() {
		if err := verb(g, spell); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	return g, me, opp, spell
}

// Every counter verb asks the countering player to cast the exiled card
// inside the resolution: the prompt is queued for them alone, the card is
// in exile and was not countered, and the cast is made through the prompt.
func TestAReplacedCounterOffersTheCastInsideTheResolution(t *testing.T) {
	for _, v := range counterVerbs() {
		t.Run(v.name, func(t *testing.T) {
			g, me, opp, spell := counteredByGuile(t, v.counter)
			offer := guileOffer(g, opp.ID)
			if offer == nil {
				t.Fatal("the countering player was not asked to cast the exiled card")
			}
			if guileOffer(g, me.ID) != nil {
				t.Error("the spell's owner was asked to cast their own countered spell")
			}
			if offer.MayCastKeyword != MayCastKeywordFreeCast || offer.DeclineLabel != "Leave it in exile" ||
				len(offer.DiscardOptions) != 1 || offer.DiscardOptions[0] != spell {
				t.Errorf("the offer: keyword %q, decline %q, options %v", offer.MayCastKeyword, offer.DeclineLabel, offer.DiscardOptions)
			}
			if !offer.midResolution {
				t.Error("the offer does not hold the countering resolution open")
			}
			if !g.Exile.Contains(spell) || hasEventFor(g, EventCounterSpell, spell) {
				t.Fatal("the spell must be in exile and not countered")
			}
			if err := g.CastForResolutionChoice(opp.ID, offer.ID, spell, CastSpellParams{FromZone: "exile"}); err != nil {
				t.Fatalf("the cast through the offer: %v", err)
			}
			if !g.Stack.Contains(spell) || g.Exile.Contains(spell) {
				t.Fatal("the card is not on the stack")
			}
			if guileOffer(g, opp.ID) != nil {
				t.Error("the offer stayed open after the one cast it was for")
			}
			var perm *CastPermission
			g.WithWriteLock(func() { perm = g.CastPermissionOnCardByIDForEffect(spell) })
			if perm != nil {
				t.Errorf("the cast left a permission behind: %+v", perm)
			}
		})
	}
}

// A decline leaves the card in exile with no permission over it.
func TestDecliningAReplacedCountersCastLeavesTheCardExiled(t *testing.T) {
	for _, v := range counterVerbs() {
		t.Run(v.name, func(t *testing.T) {
			g, _, opp, spell := counteredByGuile(t, v.counter)
			offer := guileOffer(g, opp.ID)
			if offer == nil {
				t.Fatal("no offer")
			}
			if err := g.ResolvePendingChoice(offer.ID, opp.ID, nil); err != nil {
				t.Fatalf("declining: %v", err)
			}
			if !g.Exile.Contains(spell) || g.Stack.Contains(spell) {
				t.Error("the declined card left exile")
			}
			var perm *CastPermission
			g.WithWriteLock(func() { perm = g.CastPermissionOnCardByIDForEffect(spell) })
			if perm != nil {
				t.Errorf("the decline left a permission: %+v", perm)
			}
			if err := g.CastSpell(opp.ID, spell, CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
				t.Error("a declined card was cast from exile afterwards")
			}
		})
	}
}

// The free play keeps its rules through the prompt: cast-only, and X is 0
// (CR 107.3b: a cast that pays no mana cost cannot announce a larger X).
func TestAReplacedCountersCastIsFreeWithXAtZero(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	x := xSorcery(me.ID, "test-stroke")
	pushStackSpell(t, g, x)
	g.WithWriteLock(func() { g.RegisterReplacementForTest(counterWatcher(opp.ID, "exile", nil, exileInstead(opp.ID))) })
	counteringFor(g, opp)
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(x.InstanceID); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	offer := guileOffer(g, opp.ID)
	if offer == nil {
		t.Fatal("no offer")
	}
	err := g.CastForResolutionChoice(opp.ID, offer.ID, x.InstanceID, CastSpellParams{FromZone: "exile", XValue: 5})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("a free cast announcing X=5: %v, want ErrInvalidParam", err)
	}
	if !g.Exile.Contains(x.InstanceID) || guileOffer(g, opp.ID) == nil {
		t.Fatal("the refused cast moved the card or closed the offer")
	}
	if err := g.CastForResolutionChoice(opp.ID, offer.ID, x.InstanceID, CastSpellParams{FromZone: "exile", XValue: 0}); err != nil {
		t.Fatalf("a free cast announcing X=0: %v", err)
	}
	if got := g.StackMeta[x.InstanceID].XValue; got != 0 {
		t.Errorf("X on the stack = %d, want 0", got)
	}
}

// With no resolution open the door answers false and the replaced counter
// keeps its pass-closed grant: the case a cause carried past the end of
// its resolution (#4279) reaches.
func TestAReplacedCounterWithNoResolutionOpenKeepsThePassClosedGrant(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := pushSpellFor(t, g, "Creature — Bear", me, me)
	g.WithWriteLock(func() { g.RegisterReplacementForTest(counterWatcher(opp.ID, "exile", nil, exileInstead(opp.ID))) })
	g.WithWriteLock(func() {
		if g.resolutionOpen {
			t.Fatal("the table has a resolution open")
		}
		cause := carriedCause{
			MoveCause: MoveCause{Kind: MoveCauseEffect, Controller: opp.ID, Item: uuid.New()},
			Source:    uuid.New(),
		}
		err := g.withCarriedCauseLocked(cause, func() error { return g.CounterTargetForEffect(spell) })
		if err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	if guileOffer(g, opp.ID) != nil {
		t.Fatal("a prompt was raised with no resolution to hold it")
	}
	if !g.Exile.Contains(spell) {
		t.Fatal("the spell is not in exile")
	}
	var perm *CastPermission
	g.WithWriteLock(func() { perm = g.CastPermissionOnCardByIDForEffect(spell) })
	if perm == nil || perm.Player != opp.ID || perm.Cost != "{0}" || !perm.CastOnly || perm.LapseOnPass != LapseStaysInExile {
		t.Errorf("grant %+v: want the free, cast-only, pass-closed permission", perm)
	}
}

// A commander countered under the replacement lands in exile like any
// other card and is offered; the CR 903.9a command-zone check then asks
// its owner, and if the owner takes the command zone the countering
// player's permission no longer opens anything.
func TestAReplacedCountersCommanderIsOfferedUntilItsOwnerTakesTheCommandZone(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmd := seatCommander(t, g.Stack, me)
	top, _ := g.Stack.Top()
	_, _ = g.Stack.Remove(cmd)
	pushStackSpell(t, g, top)
	g.WithWriteLock(func() { g.RegisterReplacementForTest(counterWatcher(opp.ID, "exile", nil, exileInstead(opp.ID))) })
	counteringFor(g, opp)
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(cmd); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	offer := guileOffer(g, opp.ID)
	if offer == nil || len(offer.DiscardOptions) != 1 || offer.DiscardOptions[0] != cmd {
		t.Fatalf("the exiled commander is not on offer: %+v", offer)
	}
	if err := g.ResolvePendingChoice(offer.ID, opp.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	// The owner takes the command zone (CR 903.9a), and nothing is left to cast.
	g.WithWriteLock(func() { g.resolutionOpen = false })
	answerCommanderReturn(t, g, me, cmd, true)
	if !me.Command.Contains(cmd) || g.Exile.Contains(cmd) {
		t.Fatal("the commander did not reach the command zone")
	}
	if err := g.CastSpell(opp.ID, cmd, CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("the countering player cast a commander out of the command zone")
	}
}
