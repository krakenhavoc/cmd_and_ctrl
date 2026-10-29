package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// edict_keyed.go — EachPlayerSacrifices' "then" as a registered
// continuation (#4290, ADR 0175 amendment of 2026-10-26), so a table
// waiting on any seat's choice of what to sacrifice is a restore point.

// EdictThen is the registered rest of an edict. Hold one only from
// RegisterEdictThen; the zero value is none.
type EdictThen struct{ run game.PromptRunThen }

// RegisterEdictThen registers the "then" of an EachPlayerSacrifices under
// key (ledgered "run <key>": never renamed, never reused). `then` is handed
// a Context rebuilt from the resolving item and one entry per seat that was
// asked, with the permanents that really left. Call it once, from a
// package-level var.
func RegisterEdictThen(key string, then func(ctx *Context, sacrificed game.PromptedSacrifices) error) EdictThen {
	return EdictThen{run: RunThen(key, func(ctx *Context, r game.RunDone) error {
		return then(ctx, r.Sacrifices())
	})}
}
