package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// guilePermission is the free-cast grant over the exiled card, or nil.
func guilePermission(g *game.Game, card uuid.UUID) (perm *game.CastPermission) {
	g.WithWriteLock(func() { perm = g.CastPermissionOnCardByIDForEffect(card) })
	return perm
}

// counterABear is the board the headline tests share: seat `me` (the
// active player) casts a Grizzly Bears, `opp` (who controls Guile) lets
// it be and counters it with a real Counterspell.
func counterABear(t *testing.T, guileSeat int) (g *game.Game, me, opp *game.Player, bear, counterspell uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	if guileSeat == 0 {
		effectstest.PushGuile(g, me.ID)
	} else {
		effectstest.PushGuile(g, opp.ID)
	}
	bear = effectstest.CastFromHandForTest(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{})
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	counterspell = effectstest.Batch01OpponentCasts(t, g, opp, "Counterspell", effectstest.B14CounterspellOracle, "{U}{U}",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	return g, me, opp, bear, counterspell
}

func TestGuileIsRegisteredFullWithItsThreeLines(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.GuileOracle)
	if !ok || spec.Name != "Guile" {
		t.Fatalf("Guile is not registered (found %v, name %q)", ok, spec.Name)
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %v, caveats %v: want Full and none", spec.Completeness, spec.Caveats)
	}
	if len(spec.Replacements) != 1 || len(spec.BlockRules) != 1 || len(spec.Triggered) != 1 {
		t.Errorf("replacements %d, block rules %d, triggers %d: want one of each",
			len(spec.Replacements), len(spec.BlockRules), len(spec.Triggered))
	}
}

// The headline: the opponent's Counterspell resolves, the Bears are
// exiled and NOT countered, the Counterspell's controller (not the
// Bears' owner) may cast them for free while it resolves (CR 608.2g,
// #4286), and when they do they enter under that player's control. The
// Counterspell finishes after the cast.
func TestGuileExilesTheCounteredSpellAndTheCountererMayPlayItFree(t *testing.T) {
	g, me, opp, bear, counterspell := counterABear(t, 1)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(bear) || me.Graveyard.Contains(bear) || g.Stack.Contains(bear) {
		t.Fatal("the countered spell is not in exile")
	}
	if !g.Stack.Contains(counterspell) {
		t.Error("the Counterspell finished before the free play was made")
	}
	if n := effectstest.CountEvents(g, game.EventCounterSpell); n != 0 {
		t.Errorf("%d EventCounterSpell: an exiled-instead spell is not countered", n)
	}
	perm := guilePermission(g, bear)
	if perm == nil {
		t.Fatal("no free cast over the exiled Bears")
	}
	if perm.Player != opp.ID || perm.Cost != "{0}" || perm.Zone != game.ZoneExile {
		t.Errorf("grant %+v: want a free cast from exile for Guile's controller", perm)
	}
	effectstest.MustResolutionCastOffer(t, g, opp.ID, game.MayCastKeywordFreeCast, bear)

	// The Bears' owner holds no permission over their own spell.
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("the spell's owner cast an exiled spell they had no permission over")
	}
	// Guile's controller casts them with an empty pool, through the prompt.
	if err := effectstest.CastThroughResolutionCast(t, g, opp.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, bear); c == nil || c.Controller != opp.ID {
		t.Errorf("the Bears did not enter under Guile's controller: %+v", c)
	}
	if !opp.Graveyard.Contains(counterspell) {
		t.Error("the Counterspell itself goes to its owner's graveyard")
	}
}

// "A spell or ability YOU control": Guile on the board of the player
// whose spell is being countered does nothing to the opponent's
// counterspell.
func TestGuileOnTheCounteredSpellsSideChangesNothing(t *testing.T) {
	g, me, _, bear, _ := counterABear(t, 0)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Exile.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Error("the opponent's counterspell was changed by a Guile it does not control")
	}
	if effectstest.CountEvents(g, game.EventCounterSpell) == 0 {
		t.Error("the Bears were not countered")
	}
	if guilePermission(g, bear) != nil {
		t.Error("a free cast was granted for a plain counter")
	}
}

