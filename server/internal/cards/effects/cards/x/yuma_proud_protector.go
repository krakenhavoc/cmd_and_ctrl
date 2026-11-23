package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yuma, Proud Protector:
//
//	"This spell costs {1} less to cast for each land card in your graveyard.
//	 Whenever Yuma enters or attacks, you may sacrifice a land. If you do,
//	 draw a card.
//	 Whenever a Desert card is put into your graveyard from anywhere,
//	 create a 4/2 green Plant Warrior creature token with reach."
//
// The discount is counted as the cost is determined (CR 601.2f). The
// sacrifice is a "may ... if you do": the draw is the sacrifice's
// continuation, paid only for a land that really left. The token trigger
// is for each Desert card, not a batch.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8a0b967-3b5b-46f5-830b-5c8f1037784a",
		Name:         "Yuma, Proud Protector",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.Cmdr2S14LandCardsInYourGraveyard,
				"This spell costs {1} less to cast for each land card in your graveyard."),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Yuma, Proud Protector — you may sacrifice a land; if you do, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Yuma, Proud Protector — sacrifice a land to draw a card?",
						effects.Land(), false, cmdr2S14YumaSacrificed)
				}),
			effects.OnAny(effects.Cmdr2S14GraveyardEvents, effects.Cmdr2S14DesertCardToYourGraveyard,
				"Yuma, Proud Protector — create a 4/2 green Plant Warrior creature token with reach",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("4/2 green Plant Warrior with reach"), N: 1})),
		},
	})
}

// cmdr2S14YumaSacrificed is "If you do, draw a card": keyed.
var cmdr2S14YumaSacrificed = effects.RegisterSacrificedOne("prompt-run/yuma-proud-protector-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.DrawCards{N: 1}.Apply(ctx)
})
