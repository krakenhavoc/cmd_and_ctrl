package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// target_ability_cards_test.go — #1211's card proofs: the cards that
// could not be written while a target clause could only name a card,
// and the two whose "or ability" half shipped as a caveat.
//
// One board shape serves nearly all of them: an opponent's triggered
// or activated ability on the stack, an answer cast on top of it, and
// the question of whether the ability still resolves.

const (
	stifleOracle    = "b3b00911-ece7-4484-bc36-f211ce72b6cc"
	talesEndOracle  = "8ebb6fe0-3ed1-4cc9-bcb1-5317de199efc"
	voidslimeOracle = "c4851768-e210-4aaf-935e-715942e198f9"
)

// taAnnounce puts a triggered ability on the stack for `controller`,
// with an Effect that flips `fired` when it resolves — which is how
// every test below tells "countered" from "resolved" without needing
// a real card behind the trigger.
func taAnnounce(t *testing.T, g *game.Game, controller, source uuid.UUID, label string,
	targets []game.TargetRef, fired *bool) uuid.UUID {
	t.Helper()
	if err := g.AnnounceTrigger(controller, source, game.AbilityParams{
		Label:   label,
		Targets: targets,
	}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	id := acItemLabelled(g, label)
	if id == uuid.Nil {
		t.Fatalf("the announced trigger %q is not on the stack", label)
	}
	if fired != nil {
		g.WithWriteLock(func() {
			g.StackMeta[id].Effect = func(_ *game.Game, _ *game.StackItem) error {
				*fired = true
				return nil
			}
		})
	}
	return id
}

// taLegalTargets is the set a SPELL's own target clause offers right
// now, as a membership map.
func taLegalTargets(g *game.Game, chooser uuid.UUID, oracle string) map[uuid.UUID]bool {
	spec, ok := Lookup(oracle)
	if !ok {
		return nil
	}
	return taLegalForSpec(g, chooser, spec.Targets)
}

func taLegalForSpec(g *game.Game, chooser uuid.UUID, spec *game.TargetSpec) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	g.WithWriteLock(func() {
		for _, id := range g.LegalTargetsForEffect(game.SourceChooser(chooser), spec).Cards {
			out[id] = true
		}
	})
	return out
}

// TestAbilityTargetingCardsAreWired — the declarations, before any
// board. Each card's clause is the printed sentence and nothing
// wider, and none of the four that shipped caveated still declares
// one.
func TestAbilityTargetingCardsAreWired(t *testing.T) {
	for _, tc := range []struct {
		oracle          string
		name            string
		abilities, zone bool
	}{
		{stifleOracle, "Stifle", true, false},
		{talesEndOracle, "Tale's End", true, true},
		{voidslimeOracle, "Voidslime", true, true},
		{b16DisallowOracle, "Disallow", true, true},
		{deflectingSwatOracle, "Deflecting Swat", true, true},
		{boltBendOracle, "Bolt Bend", true, true},
	} {
		spec, ok := Lookup(tc.oracle)
		if !ok {
			t.Fatalf("%s is not registered", tc.name)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %q, want full — #1211 closed the ability gap",
				tc.name, spec.Completeness)
		}
		if len(spec.Caveats) != 0 {
			t.Errorf("%s still declares %v", tc.name, spec.Caveats)
		}
		if spec.Targets == nil || spec.Targets.Abilities != tc.abilities {
			t.Errorf("%s: Abilities = %v, want %v", tc.name, spec.Targets != nil && spec.Targets.Abilities, tc.abilities)
		}
		if got := len(spec.Targets.Zones) > 0; got != tc.zone {
			t.Errorf("%s: walks the stack zone = %v, want %v", tc.name, got, tc.zone)
		}
	}
	// Sublime Epiphany's fifth bullet is the one the card shipped
	// without.
	sub, _ := Lookup(sublimeEpiphanyOracle)
	if sub.Completeness != CompletenessFull || len(sub.Caveats) != 0 {
		t.Errorf("Sublime Epiphany: %q %v", sub.Completeness, sub.Caveats)
	}
	if sub.Modes == nil || len(sub.Modes.Options) != 5 {
		t.Fatalf("Sublime Epiphany prints five bullets, got %d", len(sub.Modes.Options))
	}
	if sub.Modes.Options[1].Targets == nil || !sub.Modes.Options[1].Targets.Abilities {
		t.Error("the second bullet targets an activated or triggered ability")
	}
}

