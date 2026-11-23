package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// desertion_test.go — #5931: Desertion counters a spell and, when it is an
// artifact or creature spell, puts that card onto the battlefield under
// Desertion's controller instead of into its owner's graveyard, through
// the engine's entry pipeline (game/counter_onto_battlefield.go).

const (
	desertionOracle          = "ee3d1f44-e0ca-4ce9-be76-4b675a115156"
	desertionVisionaryOracle = "c6a3a882-a127-4590-93d7-679ef4313efe"
)

// desertionAgainst has the active player `me` cast `name` and the next
// seat `opp` answer it with Desertion. Nothing has resolved yet.
func desertionAgainst(t *testing.T, name, typeLine, cost, oracle string) (g *game.Game, me, opp *game.Player, spell, desertion uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	spell = effectstest.CastFromHandForTest(t, g, me, name, typeLine, cost, oracle, game.CastSpellParams{})
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	desertion = effectstest.Batch01OpponentCasts(t, g, opp, "Desertion", desertionOracle, "{3}{U}{U}",
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	return g, me, opp, spell, desertion
}

func TestDesertionIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(desertionOracle)
	if !ok || spec.Name != "Desertion" {
		t.Fatalf("Desertion is not registered (found %v, name %q)", ok, spec.Name)
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %v, caveats %v: want Full and none", spec.Completeness, spec.Caveats)
	}
}

// The headline: a creature spell is countered (EventCounterSpell fires)
// and enters under Desertion's controller, not into its owner's
// graveyard, and its own ETB trigger fires for its new controller.
func TestDesertionPutsACounteredCreatureOntoTheBattlefieldUnderYourControl(t *testing.T) {
	g, me, opp, elf, desertion := desertionAgainst(t, "Elvish Visionary", "Creature — Elf Shaman", "{1}{G}", desertionVisionaryOracle)
	oppHand, myHand := len(opp.Hand.Cards), len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.FindBattlefieldCardForTest(g, elf)
	if c == nil {
		t.Fatal("the countered Elvish Visionary is not on the battlefield")
	}
	if c.Controller != opp.ID || c.Owner != me.ID {
		t.Errorf("controller %v owner %v: want Desertion's controller and the original owner", c.Controller, c.Owner)
	}
	if me.Graveyard.Contains(elf) || g.Stack.Contains(elf) {
		t.Error("the card is also in its owner's graveyard or still on the stack")
	}
	if n := effectstest.CountEvents(g, game.EventCounterSpell); n != 1 {
		t.Errorf("%d EventCounterSpell: the spell was countered once", n)
	}
	if effectstest.CountEvents(g, game.EventResolve) == 0 {
		t.Fatal("nothing resolved")
	}
	if !opp.Graveyard.Contains(desertion) {
		t.Error("Desertion did not go to its owner's graveyard")
	}
	if got := len(opp.Hand.Cards); got != oppHand+1 {
		t.Errorf("Desertion's controller hand %d, want %d: the ETB draw goes to the new controller", got, oppHand+1)
	}
	if got := len(me.Hand.Cards); got != myHand {
		t.Errorf("the owner's hand %d, want %d: the owner draws nothing", got, myHand)
	}
}

// An artifact spell is the other half of "artifact or creature".
func TestDesertionTakesAnArtifactSpell(t *testing.T) {
	g, me, opp, ring, _ := desertionAgainst(t, "Sol Ring", "Artifact", "{1}", effectstest.SolRingOracle)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.FindBattlefieldCardForTest(g, ring)
	if c == nil || c.Controller != opp.ID {
		t.Fatalf("Sol Ring %+v: want it on the battlefield under Desertion's controller", c)
	}
	if me.Graveyard.Contains(ring) {
		t.Error("Sol Ring is also in its owner's graveyard")
	}
}

// Any other spell is countered into its owner's graveyard.
func TestDesertionCountersAnInstantIntoItsOwnersGraveyard(t *testing.T) {
	g, me, _, bolt, _ := desertionAgainst(t, "Their Instant", "Instant", "{R}", "")
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(bolt) || g.Battlefield.Contains(bolt) {
		t.Error("an instant countered by Desertion is not in its owner's graveyard")
	}
	if n := effectstest.CountEvents(g, game.EventCounterSpell); n != 1 {
		t.Errorf("%d EventCounterSpell, want 1", n)
	}
}

// A spell that can't be countered is not countered, so there is nothing
// to put anywhere: it resolves under its own controller.
func TestDesertionDoesNotTakeASpellThatCantBeCountered(t *testing.T) {
	g, me, _, tyrant, _ := desertionAgainst(t, "Carnage Tyrant", "Creature — Dinosaur", "{4}{G}{G}", effectstest.CarnageTyrantOracle)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.FindBattlefieldCardForTest(g, tyrant)
	if c == nil || c.Controller != me.ID {
		t.Fatalf("Carnage Tyrant %+v: want it resolved under its caster's control", c)
	}
	if n := effectstest.CountEvents(g, game.EventCounterSpell); n != 0 {
		t.Errorf("%d EventCounterSpell for a spell that can't be countered", n)
	}
}

// The card ENTERS the battlefield, so the CR 614 entry window runs on it
// with its new controller already stamped: the original owner's
// Authority of the Consuls ("Creatures your opponents control enter
// tapped") taps it, because it enters under an opponent's control.
func TestDesertionsEntryRunsTheEntryWindowUnderTheNewController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	_ = effectstest.SeedReplacementPermanent(g, effectstest.AuthorityOfTheConsulsOracle, "Authority of the Consuls", me.ID)
	bear := effectstest.CastFromHandForTest(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{})
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	effectstest.Batch01OpponentCasts(t, g, opp, "Desertion", desertionOracle, "{3}{U}{U}",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.FindBattlefieldCardForTest(g, bear)
	if c == nil || c.Controller != opp.ID {
		t.Fatalf("Grizzly Bears %+v: want them under Desertion's controller", c)
	}
	if !c.Tapped {
		t.Error("Authority of the Consuls did not tap a creature entering under its controller's opponent")
	}
}

// The counter's own replacement window still comes first: Desertion's
// controller's Guile exiles the spell instead of countering it, so it
// never reaches the battlefield.
func TestDesertionUnderGuileExilesInstead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushGuile(g, opp.ID)
	bear := effectstest.CastFromHandForTest(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{})
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	effectstest.Batch01OpponentCasts(t, g, opp, "Desertion", desertionOracle, "{3}{U}{U}",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(bear) || g.Battlefield.Contains(bear) {
		t.Error("Guile's exile-instead did not replace Desertion's counter")
	}
	if n := effectstest.CountEvents(g, game.EventCounterSpell); n != 0 {
		t.Errorf("%d EventCounterSpell for a spell exiled instead of countered", n)
	}
}

// Negative: Desertion targets spells only. With a pool that pays for it,
// a permanent on the battlefield is refused as its target, and a spell on
// the stack is accepted.
func TestDesertionTargetsSpellsNotPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	standing := effectstest.PushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	onStack := effectstest.CastFromHandForTest(t, g, me, "Other Bears", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{})
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Desertion", TypeLine: "Instant", OracleID: desertionOracle,
			ManaCost: "{3}{U}{U}", Owner: me.ID, Controller: me.ID})
		effectstest.FillPoolColored(me, "U", 2)
		effectstest.FillPool(me, 3)
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: standing}}}); err == nil {
		t.Fatal("Desertion was cast targeting a permanent")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: onStack}}}); err != nil {
		t.Fatalf("Desertion targeting a spell: %v", err)
	}
}
