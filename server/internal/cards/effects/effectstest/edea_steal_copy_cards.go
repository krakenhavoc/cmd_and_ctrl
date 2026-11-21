package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// edea_steal_copy_cards_test.go is the rest of slice E2 (#1565):
// everything but Edea herself, whose tests are in
// edea_steal_cards_test.go with the shared e2 helpers.

// e2Card reads a card wherever it is.
func E2Card(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	var ok bool
	g.ReadSnapshot(func() { c, ok = g.LookupCardForEffect(id) })
	if !ok {
		t.Fatalf("card %s is nowhere", id)
	}
	return c
}

// e2Treasures counts the Treasure tokens `player` controls, and how
// many of them are tapped.
func E2Treasures(g *game.Game, player uuid.UUID) (n, tapped int) {
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == player && c.Name == "Treasure" {
				n++
				if c.Tapped {
					tapped++
				}
			}
		}
	})
	return n, tapped
}

// e2Permission is the cast permission a card carries right now.
func E2Permission(g *game.Game, id uuid.UUID) *game.CastPermission {
	var p *game.CastPermission
	g.WithWriteLock(func() { p = g.CastPermissionOnCardByIDForEffect(id) })
	return p
}

// --- Sephiroth, Fabled SOLDIER // Sephiroth, One-Winged Angel ---------

func SephirothRow() cards.Card {
	row := TransformRow(effects.SephirothOracleID, "Sephiroth, Fabled SOLDIER", "Legendary Creature — Human Avatar Soldier", "{2}{B}",
		"Sephiroth, One-Winged Angel", "Legendary Creature — Angel Nightmare Avatar", "5", "5", []string{"B"})
	row.CardFaces[0].Power, row.CardFaces[0].Toughness = "3", "3"
	return row
}

// --- Hostage Taker -----------------------------------------------------

const HostageTakerOracle = "c5c2d209-e3ef-4b0d-85f5-e7402dcf09eb"

// e2BattlefieldNamed finds a battlefield permanent by name.
func E2BattlefieldNamed(g *game.Game, name string) (game.Card, bool) {
	var out game.Card
	var ok bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				out, ok = c, true
			}
		}
	})
	return out, ok
}

// e2AssertFaceDownFor checks a card was exiled face down with `holder`
// as its only viewer (#1573, game.FaceDownPermitted).
func E2AssertFaceDownFor(t *testing.T, g *game.Game, id, holder uuid.UUID) {
	t.Helper()
	c := E2Card(t, g, id)
	if !c.FaceDown || c.FaceDownKind != game.FaceDownPermitted {
		t.Errorf("card %s: face_down=%v kind=%q, want face down %q", id, c.FaceDown, c.FaceDownKind, game.FaceDownPermitted)
	}
	for _, p := range g.Seats {
		if got, want := c.IsKnownTo(p.ID), p.ID == holder; got != want {
			t.Errorf("card %s: seat %s knows it = %v, want %v", id, p.Name, got, want)
		}
	}
}

// e2Swamps puts n untapped Swamps on `player`'s side.
func E2Swamps(g *game.Game, player uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		g.Battlefield.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp",
			Owner: player, Controller: player,
		})
	}
}

// --- Shatterskull Smashing -----------------------------------------------

const ShatterskullOracle = "78301998-fd9b-4cd5-afad-dbcb43cac2a7"