// Stifle is the whole seam in one card: an opponent's trigger is a
// legal target, it is countered, and CR 701.6a leaves its source
// permanent exactly where it was.
func TestStifleCountersAnAbilityAndLeavesItsSource(t *testing.T) {
	g := newCatalogGame(t)
	// Reach a main phase BEFORE anything goes on the stack: every
	// advance drains it, and an ability that resolved is an ability
	// nothing can be pointed at.
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	fired := false
	ability := taAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, &fired)

	if !taLegalTargets(g, me.ID, stifleOracle)[ability] {
		t.Fatal("Stifle must offer an opponent's triggered ability")
	}
	castCatalogSpell(t, g, "Stifle", "Instant", stifleOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ability}})
	passPriorityAroundTable(t, g)

	if fired {
		t.Error("the countered ability resolved")
	}
	var stillThere bool
	g.WithWriteLock(func() { stillThere = g.StackMeta[ability] != nil })
	if stillThere {
		t.Error("the countered ability is still on the stack")
	}
	if !g.Battlefield.Contains(src) {
		t.Error("CR 701.6a: a countered ability goes nowhere and its source does not move")
	}
	if opp.Graveyard.Contains(src) {
		t.Error("the source permanent reached a graveyard")
	}
}

// "(Mana abilities can't be targeted.)" and "activated or triggered
// ABILITY" are both true of Stifle's clause without a predicate: a
// mana ability never reaches the stack, and a SPELL is not in the set
// because the clause walks no zone at all.
func TestStifleCannotTargetASpell(t *testing.T) {
	g := newCatalogGame(t)
	// Reach a main phase BEFORE anything goes on the stack: every
	// advance drains it, and an ability that resolved is an ability
	// nothing can be pointed at.
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if taLegalTargets(g, me.ID, stifleOracle)[bolt] {
		t.Error("Stifle offered a spell")
	}
	// And Disallow, whose clause is the wide one, does offer it.
	if !taLegalTargets(g, me.ID, b16DisallowOracle)[bolt] {
		t.Error("Disallow must offer a spell as well as an ability")
	}
}

// Tale's End is the clause that narrows the two halves DIFFERENTLY:
// any ability at all, but only a LEGENDARY spell.
func TestTalesEndTakesAnAbilityOrALegendarySpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	// Reach a main phase BEFORE anything goes on the stack: every
	// advance drains it, and an ability that resolved is an ability
	// nothing can be pointed at.
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	ability := taAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, nil)
	plain := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	legend := batch01OpponentCasts(t, g, opp, "Their Commander", "", "", nil)
	g.WithWriteLock(func() {
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == legend {
				g.Stack.Cards[i].TypeLine = "Legendary Creature — Human"
			}
		}
	})

	legal := taLegalTargets(g, me.ID, talesEndOracle)
	if !legal[ability] {
		t.Error("Tale's End takes any activated or triggered ability")
	}
	if !legal[legend] {
		t.Error("Tale's End takes a legendary spell")
	}
	if legal[plain] {
		t.Error("Tale's End must not take a nonlegendary spell")
	}
}

// Voidslime and Disallow print one sentence between them, and it is
// the widest clause in the family: either half of the stack. The
// counter itself is the same primitive for both kinds, which is what
// this resolves to prove.
func TestVoidslimeCountersAnActivatedAbility(t *testing.T) {
	g := newCatalogGame(t)
	// Reach a main phase BEFORE anything goes on the stack: every
	// advance drains it, and an ability that resolved is an ability
	// nothing can be pointed at.
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCatalogPermanent(g, opp.ID, "Their Engine", "Artifact", "", false)
	fired := false
	if err := g.ActivateAbility(opp.ID, src, game.AbilityParams{
		Label: "Their Engine — an activation",
	}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	ability := acItemLabelled(g, "Their Engine — an activation")
	g.WithWriteLock(func() {
		g.StackMeta[ability].Effect = func(_ *game.Game, _ *game.StackItem) error {
			fired = true
			return nil
		}
	})

	if !taLegalTargets(g, me.ID, voidslimeOracle)[ability] {
		t.Fatal("Voidslime must offer an activated ability")
	}
	castCatalogSpell(t, g, "Voidslime", "Instant", voidslimeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ability}})
	passPriorityAroundTable(t, g)
	if fired {
		t.Error("the countered activation resolved")
	}
	var stillThere bool
	g.WithWriteLock(func() { stillThere = g.StackMeta[ability] != nil })
	if stillThere {
		t.Error("the countered activation is still on the stack")
	}
}

