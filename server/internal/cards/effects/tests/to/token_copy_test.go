package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/h"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #683: a token copy of a `*` creature copies the importer's 0
// stand-in, so it must copy VariableToughness with it. Without it the
// token is a "printed 0/0" to the toughness state check, and it died
// the first time it lost its last counter, where the original stays.
// The control is a token copy of a real printed 0/0, which dies.
func TestTokenCopyOfAStarCreatureSurvivesLosingItsLastCounter(t *testing.T) {
	for _, tc := range []struct {
		name     string
		variable bool
		survives bool
	}{
		{"Star Lhurgoyf", true, true},
		{"Zero Zero Construct", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			src := uuid.New()
			me.Graveyard.PushTop(game.Card{
				InstanceID:        src,
				Name:              tc.name,
				TypeLine:          "Creature — Test",
				VariableToughness: tc.variable,
				Owner:             me.ID,
				Controller:        me.ID,
			})

			g.WithWriteLock(func() {
				_ = effects.CreateTokenCopy{Controller: me.ID, Copy: src, N: 1}.Apply(effects.NewContext(g, nil))
			})
			token := effectstest.FindBattlefieldByName(g, tc.name)
			if token == uuid.Nil {
				t.Fatal("no token copy was created")
			}
			if c, _ := effectstest.BattlefieldCard(g, token); c.VariableToughness != tc.variable {
				t.Errorf("token VariableToughness = %v, want the source's %v", c.VariableToughness, tc.variable)
			}
			effectstest.RunStateChecksViaDraw(t, g)
			if !g.Battlefield.Contains(token) {
				t.Fatal("a never-countered Toughness 0 token copy died; the placeholder skip regressed")
			}

			effectstest.GainAndLoseACounter(t, g, token)
			if got := g.Battlefield.Contains(token); got != tc.survives {
				t.Errorf("after gaining and losing a counter: on battlefield = %v, want %v", got, tc.survives)
			}
		})
	}
}

// TestTokenCopyOfACommanderIsNotACommander — #1363, CR 903.3. The
// commander designation is not one of the copiable values CR 707.2
// lists, and TokenCopyTemplate already builds a fresh game.Card field
// by field rather than copying the source wholesale, so this pins
// that IsCommander is one of the fields deliberately left out — a
// token copy of a commander is just a token, never a commander, and
// so never queues the CR 903.9a "send to command zone" prompt once it
// reaches the graveyard, never gets a commander tax (tokens are never
// cast), and never attributes commander damage (RecordCommanderDamage
// only fires for a battlefield source with IsCommander set).
func TestTokenCopyOfACommanderIsNotACommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	src := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID:  src,
		Name:        "Fixture Commander",
		TypeLine:    "Legendary Creature — Test",
		Power:       2,
		Toughness:   2,
		Owner:       me.ID,
		Controller:  me.ID,
		IsCommander: true,
	})

	g.WithWriteLock(func() {
		if err := (effects.CreateTokenCopy{Controller: me.ID, Copy: src, N: 1}).Apply(effects.NewContext(g, nil)); err != nil {
			t.Fatalf("CreateTokenCopy.Apply: %v", err)
		}
	})
	token := effectstest.FindBattlefieldByName(g, "Fixture Commander")
	if token == uuid.Nil {
		t.Fatal("no token copy was created")
	}
	tokenCard, ok := effectstest.BattlefieldCard(g, token)
	if !ok {
		t.Fatal("token copy not found on battlefield")
	}
	if tokenCard.IsCommander {
		t.Error("token copy of a commander carries IsCommander")
	}

	// Kill it and confirm nothing asks to send it to the command
	// zone — a token has nowhere to go anyway (CR 111.7), but the
	// point is that commanderZoneReplacement's own AppliesTo never
	// even considers it.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(token); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if n := len(g.PendingChoices); n != 0 {
		t.Errorf("%d prompts queued destroying a token copy, want none: %+v", n, g.PendingChoices)
	}
	if n := me.CommanderDamage[token]; n != 0 {
		t.Errorf("CommanderDamage[token] = %d, want 0", n)
	}
}

// An exception that sets a numeric toughness ("except it's a 4/4")
// replaces the `*` with a number, so the token is no stand-in.
func TestTokenCopyExceptionWithANumberClearsVariableToughness(t *testing.T) {
	src := game.Card{Name: "Star Lhurgoyf", TypeLine: "Creature — Lhurgoyf", VariableToughness: true}
	cardsh.HashatonZombieException(&src)
	if src.VariableToughness {
		t.Error("Hashaton's 4/4 exception kept VariableToughness")
	}
}
