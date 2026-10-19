package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// roonb_cards_test.go — #4128 slice b (the Roon of the Hidden Realm
// deck request): Beguiler of Wills, Blatant Thievery, Lavinia of the
// Tenth, Reflector Mage, See Double, Glasspool Mimic.

const (
	roonbBeguilerOracle = "20c20c86-4435-49e0-bbf3-b51c49c2eadd"
	roonbThieveryOracle = "8bb51dd2-cc62-48ea-ae8e-52351073a8c9"
	roonbLaviniaOracle  = "55df6fa3-3d24-46d5-b325-099336c3f0fd"
	roonbReflectOracle  = "5398362d-c45e-4d56-a622-8be1124730b9"
	roonbSeeDoubleOr    = "144d0817-348c-4171-aa7e-3468b23cf97d"
	roonbGlasspoolOr    = "c178953c-3888-4edd-9d0c-265bd82b1d24"
)

// roonbCreature seeds a vanilla creature, through the zone-move event so
// the layer listener stamps it.
func roonbCreature(g *game.Game, controller uuid.UUID, name string, power, toughness int) uuid.UUID {
	return b12Creature(g, controller, name, "Creature — Bear", power, toughness)
}

// roonbCard reads a battlefield card back.
func roonbCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	return copyBattlefieldCard(t, g, id)
}

func roonbControllerOf(g *game.Game, id uuid.UUID) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Controller
		}
	}
	return uuid.Nil
}

// --- Beguiler of Wills ----------------------------------------------

func roonbBeguilerActivate(g *game.Game, me *game.Player, beguiler, target uuid.UUID) error {
	return g.ActivateCatalogAbility(me.ID, beguiler, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
}

// The bound is the number of creatures you control, the Beguiler
// included: with the Beguiler and one other creature (two), a 2-power
// creature is takeable and a 3-power one is not.
func TestRoonbBeguilerTakesACreatureUpToYourCreatureCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beguiler := pushCatalogPermanent(g, me.ID, "Beguiler of Wills", "Creature — Human Wizard", roonbBeguilerOracle, false)
	roonbCreature(g, me.ID, "Friend", 1, 1)
	small := roonbCreature(g, opp.ID, "Small", 2, 2)
	big := roonbCreature(g, opp.ID, "Big", 3, 3)
	advanceToMainOf(t, g, 0)

	if err := roonbBeguilerActivate(g, me, beguiler, big); err == nil {
		t.Fatal("a 3-power creature is not a legal target with two creatures")
	}
	if err := roonbBeguilerActivate(g, me, beguiler, small); err != nil {
		t.Fatalf("a 2-power creature is a legal target with two creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	if roonbControllerOf(g, small) != me.ID {
		t.Fatal("Beguiler did not take the creature")
	}
	if !roonbCard(t, g, beguiler).Tapped {
		t.Error("the {T} cost was not paid")
	}
}

// No stated duration: the theft outlives the Beguiler (CR 611.2a).
func TestRoonbBeguilerTheftOutlivesTheBeguiler(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beguiler := pushCatalogPermanent(g, me.ID, "Beguiler of Wills", "Creature — Human Wizard", roonbBeguilerOracle, false)
	roonbCreature(g, me.ID, "Friend", 1, 1)
	small := roonbCreature(g, opp.ID, "Small", 2, 2)
	advanceToMainOf(t, g, 0)
	if err := roonbBeguilerActivate(g, me, beguiler, small); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(beguiler); err != nil {
			t.Fatal(err)
		}
	})
	if roonbControllerOf(g, small) != me.ID {
		t.Error("control went back when the Beguiler left; the effect has no duration")
	}
	advanceToUpkeepOf(t, g, 1)
	if roonbControllerOf(g, small) != me.ID {
		t.Error("control went back at end of turn; the effect has no duration")
	}
}

