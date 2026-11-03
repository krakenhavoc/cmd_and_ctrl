package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// target_relation_many_test.go — #4862: a dependent clause of several
// picks, each in the graveyard of the player an earlier clause picked.

// playersGraveyardSpec is "target player … up to three target cards from
// their graveyard".
func playersGraveyardSpec() *TargetSpec {
	return (&TargetSpec{Mode: "player", Players: true, Min: 1, Max: 1}).Then(&TargetSpec{
		Mode: "card_in_graveyard", Zones: []ZoneKind{ZoneGraveyard}, Min: 0, Max: 3,
		Relation: &TargetRelation{Slot: 0, Kind: TargetInTargetPlayersGraveyard},
	})
}

// playersGraveyardFixture is seat 0 and seat 1 with three graveyard cards
// each, and the announcement "seat 0 and its three cards".
func playersGraveyardFixture(t *testing.T) (g *Game, spec *TargetSpec, a, b []TargetRef) {
	t.Helper()
	g = newActiveGame(t)
	spec = playersGraveyardSpec()
	for i, p := range g.Seats[:2] {
		refs := []TargetRef{{Kind: TargetPlayer, ID: p.ID}}
		for j := 0; j < 3; j++ {
			c := ebLand(p.ID, string(rune('a'+i))+string(rune('0'+j)))
			p.Graveyard.PushTop(c)
			refs = append(refs, TargetRef{Kind: TargetCard, ID: c.InstanceID, Slot: 1})
		}
		if i == 0 {
			a = refs
		} else {
			b = refs
		}
	}
	return g, spec, a, b
}

func TestPlayersGraveyardRelationAnnouncement(t *testing.T) {
	g, spec, a, b := playersGraveyardFixture(t)
	src := SourceChooser(g.Seats[0].ID)
	steps := AnnouncedClauses(spec, nil, nil)
	g.WithWriteLock(func() {
		for n := 0; n <= 3; n++ {
			if err := g.validateAnnouncedTargetsLocked(src, steps, a[:1+n]); err != nil {
				t.Fatalf("%d cards from the player's graveyard: %v", n, err)
			}
		}
		mixed := []TargetRef{a[0], a[1], b[2]}
		if err := g.validateAnnouncedTargetsLocked(src, steps, mixed); !errors.Is(err, ErrIllegalTarget) {
			t.Fatalf("a card from another graveyard: %v", err)
		}
		extra := ebLand(g.Seats[0].ID, "a3")
		g.Seats[0].Graveyard.PushTop(extra)
		four := append(slices.Clone(a), TargetRef{Kind: TargetCard, ID: extra.InstanceID, Slot: 1})
		if err := g.validateAnnouncedTargetsLocked(src, steps, four); err == nil {
			t.Fatal("a fourth card was accepted")
		}
		// The candidate set follows the player picked.
		lt := g.LegalTargetsForAnnouncedClauseForEffect(src, steps, 1, b[:1])
		for _, r := range b[1:] {
			if !slices.Contains(lt.Cards, r.ID) {
				t.Fatalf("seat 1's card %s missing from %v", r.ID, lt.Cards)
			}
		}
		for _, r := range a[1:] {
			if slices.Contains(lt.Cards, r.ID) {
				t.Fatalf("seat 0's card %s offered beside seat 1", r.ID)
			}
		}
	})
}

// CR 608.2b: each pick is judged on its own; a card that left the
// graveyard is illegal and the rest stay legal. An anchor that has become
// an illegal target still names its graveyard.
func TestPlayersGraveyardRelationResolutionRecheck(t *testing.T) {
	g, spec, a, _ := playersGraveyardFixture(t)
	g.WithWriteLock(func() {
		item := &StackItem{Controller: g.Seats[0].ID, Targets: a, targetSpec: spec}
		_, _ = g.Seats[0].Graveyard.Remove(a[1].ID)
		if g.TargetStillLegalForEffect(item, a[1]) {
			t.Fatal("a card that left the graveyard stayed legal")
		}
		for _, r := range a[2:] {
			if !g.TargetStillLegalForEffect(item, r) {
				t.Fatalf("card %s still in the graveyard was dropped", r.ID)
			}
		}
	})
}

// CR 115.7d, 707.10c: a copy that changes the player must move every card
// to the new player's graveyard, and can.
func TestPlayersGraveyardRelationCopyRefillsEveryPick(t *testing.T) {
	g, spec, a, b := playersGraveyardFixture(t)
	src := SourceChooser(g.Seats[0].ID)
	steps := AnnouncedClauses(spec, nil, nil)
	g.WithWriteLock(func() {
		moved := append([]TargetRef{b[0]}, a[1:]...)
		if !g.copyTargetsCanFinishLocked(src, steps, a, moved, 1) {
			t.Fatal("the copy picker could not move three cards to the new player")
		}
		if err := g.retargetCheckLocked(src, steps, a, moved, RetargetChooseNew); !errors.Is(err, ErrIllegalTarget) {
			t.Fatalf("the old cards kept beside the new player: %v", err)
		}
		// Seat 1 has only two cards left: three cannot follow.
		_, _ = g.Seats[1].Graveyard.Remove(b[3].ID)
		if g.copyTargetsCanFinishLocked(src, steps, a, moved, 1) {
			t.Fatal("three cards followed to a graveyard of two")
		}
	})
}

// A related retarget asks the player, then each card, and commits only
// the finished set.
func TestPlayersGraveyardRelatedRetargetMovesEveryCard(t *testing.T) {
	g, spec, a, b := playersGraveyardFixture(t)
	id := uuid.New()
	g.StackMeta = make(map[uuid.UUID]*StackItem)
	g.StackMeta[id] = &StackItem{ID: id, Controller: g.Seats[0].ID, Targets: a, targetSpec: spec}
	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{ItemID: id, Chooser: g.Seats[0].ID, Policy: RetargetChooseNew}); err != nil {
			t.Fatal(err)
		}
	})
	choice := g.PendingChoices[0]
	if !slices.Contains(choice.PickTargetPlayers, b[0].ID) {
		t.Fatalf("the other player was not offered: %+v", choice.PickTargetPlayers)
	}
	if err := g.ResolveRetarget(choice.ID, choice.Chooser, b[:1]); err != nil {
		t.Fatal(err)
	}
	for k := 1; k <= 3; k++ {
		if !slices.Equal(g.StackMeta[id].Targets, a) {
			t.Fatal("retarget committed an incomplete set")
		}
		choice = g.PendingChoices[0]
		if choice.PickTargetMin != 1 || !slices.Contains(choice.PickTargetCards, b[k].ID) || slices.Contains(choice.PickTargetCards, a[k].ID) {
			t.Fatalf("card prompt %d: min %d, cards %v", k, choice.PickTargetMin, choice.PickTargetCards)
		}
		if err := g.ResolveRetarget(choice.ID, choice.Chooser, b[k:k+1]); err != nil {
			t.Fatalf("card %d: %v", k, err)
		}
	}
	if !slices.Equal(g.StackMeta[id].Targets, b) {
		t.Fatalf("the completed set was not committed: %v", g.StackMeta[id].Targets)
	}
}
