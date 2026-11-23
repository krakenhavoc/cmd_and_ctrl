package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oracle's Vault — Artifact {4}:
//
//	"{2}, {T}: Exile the top card of your library. Until end of turn,
//	 you may play that card. Put a brick counter on this artifact.
//	 {T}: Exile the top card of your library. Until end of turn, you may
//	 play that card without paying its mana cost. Activate only if there
//	 are three or more brick counters on this artifact."
//
// Urza, Lord High Artificer's exile with a standing permission, as the
// ability resolves: "play" rather than "cast", so a land off the top is
// playable, and the free version prices the play at {0}. The first
// ability puts the brick counter on the Vault as it resolves (it need
// not still be there to have exiled the card). The second is gated on
// the Vault's own brick counters, read as it is activated.
//
// No simplification.
func init() {
	const brick = "brick"
	effects.Register(effects.Spec{
		OracleID:     "220e3298-1663-4731-b50f-247076e1328b",
		Name:         "Oracle's Vault",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}: Exile the top card of your library. Until end of turn, you may play that card. Put a brick counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if _, err := g.ExileTopWithPermissionForEffect(item.Controller, item.Controller, 1, game.CastPermission{
						Duration: g.UntilEndOfTurnDuration(),
					}); err != nil {
						return err
					}
					ctx := effects.NewContext(g, item)
					return effects.AddCounter{Target: ctx.Source(), Kind: brick, N: 1}.Apply(ctx)
				},
			},
			{
				Label:   "{T}: Exile the top card of your library. Until end of turn, you may play that card without paying its mana cost. Activate only if there are three or more brick counters on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Condition: func(g *game.Game, _, source uuid.UUID) bool {
					c, ok := g.LookupCardForEffect(source)
					return ok && c.Counters[brick] >= 3
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					_, err := g.ExileTopWithPermissionForEffect(item.Controller, item.Controller, 1, game.CastPermission{
						Cost:     "{0}",
						Duration: g.UntilEndOfTurnDuration(),
					})
					return err
				},
			},
		},
	})
}
