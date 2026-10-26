package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "92de7cf0-ac3e-46ae-86b2-7d8ebb0276e4",
		Name:         "A Good Day to Pie",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, id := range legalTargetCardIDs(ctx) {
				if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
		Triggered: []game.TriggeredAbility{
			InGraveyard(Optional(
				On(game.EventStickerPlaced, youPutAStickerOfKindOnACreature(game.StickerName),
					"A Good Day to Pie — return this card from your graveyard to your hand",
					returnThisFromYourGraveyardToHand),
				"Return A Good Day to Pie from your graveyard to your hand?")),
		},
	})
}
