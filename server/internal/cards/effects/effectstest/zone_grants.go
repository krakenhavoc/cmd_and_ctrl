package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// zone_grants_test.go — #3233, ADR 0093's amendment of 2026-10-13: a
// static that gives an ACTIVATED ABILITY to the cards in a hand or a
// graveyard. Satoru Umezawa's ninjutsu is the headline; Rhet-Tomb
// Mystic, Tectonic Reformation, Sedris and Solemn Doomguide ride the
// same seam. The engine half (the stamp, the readers) is exercised
// through the cards, because the seam is a layer-pass step whose only
// observable output is the rows an object has.

const (
	SatoruUmezawaOracle       = "ed1d21ba-ddfe-4f82-9e37-93f9e2c7ba33"
	RhetTombMysticOracle      = "e445825a-e70e-4796-99be-4a7c4080b8cc"
	TectonicReformationOracle = "10da681b-50bd-4a1b-9934-f0c90e0af55c"
	SedrisOracle              = "48054407-b5e1-431c-ad03-c12a31d0d9a8"
	SolemnDoomguideOracle     = "c3e4a6ce-4a36-43a4-8e80-3c16a9101b79"

	// zgPlainOracle is an oracle ID no card registers, so a hand card
	// that carries it has no abilities of its own: every row it shows
	// is a granted one.
	ZgPlainOracle = "zone-grant-fixture-vanilla"
)

// zgRows reads a card's activated rows and their origins the way every
// consumer does, under the read lock (which recomputes a stale layer
// pass first).
func ZgRows(g *game.Game, id uuid.UUID) (rows []game.ActivatedAbilityShape, origins game.AbilityOrigins) {
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return
		}
		rows, origins = game.ActivatedAbilitiesWithOrigins(c)
	})
	return rows, origins
}

func ZgLabels(rows []game.ActivatedAbilityShape) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.Label
	}
	return out
}

func ZgHandCard(p *game.Player, name, typeLine string) uuid.UUID {
	return PushCatalogHandCard(p, name, typeLine, ZgPlainOracle)
}

func ZgGraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: ZgPlainOracle,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func ZgLeaveBattlefield(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(id) })
	if err != nil {
		t.Fatalf("destroy: %v", err)
	}
}
