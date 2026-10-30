package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m13_helpers.go — shared pieces of Marvel Super Heroes Commander,
// M–Z batch 13 (#4455). Append-only (the clone gate's rule).

// mscM13CastNoncreatureSpellThisTurn is "you've cast a noncreature spell
// this turn": the per-turn cast tally, bumped after a cast succeeds.
func mscM13CastNoncreatureSpellThisTurn(g *game.Game, controller uuid.UUID) bool {
	return g.CastTallyFor(controller).Noncreature > 0
}

// mscM13CombatIfCastNoncreature is "At the beginning of combat on your turn,
// if you've cast a noncreature spell this turn, <effect>" (Mister Fantastic,
// The Thing): an intervening if (CR 603.4), checked as the step begins and
// again as the trigger resolves, so a trigger that was not set off never
// goes on the stack.
func mscM13CombatIfCastNoncreature(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondBoard{
		Name: "you've cast a noncreature spell this turn",
		Test: func(ctx *Context) bool { return mscM13CastNoncreatureSpellThisTurn(ctx.Game, ctx.Controller()) },
	}, On(game.EventStepBegan, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return StepBegan(game.StepBeginCombat, true)(ev, source, game.Characteristic{}, g) &&
			mscM13CastNoncreatureSpellThisTurn(g, source.Controller)
	}, label, effect))
}
