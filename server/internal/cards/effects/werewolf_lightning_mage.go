package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	werewolfLightningMageLandfall = "werewolf-lightning-mage/landfall-counter"
	werewolfLightningMageBlocked  = "werewolf-lightning-mage/blocked-shrink"
)

// Werewolf Lightning Mage — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Werewolf", "Lightning", "Mage"; three art stickers.
//	{TK}{TK} — Landfall — Whenever a land enters under your control, put a +1/+1 counter on this permanent.
//	{TK}{TK}{TK}{TK} — Whenever a creature blocks this creature, that creature gets -4/-4 until end of turn.
//	{TK}{TK} — 4/1
//	{TK}{TK}{TK} — 3/5
//
// The first sticker gives landfall (CR 207.2c, an ability word): a +1/+1 counter on the permanent whenever a land enters under its controller's control. The second gives a trigger when a creature blocks this creature: that creature gets -4/-4 until end of turn.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a72927dc-b633-4830-be39-40674ed74ef3",
		Name:         "Werewolf Lightning Mage",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: werewolfLightningMageLandfall,
				Triggered: []game.TriggeredAbility{
					Landfall("Werewolf Lightning Mage sticker — put a +1/+1 counter on this permanent (landfall)",
						func(g *game.Game, item *game.StackItem) error {
							return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
						}),
				},
				Text: "Landfall — Whenever a land enters under your control, put a +1/+1 counter on this permanent.",
			},
			{
				Key: werewolfLightningMageBlocked,
				Triggered: []game.TriggeredAbility{
					On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
						return ev.Kind == game.EventBlock && ev.Target == source.InstanceID && ev.CardID != uuid.Nil
					}, "Werewolf Lightning Mage sticker — the creature that blocked it gets -4/-4 until end of turn",
						func(g *game.Game, item *game.StackItem) error {
							blocker := item.Trigger.Event.CardID
							if !onBattlefield(g, blocker) {
								return nil
							}
							return BoostUntilEOT{Target: blocker, Power: -4, Toughness: -4, Label: "Werewolf Lightning Mage sticker"}.Apply(NewContext(g, item))
						}),
				},
				Text: "Whenever a creature blocks this creature, that creature gets -4/-4 until end of turn.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: werewolfLightningMageLandfall},
			{Grant: werewolfLightningMageBlocked},
		}},
	})
}