// A spell that can't be countered is not countered, so there is nothing
// for Guile to replace: it resolves, and is not exiled.
func TestGuileDoesNotExileASpellThatCantBeCountered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushGuile(g, opp.ID)
	tyrant := effectstest.CastFromHandForTest(t, g, me, "Carnage Tyrant", "Creature — Dinosaur", "{4}{G}{G}", effectstest.CarnageTyrantOracle, game.CastSpellParams{})
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	effectstest.Batch01OpponentCasts(t, g, opp, "Counterspell", effectstest.B14CounterspellOracle, "{U}{U}",
		[]game.TargetRef{{Kind: game.TargetCard, ID: tyrant}})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(tyrant) || g.Exile.Contains(tyrant) {
		t.Error("an uncounterable spell was exiled by Guile instead of resolving")
	}
	if guilePermission(g, tyrant) != nil {
		t.Error("a free cast was granted over a spell that was never countered")
	}
}

// A "counter unless its controller pays" that is declined is a counter
// (CR 118.12, CR 701.6a): it reaches the window from the prompt's answer.
func TestGuileReplacesADeclinedCounterUnlessPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	effectstest.PushGuile(g, opp.ID)
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastFromSeat(t, g, opp, "Spell Pierce", "Instant", effectstest.S293SpellPierceOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassUntilTaxed(t, g, me.ID)
	if !g.Stack.Contains(victim) {
		t.Fatal("the spell resolved while the {2} was unanswered")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)

	if g.Stack.Contains(victim) || !g.Exile.Contains(victim) || me.Graveyard.Contains(victim) {
		t.Error("the declined tax did not exile the spell instead of countering it")
	}
	if n := effectstest.CountEvents(g, game.EventCounterSpell); n != 0 {
		t.Errorf("%d EventCounterSpell for a spell that was exiled instead", n)
	}
	if perm := guilePermission(g, victim); perm == nil || perm.Player != opp.ID {
		t.Errorf("grant %+v: want a free cast for Spell Pierce's controller", perm)
	}
}

// The destination counters go through the same window: Hinder's
// top-or-bottom choice is asked, and then the spell is exiled instead of
// reaching the library.
func TestGuileReplacesAHinderToTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushGuile(g, me.ID)
	effectstest.SeedLibrary(opp, "Their Top")
	shock := effectstest.CastShockAt(t, g, opp, me.ID)
	effectstest.CastCounterAt(t, g, me, "Hinder", effectstest.HinderOracle, shock)
	effectstest.PassPriorityAroundTable(t, g)

	ask := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if ask == nil {
		t.Fatal("Hinder's top-or-bottom choice was not asked")
	}
	if err := g.ResolvePutInLibrary(ask.ID, me.ID, []uuid.UUID{shock}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(shock) || g.Stack.Contains(shock) {
		t.Fatal("Shock was not exiled")
	}
	for _, c := range opp.Library.Cards {
		if c.InstanceID == shock {
			t.Error("Shock reached its owner's library as well")
		}
	}
	if n := effectstest.CountEvents(g, game.EventCounterSpell); n != 0 {
		t.Errorf("%d EventCounterSpell for a spell that was exiled instead", n)
	}
	if perm := guilePermission(g, shock); perm == nil || perm.Player != me.ID {
		t.Errorf("grant %+v: want a free cast for Hinder's controller", perm)
	}
}

// Two Guiles are one replacement: the first turns the counter into an
// exile and the second has no counter left to replace (CR 616.1f), so
// the spell is exiled once, the grant is one grant, and nobody is asked
// to order two effects that do the same thing.
func TestTwoGuilesExileTheSpellOnce(t *testing.T) {
	g, me, opp, bear, _ := counterABear(t, 1)
	effectstest.PushGuile(g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
			t.Fatalf("two Guiles opened an ordering prompt for %s", c.Chooser)
		}
	}
	if !g.Exile.Contains(bear) || me.Graveyard.Contains(bear) {
		t.Fatal("the Bears are not in exile")
	}
	if perm := guilePermission(g, bear); perm == nil || perm.Player != opp.ID {
		t.Errorf("grant %+v: want exactly the one free cast", perm)
	}
}
