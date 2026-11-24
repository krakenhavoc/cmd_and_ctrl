package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b014_helpers.go — shared pieces of the Secret Lair "Partner" cards
// (S82 OLD batch 14).

// AttackAimedAtAnOpponentOfYours is "a player attacks one of your
// opponents": the attack is aimed at a player (not a planeswalker or a battle,
// CR 508.3e) who is still in the game and is not the source's controller.
func AttackAimedAtAnOpponentOfYours(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	opp := AttackedPlayer(g, ev)
	if opp == uuid.Nil || opp == source.Controller {
		return false
	}
	p := g.PlayerByIDForEffect(opp)
	return p != nil && !p.Eliminated
}
