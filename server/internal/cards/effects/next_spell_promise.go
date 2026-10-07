package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// next_spell_promise.go — the primitive for "the next <kind> spell you
// cast [this turn] <does X>" (#1852). The engine half is
// game/next_spell_promise.go; read its header for when each rider is
// read. A card builds a game.NextSpellPromise (plain data, never a
// closure) and applies it:
//
//	GrantNextSpellPromise{From: "Savage Summoning", Promise: game.NextSpellPromise{
//	    Filter: game.PermissionFilter{CreatureOnly: true},
//	    Flash: true, CantBeCountered: true, CounterKind: "+1/+1", Counters: 1,
//	    Text: "The next creature spell you cast this turn …"}}
//	GrantNextSpellPromise{From: "Hardened Berserker", Promise: game.NextSpellPromise{
//	    Reduce: 1, Text: "The next spell you cast this turn costs {1} less to cast."}}

// GrantNextSpellPromise gives Player (zero: the controller) a one-use
// promise about the next matching spell they cast. It lasts until end
// of turn, which every "this turn" card says; set NoExpiry for one with
// no stated end.
type GrantNextSpellPromise struct {
	Player   uuid.UUID
	From     string
	Promise  game.NextSpellPromise
	NoExpiry bool
}

func (a GrantNextSpellPromise) Apply(ctx *Context) error {
	player := a.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	d := ctx.Game.UntilEndOfTurnDuration()
	if a.NoExpiry {
		d = game.IndefiniteDuration()
	}
	ctx.Game.GrantNextSpellPromiseForEffect(player, a.Promise, a.From, ctx.Source(), d)
	return nil
}
