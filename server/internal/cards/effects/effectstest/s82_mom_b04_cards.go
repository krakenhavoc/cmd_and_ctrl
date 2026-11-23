package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func S82Library(p *game.Player, cards ...[3]string) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		ids[i] = uuid.New()
		p.Library.PushTop(game.Card{InstanceID: ids[i], Name: cards[i][0], TypeLine: cards[i][1], ManaCost: cards[i][2],
			Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
	}
	return ids
}
