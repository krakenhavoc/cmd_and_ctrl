package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushSizedCreatureForTest seeds a creature with a size, a mana cost,
// colours and a subtype line.
func PushSizedCreatureForTest(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string, colors []string, p, t int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		ManaCost:   manaCost,
		Colors:     colors,
		Power:      p,
		Toughness:  t,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

func Exiled(g *game.Game, id uuid.UUID) bool { return g.Exile.Contains(id) }
