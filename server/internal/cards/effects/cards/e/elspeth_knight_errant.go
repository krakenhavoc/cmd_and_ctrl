package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elspeth, Knight-Errant — Legendary Planeswalker — Elspeth {2}{W}{W}, loyalty 4:
//
//	"+1: Create a 1/1 white Soldier creature token.
//	 +1: Target creature gets +3/+3 and gains flying until end of turn.
//	 −8: You get an emblem with "Artifacts, creatures, enchantments, and lands
//	 you control have indestructible.""
//
// The emblem's static reads the emblem owner's permanents live (CR 114.1,
// CR 702.12b): an artifact, creature, enchantment or land that arrives later
// is covered, and the emblem never leaves.
//
// No simplification.
func init() {
	const emblemText = "Artifacts, creatures, enchantments, and lands you control have indestructible."
	ownersPermanent := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller &&
			(target.IsArtifact() || target.IsCreature() || target.IsEnchantment() || target.IsLand())
	}
	effects.Register(effects.Spec{
		OracleID:        "eab0d8b9-c66f-402f-b259-f46f15cefd78",
		Name:            "Elspeth, Knight-Errant",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &effects.EmblemSpec{
			Label:  "Elspeth, Knight-Errant emblem",
			Text:   emblemText,
			Static: []game.StaticAbility{effects.KeywordGrant(ownersPermanent, "indestructible")},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Create a 1/1 white Soldier creature token.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1}),
			},
			{
				Label:   "+1: Target creature gets +3/+3 and gains flying until end of turn.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit}),
				Effect:  effects.PumpAndGrantFirstTarget(3, 3, "Elspeth, Knight-Errant — +3/+3 and flying", "flying"),
			},
			{
				Label:  "−8: You get an emblem with \"" + emblemText + "\"",
				Cost:   effects.LoyaltyCost(-8),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}
