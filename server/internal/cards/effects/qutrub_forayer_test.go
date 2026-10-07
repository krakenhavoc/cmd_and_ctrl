package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// qutrub_forayer_test.go — #1807. Qutrub Forayer was held on the
// single-graveyard row for "target creature that was dealt damage this
// turn"; that predicate has existed since Covert Cutpurse, so the card
// is built on the two halves that shipped.

const oracleQutrubForayer = "500ccb61-83b0-4b1f-ada6-65850ed5a825"

// qutrubEnters casts the Forayer and returns the open mode prompt for
// its controller.
func qutrubEnters(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	castAndResolveCreature(t, g, "Qutrub Forayer", "Creature — Zombie Horror", oracleQutrubForayer)
	mode := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt from the enters trigger: %+v", g.PendingChoices)
	}
	return mode
}

func TestQutrubForayerIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(oracleQutrubForayer)
	if !ok || spec.Name != "Qutrub Forayer" || spec.Completeness != CompletenessFull {
		t.Fatalf("registered = %v as %q (%s), want Qutrub Forayer, full", ok, spec.Name, spec.Completeness)
	}
}

// The graveyard bullet refuses cards from two graveyards at announce,
// then exiles two from one and leaves the rest alone.
func TestQutrubForayerExilesTwoFromOneGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := batch01GraveyardCard(opp, "A", "Sorcery")
	b := batch01GraveyardCard(opp, "B", "Sorcery")
	c := batch01GraveyardCard(opp, "C", "Sorcery")
	mine := batch01GraveyardCard(me, "Mine", "Sorcery")

	mode := qutrubEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("choosing the graveyard bullet: %v", err)
	}
	if err := sgAnswerTargets(t, g, a, mine); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("two graveyards: err = %v, want ErrIllegalTarget", err)
	}
	if err := sgAnswerTargets(t, g, a, b, c); err == nil {
		t.Fatal("three cards were accepted for an up-to-two clause")
	}
	if err := sgAnswerTargets(t, g, a, b); err != nil {
		t.Fatalf("one graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !sgInExile(g, a, b) {
		t.Error("the two picked cards are exiled")
	}
	if !opp.Graveyard.Contains(c) || !me.Graveyard.Contains(mine) {
		t.Error("the unpicked cards stay in their graveyards")
	}
}

// A picked card that leaves its graveyard in response is skipped and
// the other still goes (CR 608.2b).
func TestQutrubForayerSkipsATargetThatLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := batch01GraveyardCard(opp, "A", "Sorcery")
	b := batch01GraveyardCard(opp, "B", "Sorcery")

	mode := qutrubEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("choosing the graveyard bullet: %v", err)
	}
	if err := sgAnswerTargets(t, g, a, b); err != nil {
		t.Fatalf("one graveyard: %v", err)
	}
	g.WithWriteLock(func() {
		if _, err := opp.Graveyard.Remove(a); err != nil {
			t.Fatalf("removing A: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !sgInExile(g, b) {
		t.Error("the card that stayed is exiled")
	}
}

// The destroy bullet offers a creature that was dealt damage this turn
// — anyone's, including the Forayer's controller — and nothing else.
func TestQutrubForayerDestroysADamagedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	hurt := auraBear(g, opp.ID)
	fresh := auraBear(g, opp.ID)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})

	mode := qutrubEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("choosing the destroy bullet: %v", err)
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		if !hasID(p.PickTargetCards, hurt) || hasID(p.PickTargetCards, fresh) {
			t.Errorf("the Forayer offers %v, want the damaged creature only", p.PickTargetCards)
		}
		pickCard(t, g, me.ID, hurt)
	}
	passPriorityAroundTable(t, g)
	if zone, _ := cardWhere(g, opp, hurt); zone != "graveyard" {
		t.Errorf("the damaged creature is in %s, want the graveyard", zone)
	}
	if zone, _ := cardWhere(g, opp, fresh); zone != "battlefield" {
		t.Errorf("the undamaged creature is in %s, want the battlefield", zone)
	}
}

// With no creature dealt damage this turn the destroy bullet has no
// legal target, so it cannot be chosen (CR 603.3c).
func TestQutrubForayerDestroyBulletNeedsADamagedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	fresh := auraBear(g, opp.ID)

	mode := qutrubEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err == nil {
		t.Fatal("the destroy bullet was accepted with no damaged creature on the table")
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("the graveyard bullet is always choosable: %v", err)
	}
	passPriorityAroundTable(t, g)
	if zone, _ := cardWhere(g, opp, fresh); zone != "battlefield" {
		t.Errorf("the undamaged creature is in %s, want the battlefield", zone)
	}
}
