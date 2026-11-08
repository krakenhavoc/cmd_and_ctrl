package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "b93eaf66-91b3-44da-851c-79810901be0c",
		Name:         "Wayspeaker Bodyguard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Wayspeaker Bodyguard — return target nonland permanent card with mana value 2 or less from your graveyard to your hand",
				returnTargetGraveyardCardToHand),
				TargetCardInGraveyard("target nonland permanent card with mana value 2 or less from your graveyard",
					Permanent(), Nonland(), ManaValueLE(2), YouOwn())),
			Targeting(On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Wayspeaker Bodyguard — flurry: tap target creature an opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					target := FirstLegalBattlefieldTarget(ctx)
					if target == uuid.Nil {
						return nil
					}
					return TapTarget{Target: target}.Apply(ctx)
				}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
