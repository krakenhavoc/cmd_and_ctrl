package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pin Collection — Artifact — Equipment {X}{W}:
//
//	"When this Equipment enters, you may put an ability sticker with ticket
//	 cost X or less on it without paying that sticker's ticket cost.
//	 Equipped creature gets +1/+1 and has all abilities of ability stickers
//	 on this Equipment.
//	 Equip {2}"
//
// X is the X the Equipment was cast with. The sticker is the engine's keyed
// questions with only ability stickers offered, those costing X or fewer
// tickets and whatever tickets the player has, since none is paid (a Free
// limit, ADR 0208 §4).
//
// "Has all abilities of ability stickers on this Equipment" is the ability
// each of those stickers grants to the object it is on (CR 123.7a), written
// onto the equipped creature in layer 6 (CR 613.1f) with that creature as
// "this creature": a keyword, or a bundle (ADR 0093) for a triggered or
// activated ability. It reads the Equipment's stickers as data, so it holds
// when the Equipment is not a creature, and it ends when the Equipment is
// unattached or loses the sticker.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69fedfbb-5d88-4fe0-81e1-1e97a56c6aca",
		Name:         "Pin Collection",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.HasAllAbilitiesOfAbilityStickers(
				"Equipped creature has all abilities of ability stickers on this Equipment.",
				effects.AttachedToSource,
				func(_ *game.Game, source *game.Card) []game.PlacedSticker { return source.Stickers }),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pin Collection — you may put an ability sticker with ticket cost X or less on it without paying that sticker's ticket cost",
				effects.Do(effects.PutASticker{
					On:             effects.StickerOnThis,
					Kinds:          []game.StickerKind{game.StickerAbility},
					MaxCostIsCastX: true,
					Free:           true,
				})),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}
