package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Genju of the Fens — Enchantment — Aura {B} (Betrayers of Kamigawa, S82 OLD b594):
//
//	"Enchant Swamp
//	 {2}: Until end of turn, enchanted Swamp becomes a 2/2 black Spirit creature with
//	 "{B}: This creature gets +1/+1 until end of turn." It's still a land.
//	 When enchanted Swamp is put into a graveyard, you may return this card from your
//	 graveyard to your hand."
//
// Genju of the Falls' shape with a granted activated ability (a bundle in Spec.Grants,
// given by the effect for the turn, ADR 0093): the land's controller pays {B} to pump the
// land, as often as they can.
//
// No simplification.
const genjuOfTheFensGrant = "genju-of-the-fens/pump"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6c9c0e6-b851-4ede-8d7a-190c7cfd9051",
		Name:         "Genju of the Fens",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(effects.Subtype("Swamp")),
		Grants: []effects.AbilityGrant{{
			Key: genjuOfTheFensGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{B}: This creature gets +1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{B}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Genju of the Fens — +1/+1 until end of turn"),
			}},
			Text: "{B}: This creature gets +1/+1 until end of turn.",
		}},
		Activated: []effects.ActivatedAbility{effects.GenjuAbility("Genju of the Fens",
			"{2}: Until end of turn, enchanted Swamp becomes a 2/2 black Spirit creature with \"{B}: This creature gets +1/+1 until end of turn.\" It's still a land.",
			effects.GenjuForm{Colors: []string{"B"}, Power: 2, Toughness: 2, GrantKeys: []string{genjuOfTheFensGrant}})},
		Triggered: []game.TriggeredAbility{effects.GenjuReturnTrigger("Genju of the Fens")},
	})
}
