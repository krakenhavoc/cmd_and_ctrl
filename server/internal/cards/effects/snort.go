package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Snort — Sorcery {3}{R}:
//
//	"Each player may discard their hand and draw five cards. Then Snort
//	 deals 5 damage to each opponent who discarded their hand this way.
//	 Flashback {5}{R}"
//
// Each player is asked in turn order starting with the caster, and each
// finishes their own discard-and-draw before the next is asked. The
// damage is dealt once the last player has answered ("Then"), to each
// opponent of the caster who said yes. An opponent who chose to discard
// an empty hand still counts as having discarded it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "ab94985c-5a3e-42c7-be10-29c5b638346b",
		Name:             "Snort",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return snortAsk(ctx, tablePlayers(ctx), nil)
		},
	})
}

// snortAsk asks the first remaining player, then continues down the
// list; when it is empty the damage goes out to the opponents who said
// yes.
func snortAsk(ctx *Context, players []uuid.UUID, discarded []uuid.UUID) error {
	if len(players) == 0 {
		// One printed instruction: the 5 damage to each of them is one
		// damage instance (CR 615.8).
		return ctx.Game.DamageInstanceForEffect(func() error {
			for _, p := range discarded {
				if p == ctx.Controller() {
					continue
				}
				if err := (DealDamage{Source: ctx.Item.SourceCardID, Target: p, Amount: 5}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	}
	player, rest := players[0], players[1:]
	return MayChoice{
		Player:   player,
		Question: "Snort — discard your hand and draw five cards?",
		Draws:    5,
		OnYes: func(c *Context) error {
			if _, err := discardWholeHand(c.Game, player); err != nil {
				return err
			}
			if err := c.Game.DrawNForEffect(player, 5); err != nil {
				return err
			}
			return snortAsk(c, rest, append(discarded, player))
		},
		OnNo: func(c *Context) error {
			return snortAsk(c, rest, discarded)
		},
	}.Apply(ctx)
}
