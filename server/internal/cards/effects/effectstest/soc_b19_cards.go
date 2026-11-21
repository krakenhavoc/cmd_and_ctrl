package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// socB19CastWithPool puts a catalog card in the active seat's hand, floats
// `color` x `n` in its pool in a main phase and casts it Strict.
func SocB19CastWithPool(g *game.Game, me *game.Player, c game.Card, color string, n int, params game.CastSpellParams) error {
	c.Owner, c.Controller = me.ID, me.ID
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	me.Hand.PushTop(c)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	FillPoolColored(me, color, n)
	params.Strict = true
	return g.CastSpell(me.ID, c.InstanceID, params)
}

// socB19Named lists the battlefield tokens or cards of that name `owner` controls.
func SocB19Named(g *game.Game, owner uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == name && c.IsToken() {
			out = append(out, c)
		}
	}
	return out
}
