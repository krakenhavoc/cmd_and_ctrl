package testsa

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activate_only_once_cards_test.go — the cards of #5221, each a plain
// "Activate only once" (CR 602.5b) without the exhaust keyword: Possessed
// Goat and Haunted Screen (Duskmourn, #4466) and Mild-Mannered Librarian
// (Foundations, #4465). Each test plays the real catalog entry. The
// engine half is game/activate_only_once_test.go.

const (
	possessedGoatOracle         = "821f4ff3-ebc7-4f47-ac35-d00093b6d35b"
	hauntedScreenOracle         = "49f92c0b-16ac-4cf0-8ec3-e3638990ee75"
	mildManneredLibrarianOracle = "fb7070ab-c234-4d54-8cdc-caf106bf24a6"
)

func TestActivateOnlyOnceCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		possessedGoatOracle:         "Possessed Goat",
		hauntedScreenOracle:         "Haunted Screen",
		mildManneredLibrarianOracle: "Mild-Mannered Librarian",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
		if len(spec.Activated) != 1 || !spec.Activated[0].OnlyOnce || spec.Activated[0].Exhaust {
			t.Errorf("%s: want one activated ability with OnlyOnce and without Exhaust", name)
		}
	}
}

// checkOnlyOnce holds the label and the bit together: the printed clause
// without the bit, the bit without the clause, the bit beside a keyword
// that already carries it, and two such rows sharing one label. "Activate
// only once each turn" is a different limit and needs nothing.
func TestCheckOnlyOnceRefusesWhatTheRecordCannotKeep(t *testing.T) {
	const once = "{1}: Draw a card. Activate only once."
	effectstest.ExpectRegisterPanic(t, "does not set OnlyOnce", func() {
		effects.CheckOnlyOnce(effects.Spec{Name: "Fixture", Activated: []effects.ActivatedAbility{{Label: once}}})
	})
	effectstest.ExpectRegisterPanic(t, "does not print", func() {
		effects.CheckOnlyOnce(effects.Spec{Name: "Fixture", Activated: []effects.ActivatedAbility{{Label: "{1}: Draw a card.", OnlyOnce: true}}})
	})
	effectstest.ExpectRegisterPanic(t, "beside Exhaust or PowerUp", func() {
		effects.CheckOnlyOnce(effects.Spec{Name: "Fixture", Activated: []effects.ActivatedAbility{{Label: "Exhaust — " + once, OnlyOnce: true, Exhaust: true}}})
	})
	effectstest.ExpectRegisterPanic(t, "share one use", func() {
		effects.CheckOnlyOnce(effects.Spec{Name: "Fixture", Activated: []effects.ActivatedAbility{{Label: once, OnlyOnce: true}, {Label: once, OnlyOnce: true}}})
	})
	effects.CheckOnlyOnce(effects.Spec{Name: "Fixture", Activated: []effects.ActivatedAbility{
		{Label: "{1}: Draw a card. Activate only once each turn."},
		{Label: once, OnlyOnce: true},
	}})
}

// --- Possessed Goat ----------------------------------------------------------

// The Goat pays {3} and a discard, becomes a black Demon Goat with three
// +1/+1 counters, and cannot do it again. A Rangers' Refueler beside it
// does not draw: the ability is not an exhaust ability.
func TestPossessedGoatBecomesABlackDemonOnlyOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	goat := effectstest.PushCatalogPermanent(g, me.ID, "Possessed Goat", "Creature — Goat", possessedGoatOracle, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Rangers' Refueler", "Artifact — Vehicle", effectstest.RangersRefuelerOracle, false)

	pitched := effectstest.PushHandCard(g, me)
	before := effectstest.HandSizeOf(g, me)
	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, goat, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitched}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.HandSizeOf(g, me); got != before-1 {
		t.Errorf("hand went %d → %d, want one card discarded and none drawn — Rangers' Refueler watches exhaust abilities only", before, got)
	}
	if got := effectstest.PlusOnes(g, goat); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
	c := effectstest.LayeredCard(t, g, goat)
	if !c.HasColor("B") || !c.HasSubtype("Demon") || !c.HasSubtype("Goat") {
		t.Errorf("the Goat is not a black Demon Goat: %+v", c.Effective())
	}

	second := effectstest.PushHandCard(g, me)
	effectstest.FillPool(me, 3)
	err := g.ActivateCatalogAbility(me.ID, goat, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{second}})
	if !errors.Is(err, game.ErrAbilityActivatedOnce) {
		t.Fatalf("second activation: err = %v, want ErrAbilityActivatedOnce", err)
	}
	if !slices.ContainsFunc(me.Hand.Cards, func(c game.Card) bool { return c.InstanceID == second }) {
		t.Error("the refused activation discarded a card")
	}
}

