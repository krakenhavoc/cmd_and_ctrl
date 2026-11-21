package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func PushToken(g *game.Game, controller uuid.UUID, tmpl game.Card) uuid.UUID {
	tmpl.InstanceID = uuid.New()
	tmpl.Owner, tmpl.Controller = controller, controller
	g.Battlefield.PushTop(tmpl)
	return tmpl.InstanceID
}

func CardByID(g *game.Game, id uuid.UUID) game.Card {
	c, _ := g.LookupCardForEffect(id)
	return c
}
