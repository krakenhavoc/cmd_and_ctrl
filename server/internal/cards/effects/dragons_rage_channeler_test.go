package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const dragonsRageChannelerOracle = "0c016ccc-a341-4b76-87ba-69c639d2746d"

// TestDragonsRageChannelerSurveilsOnNoncreatureSpell checks the
// surveil trigger, and that it does NOT fire for a creature spell.
func TestDragonsRageChannelerSurveilsOnNoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Bin Me")
	pushCatalogPermanent(g, me.ID, "Dragon's Rage Channeler", "Creature — Human Shaman",
		dragonsRageChannelerOracle, false)

	castCatalogSpell(t, g, "Shock", "Instant", "shock-test", nil)
	passPriorityAroundTable(t, g)

	if c := surveilChoiceFor(g, me.ID); c == nil {
		t.Fatalf("casting a noncreature spell should surveil 1: %+v", g.PendingChoices)
	}
}

// TestDragonsRageChannelerNoSurveilOnCreatureSpell.
func TestDragonsRageChannelerNoSurveilOnCreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Dragon's Rage Channeler", "Creature — Human Shaman",
		dragonsRageChannelerOracle, false)

	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "bears-test", nil)
	passPriorityAroundTable(t, g)

	if c := surveilChoiceFor(g, me.ID); c != nil {
		t.Error("a creature spell should not trigger the surveil")
	}
}

// TestDragonsRageChannelerDeliriumTurnsOnAndOff pins the CDA-flavoured
// static: four distinct card types in the controller's graveyard
// turns on +2/+2, flying and "attacks each combat if able" together,
// and losing delirium turns them all back off.
func TestDragonsRageChannelerDeliriumTurnsOnAndOff(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	drc := pushCatalogPermanent(g, me.ID, "Dragon's Rage Channeler", "Creature — Human Shaman",
		dragonsRageChannelerOracle, false)

	if effectivePower(t, g, drc) != 1 || effectiveToughness(t, g, drc) != 1 {
		t.Fatalf("no delirium yet: should be the printed 1/1")
	}
	if effectiveAbilitiesContain(t, g, drc, "flying") {
		t.Error("no delirium yet: should not have flying")
	}

	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{Name: "A Land", TypeLine: "Land"})
		me.Graveyard.PushTop(game.Card{Name: "An Instant", TypeLine: "Instant"})
		me.Graveyard.PushTop(game.Card{Name: "A Sorcery", TypeLine: "Sorcery"})
		me.Graveyard.PushTop(game.Card{Name: "An Artifact", TypeLine: "Artifact"})
		g.BumpLayerVersionForTest()
	})

	if effectivePower(t, g, drc) != 3 || effectiveToughness(t, g, drc) != 3 {
		t.Errorf("delirium should be +2/+2: %d/%d, want 3/3",
			effectivePower(t, g, drc), effectiveToughness(t, g, drc))
	}
	if !effectiveAbilitiesContain(t, g, drc, "flying") {
		t.Error("delirium should grant flying")
	}
}
