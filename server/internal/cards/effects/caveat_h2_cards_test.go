package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3172 cav-h2: cards whose caveat closed on a layer-cache hint
// (StaticAbility.DependsOnEvents) and the other pieces this slice added.

const (
	surveillancePhantasmOracle = "b97648b7-ae98-45c9-8f0f-a52d6faf2d2d"
	whiplashWordsmithOracle    = "e1ffb884-a89e-4e2a-9beb-8bd5838b6a94"
)

func TestStingGivesFirstStrikeToAnAttackerBlockedByAGoblinOrOrc(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		want           bool
	}{
		{"goblin", "Creature — Goblin", true},
		{"orc", "Creature — Orc Warrior", true},
		{"elf", "Creature — Elf", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			sting := pushCatalogPermanent(g, me.ID, "Sting, the Glinting Dagger", "Legendary Artifact — Equipment", stingOracle, false)
			bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
			blocker := b12Creature(g, opp.ID, "Blocker", tc.typeLine, 0, 9)
			brAttach(t, g, sting, bear)
			if effectiveAbilitiesContain(t, g, bear, "first strike") {
				t.Fatal("first strike before any block")
			}
			brAttack(t, g, bear)
			if err := g.DeclareBlocker(blocker, bear); err != nil {
				t.Fatalf("DeclareBlocker: %v", err)
			}
			lockInBlocks(t, g)
			if got := effectiveAbilitiesContain(t, g, bear, "first strike"); got != tc.want {
				t.Errorf("first strike = %v, want %v", got, tc.want)
			}
			if tc.want {
				if err := g.ClearCombat(); err != nil {
					t.Fatalf("ClearCombat: %v", err)
				}
				if effectiveAbilitiesContain(t, g, bear, "first strike") {
					t.Error("first strike survived the end of combat")
				}
			}
		})
	}
}

func TestStingGivesFirstStrikeToABlockerOfAGoblin(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sting := pushCatalogPermanent(g, opp.ID, "Sting, the Glinting Dagger", "Legendary Artifact — Equipment", stingOracle, false)
	goblin := b12Creature(g, me.ID, "Goblin Raider", "Creature — Goblin", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	brAttach(t, g, sting, wall)
	brAttack(t, g, goblin)
	if err := g.DeclareBlocker(wall, goblin); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	if !effectiveAbilitiesContain(t, g, wall, "first strike") {
		t.Error("the equipped blocker lacks first strike while blocking a Goblin")
	}
}

func TestSurveillancePhantasmHasDefenderUntilYouHaveScriedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ph := pushCatalogPermanent(g, me.ID, "Surveillance Phantasm", "Creature — Bird Illusion", surveillancePhantasmOracle, false)
	if !effectiveAbilitiesContain(t, g, ph, "defender") {
		t.Fatal("no defender before any scry")
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventScry, Actor: me.ID})
	})
	if effectiveAbilitiesContain(t, g, ph, "defender") {
		t.Error("still has defender after its controller scried")
	}
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectiveAbilitiesContain(t, g, ph, kw) {
			t.Errorf("lost %s", kw)
		}
	}
}

func TestSurveillancePhantasmCountsAnEarlierSurveilAndIgnoresAnOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventSurveil, Actor: opp.ID})
	})
	ph := pushCatalogPermanent(g, me.ID, "Surveillance Phantasm", "Creature — Bird Illusion", surveillancePhantasmOracle, false)
	if !effectiveAbilitiesContain(t, g, ph, "defender") {
		t.Error("an opponent's surveil opened the Phantasm")
	}
	// A surveil from before the Phantasm entered, by its controller.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventSurveil, Actor: me.ID})
	})
	if effectiveAbilitiesContain(t, g, ph, "defender") {
		t.Error("its controller surveilled this turn; it should be able to attack")
	}
	late := pushCatalogPermanent(g, me.ID, "Surveillance Phantasm", "Creature — Bird Illusion", surveillancePhantasmOracle, false)
	if effectiveAbilitiesContain(t, g, late, "defender") {
		t.Error("a Phantasm that entered after the surveil still has defender")
	}
}

func TestWhiplashWordsmithFliesAndHastesOnceAnOpponentTookNoncombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := pushCatalogPermanent(g, me.ID, "Whiplash Wordsmith", "Creature — Vampire Sorcerer", whiplashWordsmithOracle, false)
	if effectiveAbilitiesContain(t, g, w, "flying") || effectiveAbilitiesContain(t, g, w, "haste") {
		t.Fatal("abilities before any damage")
	}
	src := b12Creature(g, me.ID, "Pinger", "Creature — Wizard", 1, 1)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, me.ID, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if effectiveAbilitiesContain(t, g, w, "flying") {
		t.Error("damage to its own controller turned it on")
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, opp.ID, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	for _, kw := range []string{"flying", "haste"} {
		if !effectiveAbilitiesContain(t, g, w, kw) {
			t.Errorf("no %s after an opponent was dealt noncombat damage", kw)
		}
	}
}
