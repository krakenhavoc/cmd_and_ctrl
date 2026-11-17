package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dimir Guildmage — Creature — Human Wizard {U/B}{U/B}, 2/2 (Ravnica: City
// of Guilds, #5600):
//
//	"({U/B} can be paid with either {U} or {B}.)
//	 {3}{U}: Target player draws a card. Activate only as a sorcery.
//	 {3}{B}: Target player discards a card. Activate only as a sorcery."
//
// Two sorcery-speed rows, each on its own target player, who may be the
// activator. A player with an empty hand discards nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b6c67e5-5913-4db9-85d9-8dae57da02ee",
		Name:         "Dimir Guildmage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:        "{3}{U}: Target player draws a card. Activate only as a sorcery.",
				Purpose:      ForTargets(game.TargetPurpose{Slot: 0, Draws: 1}),
				Cost:         ManaCost("{3}{U}"),
				Targets:      TargetPlayer("target player"),
				SorcerySpeed: true,
				Effect: onTheTargetPlayer(func(ctx *Context, p uuid.UUID) error {
					return DrawCards{Player: p, N: 1}.Apply(ctx)
				}),
			},
			{
				Label:        "{3}{B}: Target player discards a card. Activate only as a sorcery.",
				Purpose:      ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
				Cost:         ManaCost("{3}{B}"),
				Targets:      TargetPlayer("target player"),
				SorcerySpeed: true,
				Effect: onTheTargetPlayer(func(ctx *Context, p uuid.UUID) error {
					return discardOne(ctx.Game, ctx.Item, p)
				}),
			},
		},
	})
}

// onTheTargetPlayer runs `do` on the ability's target player while that
// player is still a legal target (CR 608.2b).
func onTheTargetPlayer(do func(ctx *Context, p uuid.UUID) error) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		p, ok := firstLegalPlayerTarget(ctx)
		if !ok {
			return nil
		}
		return do(ctx, p)
	}
}
