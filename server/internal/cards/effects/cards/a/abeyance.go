package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abeyance — Instant {1}{W} (Weatherlight, S82 older sets batch 777, #5629):
//
//	"Until end of turn, target player can't cast instant or sorcery spells, and that player can't
//	 activate abilities that aren't mana abilities.
//	 Draw a card."
//
// Two player-wide bans until end of turn on the target: an outright cast ban over instant and
// sorcery spells (as Sphinx's Decree's) and an activation ban that spares mana abilities
// (CR 605.1a, game.GrantCantActivateNonManaAbilitiesForEffect). The draw is the caster's and
// happens even if the target is gone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6300de53-6e71-4f0e-87e1-b0acd25c59a8",
		Name:         "Abeyance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if player := effects.FirstLegalTargetPlayer(ctx); player != uuid.Nil {
				if err := (effects.RestrictCasting{
					Player: player,
					Rule: game.CastBanRule{
						Kind:   game.CastBanOutright,
						Filter: game.PermissionFilter{InstantOrSorceryOnly: true},
					},
					Label:    "Abeyance — can't cast instant or sorcery spells this turn",
					Duration: effects.DurationUntilEndOfTurn(ctx),
				}).Apply(ctx); err != nil {
					return err
				}
				ctx.Game.GrantCantActivateNonManaAbilitiesForEffect(player,
					"Abeyance — can't activate abilities that aren't mana abilities this turn", item.SourceCardID, ctx.Game.UntilEndOfTurnDuration())
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
