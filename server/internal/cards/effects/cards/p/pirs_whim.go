package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pir's Whim — Sorcery {3}{G} (Battlebond):
//
//	"For each player, choose friend or foe. Each friend searches their library
//	 for a land card, puts it onto the battlefield tapped, then shuffles. Each
//	 foe sacrifices an artifact or enchantment of their choice."
//
// Zndrsplt's Judgment's friend-or-foe question, one player at a time starting
// with the caster; then each friend searches (their own prompt, a land card or
// nothing, CR 701.23b) and each foe chooses an artifact or enchantment they
// control to sacrifice. A foe with neither does nothing.
//
// No simplification.
func init() {
	effects.RegisterOldB206FriendFoe(effects.OldB206FriendFoe{
		Name: "Pir's Whim",
		Friend: func(ctx *effects.Context, who uuid.UUID) error {
			return effects.SearchLibrary{
				Player:        who,
				Predicate:     effects.IsLandCard,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Pir's Whim — a land card onto the battlefield tapped",
			}.Apply(ctx)
		},
		Foe: func(ctx *effects.Context, who uuid.UUID) error {
			effects.OldB206PickToSacrifice(ctx, who, "Pir's Whim — sacrifice an artifact or enchantment",
				effects.Or(effects.Artifact(), effects.Enchantment()))
			return nil
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "01c2b66a-66d1-4282-a078-b1fa737c57f4",
		Name:         "Pir's Whim",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.OldB206ChooseFriendOrFoe(ctx, "Pir's Whim")
		},
	})
}
