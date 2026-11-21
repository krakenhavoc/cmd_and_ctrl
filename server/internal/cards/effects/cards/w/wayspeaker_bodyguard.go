package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wayspeaker Bodyguard — Creature — Orc Monk {3}{W}, 3/4 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, return target nonland permanent card with
//	 mana value 2 or less from your graveyard to your hand.
//	 Flurry — Whenever you cast your second spell each turn, tap target
//	 creature an opponent controls."
//
// The enters trigger is Dundoolin Weaver's graveyard return
// (returnTargetGraveyardCardToHand) with the target narrowed to a nonland
// permanent card with mana value 2 or less you own. Flurry is an ability
// word (CR 207.2c): YouCastYourSecondSpellEachTurn, Cori-Steel Cutter's
// condition, with an opponent's creature as the target. Each target is
// checked again as its trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b93eaf66-91b3-44da-851c-79810901be0c",
		Name:         "Wayspeaker Bodyguard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Wayspeaker Bodyguard — return target nonland permanent card with mana value 2 or less from your graveyard to your hand",
				effects.ReturnTargetGraveyardCardToHand),
				effects.TargetCardInGraveyard("target nonland permanent card with mana value 2 or less from your graveyard",
					effects.Permanent(), effects.Nonland(), effects.ManaValueLE(2), effects.YouOwn())),
			effects.Targeting(effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Wayspeaker Bodyguard — flurry: tap target creature an opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target := effects.FirstLegalBattlefieldTarget(ctx)
					if target == uuid.Nil {
						return nil
					}
					return effects.TapTarget{Target: target}.Apply(ctx)
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
