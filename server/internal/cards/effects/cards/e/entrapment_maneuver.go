package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Entrapment Maneuver — Instant {3}{W} (Commander 2016, #5537):
//
//	"Target player sacrifices an attacking creature of their choice. You
//	 create X 1/1 white Soldier creature tokens, where X is that
//	 creature's toughness."
//
// Wing Shards' edict (a creature that is attacking right now, chosen by
// the target player), run keyed so the tokens wait for the sacrifice
// (Consuming Vapors' shape). X is the sacrificed creature's toughness as
// it last existed on the battlefield (CR 608.2h), pumps and counters
// included; a player with no attacking creature sacrifices nothing and
// you create nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2244bef9-1e4e-44ed-9b27-810f04f9e611",
		Name:         "Entrapment Maneuver",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return ctx.SacrificeThen(player,
				effects.SacrificeSpec("an attacking creature", effects.And(effects.Creature(), effects.IsAttackingCreature())),
				"Entrapment Maneuver — sacrifice an attacking creature", 1, entrapmentManeuverSoldiers, player)
		},
	})
}

// entrapmentManeuverSoldiers is "you create X 1/1 white Soldier creature
// tokens, where X is that creature's toughness", once the creature has
// really been sacrificed. The carry is the target player. An on-disk key.
var entrapmentManeuverSoldiers = effects.RunThen("prompt-run/entrapment-maneuver-soldiers", func(ctx *effects.Context, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	ids := r.Sacrifices().By(r.Carry[0])
	if len(ids) == 0 {
		return nil
	}
	x := effects.DepartedCreatureToughness(ctx.Game, ids[0])
	if x <= 0 {
		return nil
	}
	return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 white Soldier"), N: x}.Apply(ctx)
})
