package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// A Good Day to Pie — Instant {1}{W}:
//
//	"Tap up to two target creatures.
//	 Whenever you put a name sticker on a creature, you may return this
//	 card from your graveyard to your hand."
//
// The second ability works only from the graveyard (CR 113.6k); see Unlawful
// Entry. "Up to two" is a zero-to-two target clause, and each target still
// legal as the spell resolves is tapped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92de7cf0-ac3e-46ae-86b2-7d8ebb0276e4",
		Name:         "A Good Day to Pie",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, id := range effects.LegalTargetCardIDs(ctx) {
				if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.Optional(
				effects.On(game.EventStickerPlaced, effects.YouPutAStickerOfKindOnACreature(game.StickerName),
					"A Good Day to Pie — return this card from your graveyard to your hand",
					effects.ReturnThisFromYourGraveyardToHand),
				"Return A Good Day to Pie from your graveyard to your hand?")),
		},
	})
}
