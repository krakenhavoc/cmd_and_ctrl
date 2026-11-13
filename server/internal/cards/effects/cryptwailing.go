package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cryptwailing — Enchantment {3}{B} (Guildpact, #5599):
//
//	"{1}, Exile two creature cards from your graveyard: Target player
//	 discards a card. Activate only as a sorcery."
//
// Dimir Guildmage's sorcery-speed discard row, paid for with Coin of
// Fate's exile of two creature cards from your graveyard, named as the
// ability is activated. The target player chooses the card they discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "01d4c859-8b15-4567-a2c4-70a30c709999",
		Name:         "Cryptwailing",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{1}, Exile two creature cards from your graveyard: Target player discards a card. Activate only as a sorcery.",
			Purpose:      ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
			Cost:         Plus(ManaCost("{1}"), ExileFromGraveyard(2, "two creature cards", MatchCreature)),
			Targets:      TargetPlayer("target player"),
			SorcerySpeed: true,
			Effect: onTheTargetPlayer(func(ctx *Context, p uuid.UUID) error {
				return discardOne(ctx.Game, ctx.Item, p)
			}),
		}},
	})
}
