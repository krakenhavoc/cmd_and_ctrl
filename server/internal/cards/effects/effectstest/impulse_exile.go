package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// impulse_exile_test.go — S21 sub-PR 6. Both cards hang off combat
// damage, so the tests declare an attack and let the damage step
// drain the trigger, the same shape combat_triggers_test.go uses.

const (
	RagavanOracle  = "37108cd4-bbab-4ce3-9ed6-f60e8422e703"
	BreechesOracle = "eb77f7dc-e9e4-44ef-8616-9f4e737e8ca5"
)

// stackLibrary replaces a player's library with one named card on
// top, so the test knows exactly what an impulse trigger will steal.
func StackLibrary(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	p.Library.Cards = nil
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// exiledPermission returns the impulse grant on a card in exile.
func ExiledPermission(g *game.Game, id uuid.UUID) *game.CastPermission {
	if perm := g.CastPermissionOnCardByIDForEffect(id); perm != nil {
		return perm
	}
	return &game.CastPermission{}
}

// permissionLive asks the engine's ONE liveness predicate (#945): the
// permission names this player, its CR 702.185a floor has been
// reached, and its CR 611.2 duration has not run out. Card tests ask
// it rather than reading a window field, so they cannot disagree with
// the cast path about what "live" means.
func PermissionLive(g *game.Game, perm *game.CastPermission, player uuid.UUID) bool {
	return g.CastPermissionActiveForEffect(perm, player)
}