// A Goat that leaves and comes back is a new object (CR 400.7): it is a
// plain Goat again, and it may activate the ability again.
func TestPossessedGoatMayActivateAgainAsANewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	goat := effectstest.PushCatalogPermanent(g, me.ID, "Possessed Goat", "Creature — Goat", possessedGoatOracle, false)
	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, goat, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{effectstest.PushHandCard(g, me)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.BounceAndReturnForTest(t, g, goat, me.ID)
	if c := effectstest.LayeredCard(t, g, goat); c.HasSubtype("Demon") || c.HasColor("B") {
		t.Errorf("the returned Goat kept the old object's change: %+v", c.Effective())
	}
	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, goat, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{effectstest.PushHandCard(g, me)}}); err != nil {
		t.Fatalf("the returned Goat cannot activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, goat); got != 3 {
		t.Errorf("+1/+1 counters on the new object = %d, want 3", got)
	}
}

// --- Haunted Screen ----------------------------------------------------------

func TestHauntedScreenMakesFiveColorsOfMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	screen := effectstest.PushCatalogPermanent(g, me.ID, "Haunted Screen", "Artifact", hauntedScreenOracle, false)
	if got := effectstest.LtProduced(t, g, screen); !slices.Equal(got, []string{"{G|U|R}", "{W|B}"}) {
		t.Fatalf("the Screen adds %v, want {W|B} and {G|U|R}", got)
	}
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, screen, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}, Pay 1 life: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want 1 paid", life, me.Life)
	}
	if n := effectstest.B10ResolveAllManaPicks(t, g, me.ID, "U"); n != 1 {
		t.Fatalf("expected one colour pick, answered %d", n)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want [U]", got)
	}
}

// {7}: seven +1/+1 counters, then a 0/0 Spirit artifact creature, so a
// 7/7. The second activation is refused, and a restore keeps it so.
func TestHauntedScreenBecomesASevenSevenSpiritOnlyOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	screen := effectstest.PushCatalogPermanent(g, me.ID, "Haunted Screen", "Artifact", hauntedScreenOracle, false)
	effectstest.FillPool(me, 7)
	if err := g.ActivateCatalogAbility(me.ID, screen, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.PlusOnes(g, screen); got != 7 {
		t.Errorf("+1/+1 counters = %d, want 7", got)
	}
	c := effectstest.LayeredCard(t, g, screen)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Spirit") {
		t.Errorf("the Screen is not a Spirit artifact creature: %+v", c.Effective())
	}
	if p, tough := c.CurrentPower(), c.CurrentToughness(); p != 7 || tough != 7 {
		t.Errorf("the Spirit is %d/%d, want 7/7", p, tough)
	}

	effectstest.FillPool(me, 7)
	if err := g.ActivateCatalogAbility(me.ID, screen, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrAbilityActivatedOnce) {
		t.Errorf("second activation: err = %v, want ErrAbilityActivatedOnce", err)
	}

	restored := effectstest.RestoreRoundTrip(t, g, true)
	rme := restored.Seats[0]
	rme.ManaPool = nil
	effectstest.FillPool(rme, 7)
	if err := restored.ActivateCatalogAbility(rme.ID, screen, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrAbilityActivatedOnce) {
		t.Errorf("after a restore: err = %v, want ErrAbilityActivatedOnce", err)
	}
	if c := effectstest.LayeredCard(t, restored, screen); !c.IsCreature() || !c.HasSubtype("Spirit") {
		t.Errorf("after a restore the Screen is no longer a Spirit creature: %+v", c.Effective())
	}
}

// --- Mild-Mannered Librarian -------------------------------------------------

// "Becomes a Werewolf" replaces its creature types (CR 205.1a): a
// Werewolf, not a Human Werewolf, with two counters, and a card drawn.
func TestMildManneredLibrarianBecomesAWerewolfOnlyOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lib := effectstest.PushCatalogPermanent(g, me.ID, "Mild-Mannered Librarian", "Creature — Human", mildManneredLibrarianOracle, false)
	before := effectstest.HandSizeOf(g, me)
	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, lib, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.HandSizeOf(g, me); got != before+1 {
		t.Errorf("hand went %d → %d, want one card drawn", before, got)
	}
	if got := effectstest.PlusOnes(g, lib); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if c := effectstest.LayeredCard(t, g, lib); !c.HasSubtype("Werewolf") || c.HasSubtype("Human") {
		t.Errorf("want a Werewolf that is no longer a Human: %+v", c.Effective())
	}

	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, lib, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrAbilityActivatedOnce) {
		t.Errorf("second activation: err = %v, want ErrAbilityActivatedOnce", err)
	}
}
