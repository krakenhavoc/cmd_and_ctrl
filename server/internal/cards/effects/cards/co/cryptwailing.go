package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "01d4c859-8b15-4567-a2c4-70a30c709999",
		Name:         "Cryptwailing",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}, Exile two creature cards from your graveyard: Target player discards a card. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
			Cost:         effects.Plus(effects.ManaCost("{1}"), effects.ExileFromGraveyard(2, "two creature cards", effects.MatchCreature)),
			Targets:      effects.TargetPlayer("target player"),
			SorcerySpeed: true,
			Effect: effects.OnTheTargetPlayer(func(ctx *effects.Context, p uuid.UUID) error {
				return effects.DiscardOne(ctx.Game, ctx.Item, p)
			}),
		}},
	})
}