// Sublime Epiphany's SECOND bullet — the mode the card shipped
// without, because it could not be pointed at anything.
func TestSublimeEpiphanyCountersAnAbilityWithItsSecondBullet(t *testing.T) {
	g := newCatalogGame(t)
	// Reach a main phase BEFORE anything goes on the stack: every
	// advance drains it, and an ability that resolved is an ability
	// nothing can be pointed at.
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	fired := false
	ability := taAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, &fired)

	spec, _ := Lookup(sublimeEpiphanyOracle)
	if !taLegalForSpec(g, me.ID, spec.Modes.Options[1].Targets)[ability] {
		t.Fatal("the second bullet must offer the ability")
	}
	castModal(t, g, "Sublime Epiphany", "Instant", sublimeEpiphanyOracle,
		[]int{1}, []game.TargetRef{modeRef(game.TargetCard, ability, 0, 0)})
	passPriorityAroundTable(t, g)
	if fired {
		t.Error("the countered ability resolved")
	}
}

// Deflecting Swat's "or ability" half (#1196 left it as a caveat):
// the Swat redirects an opponent's TRIGGERED ability, and legality is
// still judged from the ability's point of view.
func TestDeflectingSwatRedirectsAnAbility(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	// A real catalog activation, so the item carries the clause its
	// targets were announced under — which is what the retarget's
	// legality check re-runs. "{T}: Target player mills a card."
	shredder := b12Push(g, opp.ID, "Codex Shredder", "Artifact", b22CodexShredderOracle, 0, 0)
	if err := g.ActivateCatalogAbility(opp.ID, shredder, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	ability := acItemLabelled(g, "{T}: Target player mills a card.")
	if ability == uuid.Nil {
		t.Fatal("the activation is not on the stack")
	}
	mine, theirs := me.Library.Size(), third.Library.Size()

	if !taLegalTargets(g, me.ID, deflectingSwatOracle)[ability] {
		t.Fatal("Deflecting Swat must offer an ability on the stack")
	}
	castCatalogSpell(t, g, "Deflecting Swat", "Instant", deflectingSwatOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ability}})
	passPriorityAroundTable(t, g)

	prompt := latestRetarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the Swat opened no retarget prompt for the ability")
	}
	if !hasID(prompt.PickTargetPlayers, third.ID) || hasID(prompt.PickTargetPlayers, me.ID) {
		t.Fatalf("alternatives = %v, want the other seats and not the current target",
			prompt.PickTargetPlayers)
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: third.ID}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Library.Size() != mine {
		t.Error("the mill still landed on the original target")
	}
	if third.Library.Size() != theirs-1 {
		t.Errorf("the redirected ability did not mill the new target: %d → %d", theirs, third.Library.Size())
	}
}

// Bolt Bend's printed "with a single target" is one predicate over
// the stack ITEM, so it counts an ability's slots exactly as it
// counts a spell's — a two-target ability is not offered at all
// rather than refused after the click.
func TestBoltBendKeepsAMultiTargetAbilityOffThePicker(t *testing.T) {
	g := newCatalogGame(t)
	// Reach a main phase BEFORE anything goes on the stack: every
	// advance drains it, and an ability that resolved is an ability
	// nothing can be pointed at.
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	one := taAnnounce(t, g, opp.ID, src, "Their Rock — one target",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}, nil)
	two := taAnnounce(t, g, opp.ID, src, "Their Rock — two targets",
		[]game.TargetRef{
			{Kind: game.TargetPlayer, ID: me.ID},
			{Kind: game.TargetPlayer, ID: g.Seats[2].ID},
		}, nil)

	legal := taLegalTargets(g, me.ID, boltBendOracle)
	if !legal[one] {
		t.Error("a single-target ability is Bolt Bend's whole clause")
	}
	if legal[two] {
		t.Error("a two-target ability has no single target to change")
	}
}
