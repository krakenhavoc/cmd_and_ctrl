package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
		{effectstest.StifleOracle, "Stifle", true, false},
		{effectstest.TalesEndOracle, "Tale's End", true, true},
		{effectstest.VoidslimeOracle, "Voidslime", true, true},
		{effectstest.B16DisallowOracle, "Disallow", true, true},
		{effectstest.DeflectingSwatOracle, "Deflecting Swat", true, true},
		{effectstest.BoltBendOracle, "Bolt Bend", true, true},
	} {
		spec, ok := effects.Lookup(tc.oracle)
		if !ok {
			t.Fatalf("%s is not registered", tc.name)
		}
		if spec.Completeness != effects.CompletenessFull {
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
	sub, _ := effects.Lookup(effectstest.SublimeEpiphanyOracle)
	if sub.Completeness != effects.CompletenessFull || len(sub.Caveats) != 0 {
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
	g := effectstest.NewCatalogGame(t)
	// Reach a main phase BEFORE anything goes on the stack: every
	// advance drains it, and an ability that resolved is an ability
	// nothing can be pointed at.
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	fired := false
	ability := effectstest.TaAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, &fired)

	if !effectstest.TaLegalTargets(g, me.ID, effectstest.StifleOracle)[ability] {
		t.Fatal("Stifle must offer an opponent's triggered ability")
	}
	effectstest.CastCatalogSpell(t, g, "Stifle", "Instant", effectstest.StifleOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ability}})
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	// Reach a main phase BEFORE anything goes on the stack: every
	// advance drains it, and an ability that resolved is an ability
	// nothing can be pointed at.
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if effectstest.TaLegalTargets(g, me.ID, effectstest.StifleOracle)[bolt] {
		t.Error("Stifle offered a spell")
	}
	// And Disallow, whose clause is the wide one, does offer it.
	if !effectstest.TaLegalTargets(g, me.ID, effectstest.B16DisallowOracle)[bolt] {
		t.Error("Disallow must offer a spell as well as an ability")
	}
}

// Tale's End is the clause that narrows the two halves DIFFERENTLY:
// any ability at all, but only a LEGENDARY spell.
func TestTalesEndTakesAnAbilityOrALegendarySpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	// Reach a main phase BEFORE anything goes on the stack: every
	// advance drains it, and an ability that resolved is an ability
	// nothing can be pointed at.
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	ability := effectstest.TaAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, nil)
	plain := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	legend := effectstest.Batch01OpponentCasts(t, g, opp, "Their Commander", "", "", nil)
	g.WithWriteLock(func() {
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == legend {
				g.Stack.Cards[i].TypeLine = "Legendary Creature — Human"
			}
		}
	})

	legal := effectstest.TaLegalTargets(g, me.ID, effectstest.TalesEndOracle)
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
	g := effectstest.NewCatalogGame(t)
	// Reach a main phase BEFORE anything goes on the stack: every
	// advance drains it, and an ability that resolved is an ability
	// nothing can be pointed at.
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.PushCatalogPermanent(g, opp.ID, "Their Engine", "Artifact", "", false)
	fired := false
	if err := g.ActivateAbility(opp.ID, src, game.AbilityParams{
		Label: "Their Engine — an activation",
	}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	ability := effectstest.AcItemLabelled(g, "Their Engine — an activation")
	g.WithWriteLock(func() {
		g.StackMeta[ability].Effect = func(_ *game.Game, _ *game.StackItem) error {
			fired = true
			return nil
		}
	})

	if !effectstest.TaLegalTargets(g, me.ID, effectstest.VoidslimeOracle)[ability] {
		t.Fatal("Voidslime must offer an activated ability")
	}
	effectstest.CastCatalogSpell(t, g, "Voidslime", "Instant", effectstest.VoidslimeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ability}})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	// Reach a main phase BEFORE anything goes on the stack: every
	// advance drains it, and an ability that resolved is an ability
	// nothing can be pointed at.
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	fired := false
	ability := effectstest.TaAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, &fired)

	spec, _ := effects.Lookup(effectstest.SublimeEpiphanyOracle)
	if !effectstest.TaLegalForSpec(g, me.ID, spec.Modes.Options[1].Targets)[ability] {
		t.Fatal("the second bullet must offer the ability")
	}
	effectstest.CastModal(t, g, "Sublime Epiphany", "Instant", effectstest.SublimeEpiphanyOracle,
		[]int{1}, []game.TargetRef{effectstest.ModeRef(game.TargetCard, ability, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	if fired {
		t.Error("the countered ability resolved")
	}
}

// Deflecting Swat's "or ability" half (#1196 left it as a caveat):
// the Swat redirects an opponent's TRIGGERED ability, and legality is
// still judged from the ability's point of view.
func TestDeflectingSwatRedirectsAnAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	// A real catalog activation, so the item carries the clause its
	// targets were announced under — which is what the retarget's
	// legality check re-runs. "{T}: Target player mills a card."
	shredder := effectstest.B12Push(g, opp.ID, "Codex Shredder", "Artifact", effectstest.B22CodexShredderOracle, 0, 0)
	if err := g.ActivateCatalogAbility(opp.ID, shredder, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	ability := effectstest.AcItemLabelled(g, "{T}: Target player mills a card.")
	if ability == uuid.Nil {
		t.Fatal("the activation is not on the stack")
	}
	mine, theirs := me.Library.Size(), third.Library.Size()

	if !effectstest.TaLegalTargets(g, me.ID, effectstest.DeflectingSwatOracle)[ability] {
		t.Fatal("Deflecting Swat must offer an ability on the stack")
	}
	effectstest.CastCatalogSpell(t, g, "Deflecting Swat", "Instant", effectstest.DeflectingSwatOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ability}})
	effectstest.PassPriorityAroundTable(t, g)

	prompt := effectstest.LatestRetarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the Swat opened no retarget prompt for the ability")
	}
	if !effectstest.HasID(prompt.PickTargetPlayers, third.ID) || effectstest.HasID(prompt.PickTargetPlayers, me.ID) {
		t.Fatalf("alternatives = %v, want the other seats and not the current target",
			prompt.PickTargetPlayers)
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: third.ID}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	// Reach a main phase BEFORE anything goes on the stack: every
	// advance drains it, and an ability that resolved is an ability
	// nothing can be pointed at.
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	one := effectstest.TaAnnounce(t, g, opp.ID, src, "Their Rock — one target",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}, nil)
	two := effectstest.TaAnnounce(t, g, opp.ID, src, "Their Rock — two targets",
		[]game.TargetRef{
			{Kind: game.TargetPlayer, ID: me.ID},
			{Kind: game.TargetPlayer, ID: g.Seats[2].ID},
		}, nil)

	legal := effectstest.TaLegalTargets(g, me.ID, effectstest.BoltBendOracle)
	if !legal[one] {
		t.Error("a single-target ability is Bolt Bend's whole clause")
	}
	if legal[two] {
		t.Error("a two-target ability has no single target to change")
	}
}