// The bound is read again as the ability resolves (CR 608.2b): losing
// a creature in response makes the target illegal.
func TestRoonbBeguilerRechecksTheBoundOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beguiler := pushCatalogPermanent(g, me.ID, "Beguiler of Wills", "Creature — Human Wizard", roonbBeguilerOracle, false)
	friend := roonbCreature(g, me.ID, "Friend", 1, 1)
	small := roonbCreature(g, opp.ID, "Small", 2, 2)
	advanceToMainOf(t, g, 0)
	if err := roonbBeguilerActivate(g, me, beguiler, small); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(friend); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if roonbControllerOf(g, small) != opp.ID {
		t.Error("with one creature left a 2-power target is illegal; nothing should be taken")
	}
}

// --- Blatant Thievery -----------------------------------------------

func roonbThieveryCast(t *testing.T, g *game.Game, targets ...uuid.UUID) error {
	t.Helper()
	refs := make([]game.TargetRef, 0, len(targets))
	for _, id := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return castCatalogSpellErr(t, g, "Blatant Thievery", "Sorcery", roonbThieveryOracle, refs)
}

func roonbPermanent(g *game.Game, controller uuid.UUID, name string) uuid.UUID {
	return b12Permanent(g, controller, name, "Artifact")
}

// One target per opponent, each from its own player, and all three are
// taken.
func TestRoonbBlatantThieveryTakesOnePermanentFromEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	a := roonbPermanent(g, g.Seats[1].ID, "A")
	b := roonbPermanent(g, g.Seats[2].ID, "B")
	c := roonbPermanent(g, g.Seats[3].ID, "C")
	if err := roonbThieveryCast(t, g, a, b, c); err != nil {
		t.Fatalf("one target per opponent is the printed cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if roonbControllerOf(g, id) != g.Seats[0].ID {
			t.Errorf("permanent %s was not taken", id)
		}
	}
}

// A target is REQUIRED for each live opponent (CR 601.2c), and each
// pick must be controlled by ITS opponent.
func TestRoonbBlatantThieveryRequiresAPickPerOpponentFromThatOpponent(t *testing.T) {
	g := newCatalogGame(t)
	a := roonbPermanent(g, g.Seats[1].ID, "A")
	b := roonbPermanent(g, g.Seats[2].ID, "B")
	c := roonbPermanent(g, g.Seats[3].ID, "C")
	b2 := roonbPermanent(g, g.Seats[2].ID, "B2")
	if err := roonbThieveryCast(t, g, a, b); err == nil {
		t.Fatal("two picks at a table with three live opponents must be refused")
	}
	if err := roonbThieveryCast(t, g, a, b, b2); err == nil {
		t.Fatal("the third pick must be the third opponent's permanent")
	}
	if err := roonbThieveryCast(t, g, b, a, c); err == nil {
		t.Fatal("the first pick must be the first opponent's permanent")
	}
}

// An opponent who has been eliminated is not an opponent: their clause
// is not required and nothing of theirs can be named.
func TestRoonbBlatantThieveryDoesNotReachAnEliminatedOpponent(t *testing.T) {
	g := newCatalogGame(t)
	a := roonbPermanent(g, g.Seats[1].ID, "A")
	b := roonbPermanent(g, g.Seats[2].ID, "B")
	g.Seats[3].Eliminated = true
	if err := roonbThieveryCast(t, g, a, b); err != nil {
		t.Fatalf("with the third opponent out, two picks are the whole cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if roonbControllerOf(g, id) != g.Seats[0].ID {
			t.Errorf("permanent %s was not taken", id)
		}
	}
}

// A live opponent with nothing to target makes the spell uncastable, as
// printed.
func TestRoonbBlatantThieveryNeedsATargetForEveryLiveOpponent(t *testing.T) {
	g := newCatalogGame(t)
	a := roonbPermanent(g, g.Seats[1].ID, "A")
	b := roonbPermanent(g, g.Seats[2].ID, "B")
	if err := roonbThieveryCast(t, g, a, b); err == nil {
		t.Fatal("the third opponent controls nothing and is still an opponent; the cast must be refused")
	}
}

