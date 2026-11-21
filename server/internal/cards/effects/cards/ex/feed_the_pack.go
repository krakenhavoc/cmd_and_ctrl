package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feed the Pack — Enchantment {5}{G} (Jumpstart 2022):
//
//	"At the beginning of your end step, you may sacrifice a nontoken creature.
//	 If you do, create X 2/2 green Wolf creature tokens, where X is the
//	 sacrificed creature's toughness."
//
// Ashnod's "you may sacrifice, if you do" asked as the trigger resolves, over
// the nontoken creatures its controller controls. X is the sacrificed
// creature's toughness as it last existed (CR 608.2h), so an anthem's bonus
// counts; a creature with toughness 0 or less makes no Wolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48981194-6465-4900-8cea-206f8ab5d00e",
		Name:         "Feed the Pack",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Feed the Pack — you may sacrifice a nontoken creature to create Wolves", func(g *game.Game, item *game.StackItem) error {
				return effects.MaySacrificeOne(g, item, "Feed the Pack — sacrifice a nontoken creature? If you do, create X 2/2 Wolves, X being its toughness.",
					effects.And(effects.Creature(), effects.Not(effects.IsTokenPredicate())), false, feedThePackWolves)
			}),
		},
	})
}

// feedThePackWolves is "if you do, create X 2/2 green Wolf tokens". An
// on-disk key: never renamed, never reused.
var feedThePackWolves = effects.RegisterSacrificedOne("prompt-run/feed-the-pack-wolves", func(ctx *effects.Context, sacrificed uuid.UUID) error {
	x := effects.DepartedCreatureToughness(ctx.Game, sacrificed)
	if x <= 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: x}.Apply(ctx)
})
