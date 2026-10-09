package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// thoctar_mystic_forge_test.go — #2850: Deathbringer Thoctar and Mystic
// Forge, and the artifact-or-colorless cast-permission filter.

const (
	deathbringerThoctarOracle = "2500a811-2435-4915-ac83-9bfe2887621a"
	mysticForgeOracle         = "994bc16d-fdc6-475c-96df-beeb4d5aa03e"
)

func pushThoctar(g *game.Game, owner uuid.UUID, counters int) uuid.UUID {
	id := pushCatalogPermanent(g, owner, "Deathbringer Thoctar", "Creature — Zombie Beast", deathbringerThoctarOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 3, 3
			}
		}
	})
	if counters > 0 {
		setCounters(g, id, map[string]int{game.CounterPlusOne: counters})
	}
	return id
}

func TestDeathbringerThoctarGrowsOnAnotherCreatureDyingIfYouSayYes(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		thoctar := pushThoctar(g, me.ID, 0)
		bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
		advanceToMain(t, g)

		killCreature(t, g, opp.ID, bear)
		passPriorityAroundTable(t, g)
		if latestTriggerPrompt(g, me.ID) == nil {
			t.Fatal("another creature dying must offer Thoctar's counter")
		}
		answerLatestTriggerPrompt(t, g, me.ID, yes)
		passPriorityAroundTable(t, g)

		want := 0
		if yes {
			want = 1
		}
		if got := counterCount(g, thoctar, game.CounterPlusOne); got != want {
			t.Errorf("answer %v: counters = %d, want %d", yes, got, want)
		}
	}
}

func TestDeathbringerThoctarDoesNotTriggerOnItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	thoctar := pushThoctar(g, me.ID, 0)
	advanceToMain(t, g)

	killCreature(t, g, me.ID, thoctar)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Error("Thoctar's own death offered a counter; the trigger says another creature")
	}
}

func TestDeathbringerThoctarRemovesACounterAtAnnouncementToShootAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thoctar := pushThoctar(g, me.ID, 2)
	advanceToMain(t, g)
	before := lifeOf(g, opp.ID)

	if err := g.ActivateCatalogAbility(me.ID, thoctar, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// The cost is paid before the ability resolves.
	if got := counterCount(g, thoctar, game.CounterPlusOne); got != 1 {
		t.Errorf("counters on the stack = %d, want 1 (paid at announcement)", got)
	}
	if got := lifeOf(g, opp.ID); got != before {
		t.Errorf("life changed before resolution: %d -> %d", before, got)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != before-1 {
		t.Errorf("opponent life = %d, want %d", got, before-1)
	}
}

func TestDeathbringerThoctarCannotActivateWithoutACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thoctar := pushThoctar(g, me.ID, 0)
	advanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, thoctar, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Error("activated with no +1/+1 counter to remove")
	}
}

func TestMysticForgeFilterIsArtifactOrColorlessNonland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushTopSource(t, g, me.ID, "Mystic Forge", "Artifact", mysticForgeOracle)

	for _, tc := range []struct {
		name string
		card game.Card
		want bool
	}{
		{"colorless artifact", game.Card{Name: "Top Rock", TypeLine: "Artifact", ManaCost: "{0}"}, true},
		{"colored artifact creature", game.Card{Name: "Top Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{0}", Colors: []string{"R"}}, true},
		{"colorless non-artifact spell", game.Card{Name: "Top Eldrazi", TypeLine: "Creature — Eldrazi", ManaCost: "{0}"}, true},
		{"colored non-artifact creature", game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{G}"}, false},
		{"colored sorcery", game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{R}"}, false},
		{"basic land (colorless but not a spell)", game.Card{Name: "Top Wastes", TypeLine: "Basic Land"}, false},
		{"artifact land", game.Card{Name: "Top Artifact Land", TypeLine: "Artifact Land"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			top := seedTopOfLibrary(me, tc.card)
			if got := grantedPermissionOn(g, me.ID, top, game.ZoneLibrary).Granted(); got != tc.want {
				t.Fatalf("permission granted = %v, want %v", got, tc.want)
			}
			if !tc.want {
				if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "library"}); err == nil {
					t.Errorf("a card the Forge does not open was cast off the top")
				}
			}
		})
	}
}

func TestMysticForgeCastsAnArtifactFromTheTopAndLooksPrivately(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	pushTopSource(t, g, me.ID, "Mystic Forge", "Artifact", mysticForgeOracle)

	if !topVisibleTo(g, me.ID, me.ID) {
		t.Error("the Forge's controller cannot look at the top card")
	}
	if topVisibleTo(g, me.ID, opp.ID) {
		t.Error("an opponent can see the top card")
	}
	rock := seedTopOfLibrary(me, game.Card{Name: "Top Rock", TypeLine: "Artifact", ManaCost: "{0}"})
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast the artifact off the top: %v", err)
	}
	if me.Library.Contains(rock) {
		t.Error("the cast card is still in the library")
	}
}

func TestMysticForgeExilesTheTopCardPayingTapAndLifeAtAnnouncement(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	forge := pushCatalogPermanent(g, me.ID, "Mystic Forge", "Artifact", mysticForgeOracle, false)
	first := seedTopOfLibrary(me, game.Card{Name: "First", TypeLine: "Creature — Bear", ManaCost: "{G}"})
	lifeBefore := lifeOf(g, me.ID)

	if err := g.ActivateCatalogAbility(me.ID, forge, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !b31Tapped(t, g, forge) {
		t.Error("the Forge is not tapped on announcement")
	}
	if got := lifeOf(g, me.ID); got != lifeBefore-1 {
		t.Errorf("life = %d, want %d (paid at announcement)", got, lifeBefore-1)
	}
	if !me.Library.Contains(first) {
		t.Error("the card left the library before the ability resolved")
	}

	// A different card is on top by the time it resolves: it exiles
	// the CURRENT top card, not the one that was there at activation.
	second := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: second, Owner: me.ID, Controller: me.ID,
		Name: "Second", TypeLine: "Creature — Bear", ManaCost: "{G}"})
	passPriorityAroundTable(t, g)
	if me.Library.Contains(second) {
		t.Error("the current top card was not exiled")
	}
	if !me.Library.Contains(first) {
		t.Error("the former top card was exiled instead of the current one")
	}
	if err := g.ActivateCatalogAbility(me.ID, forge, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the Forge activated while tapped")
	}
}

func TestThoctarAndForgeAreFull(t *testing.T) {
	for _, id := range []string{deathbringerThoctarOracle, mysticForgeOracle} {
		spec, ok := Lookup(id)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v %v, want Full with no caveats", id, spec.Completeness, spec.Caveats)
		}
	}
}