// A permanent that changed hands in response is an illegal target for
// its slot (CR 608.2b), even when it went to another opponent; the other
// picks are still taken.
func TestRoonbBlatantThieveryRechecksEachPickAgainstItsOpponent(t *testing.T) {
	g := newCatalogGame(t)
	a := roonbPermanent(g, g.Seats[1].ID, "A")
	b := roonbPermanent(g, g.Seats[2].ID, "B")
	c := roonbPermanent(g, g.Seats[3].ID, "C")
	if err := roonbThieveryCast(t, g, a, b, c); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.Nil, a, g.Seats[2].ID, game.IndefiniteDuration(), "in response")
	})
	passPriorityAroundTable(t, g)
	if roonbControllerOf(g, a) == g.Seats[0].ID {
		t.Error("the first pick moved to another opponent; it is no longer the first opponent's")
	}
	if roonbControllerOf(g, b) != g.Seats[0].ID || roonbControllerOf(g, c) != g.Seats[0].ID {
		t.Error("the other two picks should still be taken")
	}
}

// --- Glasspool Mimic ------------------------------------------------

func roonbGlasspoolEnter(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	return castCatalogSpell(t, g, "Glasspool Mimic", "Creature — Shapeshifter Rogue", roonbGlasspoolOr, nil)
}

func TestRoonbGlasspoolMimicCopiesYourCreatureAndAddsTwoTypes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := seedCopyableCreature(g, me.ID, "Serra Angel", "Creature — Angel", 4, 4)
	theirs := seedCopyableCreature(g, opp.ID, "Their Dragon", "Creature — Dragon", 5, 5)

	id := roonbGlasspoolEnter(t, g, me)
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := copyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if len(pc.CopyOptions) != 1 || pc.CopyOptions[0] != angel {
		t.Fatalf("CopyOptions = %v, want only the creature you control (%s), never %s", pc.CopyOptions, angel, theirs)
	}
	resolveWithCopyChoice(t, g, angel)

	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Serra Angel" || !got.IsCopy() {
		t.Fatalf("name = %q, copy = %v, want a copy of Serra Angel", got.Name, got.IsCopy())
	}
	if got.CurrentPower() != 4 || got.CurrentToughness() != 4 {
		t.Errorf("copy is %d/%d, want 4/4", got.CurrentPower(), got.CurrentToughness())
	}
	subs := effectiveSubtypes(t, g, id)
	for _, want := range []string{"Angel", "Shapeshifter", "Rogue"} {
		found := false
		for _, s := range subs {
			if s == want {
				found = true
			}
		}
		if !found {
			t.Errorf("subtypes = %v, missing %q", subs, want)
		}
	}
}

// With no creature of yours to copy (an opponent's does not count) it
// enters uncopied. (The fixture carries no printing, so the 0/0 is not
// buried by the toughness state-based action; see stunt_double_test.go.)
func TestRoonbGlasspoolMimicCannotCopyAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedCopyableCreature(g, opp.ID, "Their Dragon", "Creature — Dragon", 5, 5)
	id := roonbGlasspoolEnter(t, g, me)
	passPriorityAroundTable(t, g)
	if copyPrompt(g) != nil {
		t.Fatal("a copy prompt opened with nothing of mine to copy")
	}
	if got := copyBattlefieldCard(t, g, id); got.IsCopy() || got.Name != "Glasspool Mimic" {
		t.Errorf("entered as %q (copy %v), want the uncopied Glasspool Mimic", got.Name, got.IsCopy())
	}
}

func TestRoonbGlasspoolMimicBackFaceIsALand(t *testing.T) {
	spec, ok := Lookup(game.CatalogKeyForFace(roonbGlasspoolOr, 1))
	if !ok || spec.Name != "Glasspool Shore" || spec.Completeness != CompletenessFull {
		t.Fatalf("Glasspool Shore spec = %+v, %v", spec.Name, ok)
	}
}
