package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// closure_doors_restore_test.go — #4253 slice 6: the put and return doors
// and token creation take their "then" as a registered tail, not a closure,
// so a table waiting on the entry question of a card a spell just put onto
// the battlefield, or on the ordering prompt of a token creation, is a
// restore point. Each test takes a STRICT restore (the boot path) while the
// prompt is open, answers it on the restored table, and checks the card
// still finishes: the tail ran off the disk, not off a closure.
//
// The entry question is a shockland's "pay 2 life?" or a Clone's copy
// pick; the ordering prompt is a Doubling Season beside a Parallel Lives.

// libTop pushes a card on top of a player's library and returns its ID.
func LibTop(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.PushTop(c)
	return c.InstanceID
}

// tokenOrderBoard puts a Doubling Season and a Parallel Lives under `me`:
// every token creation then asks for the CR 616 order.
func TokenOrderBoard(g *game.Game, me uuid.UUID) (ds, pl uuid.UUID) {
	ds = PushTokenReplacementCard(g, DoublingSeasonOracle, "Doubling Season", "Enchantment", me)
	pl = PushTokenReplacementCard(g, ParallelLivesOracle, "Parallel Lives", "Enchantment", me)
	return ds, pl
}
