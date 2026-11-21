package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// put_from_library_test.go — the catalog half of #745: the
// PutFromLibraryOntoBattlefield primitive and the cards that print
// "put [it / a card from among them] onto the battlefield" off the top
// of a library.
//
// The engine half (the move, the batch, the random-order bottom) is
// pinned in server/internal/game/put_from_library_test.go. These tests
// are about each card: what is offered, what enters and how, and where
// the rest go.

const (
	PlChaosWarpOracle      = "07a0cba9-8768-4fd9-a3d5-b0f83b4bf8e8"
	PlCoilingOracleOracle  = "69fd4ddf-9ed8-4c56-bef3-9944daf05e4f"
	PlGenesisWaveOracle    = "e2487868-f386-438e-a73f-b494f6d35fac"
	PlRisenReefOracle      = "2ae71e86-4400-4a30-9077-4d57a43e7395"
	PlExplorersScopeOracle = "a563ede9-b92f-4285-88f8-abcbdd017742"
	PlThrasiosOracle       = "3d867016-2601-4a37-a73d-308898d3bd37"
	PlAnimistsOracle       = "6f1bfe50-b61a-4fdd-963c-40d59117bdf4"
	PlRegaliaOracle        = "27a7610e-acbe-4a2b-9f61-81b383eb20a5"
	PlLurkingOracle        = "15fbb7b1-c62d-4f82-9f35-2c10299779f4"
	PlAtlaPalaniOracle     = "b56cebe0-3752-4ce5-afbd-911543784015"
	PlUreniOracle          = "99c2d3ef-e5b4-48cd-b3f5-de9b02c7c36a"
	PlMajesticOracle       = "15039c85-31e2-4a2b-82f8-2f8270ff9a00"
	PlGilgameshOracle      = "c9432e87-38f6-4889-8f07-ae7cd045e532"
	PlSkyhunterOracle      = "b4dbbf56-d7df-4183-bb48-cfc6b4d0468f"
)

// plTop pushes a card onto the TOP of a player's library.
func PlTop(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// plBottomIDs returns the bottom n library card IDs as a set.
func PlBottomIDs(p *game.Player, n int) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	for _, c := range p.Library.Cards[:n] {
		out[c.InstanceID] = true
	}
	return out
}

func PlSearchFired(g *game.Game, from int) bool {
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventSearchLibrary {
			return true
		}
	}
	return false
}
