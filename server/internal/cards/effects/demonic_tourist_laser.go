package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	demonicTouristOutlast = "demonic-tourist-laser/outlast-1"
	demonicTouristTickets = "demonic-tourist-laser/dies-seven-tickets"
)

// Demonic Tourist Laser — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Demonic", "Tourist", "Laser"; three art stickers.
//	{TK}{TK} — Outlast {1}
//	{TK}{TK}{TK} — When this permanent dies, you get seven {TK}.
//	{TK}{TK} — 1/4
//	{TK}{TK}{TK}{TK}{TK} — 9/6
//
// Outlast {1} is an activated ability (CR 702.107a): "{1}, {T}: Put a +1/+1 counter on this creature. Activate only as a sorcery." The second sticker gives a dies trigger, judged on the permanent as it last existed (CR 603.10a).
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13d9822f-0398-4915-818b-b9fbaf63b93c",
		Name:         "Demonic Tourist Laser",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: demonicTouristOutlast,
				Activated: []ActivatedAbility{{
					Label:        "Outlast {1} — {1}, {T}: Put a +1/+1 counter on this creature. Activate only as a sorcery.",
					Cost:         Plus(ManaCost("{1}"), TapCost()),
					SorcerySpeed: true,
					Effect: func(g *game.Game, item *game.StackItem) error {
						return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
					},
				}},
				Text: "Outlast {1} ({1}, {T}: Put a +1/+1 counter on this creature. Outlast only as a sorcery.)",
			},
			{
				Key: demonicTouristTickets,
				Triggered: []game.TriggeredAbility{
					WhenThisDies("Demonic Tourist Laser sticker — you get seven {TK}", Do(GetTickets{N: 7})),
				},
				Text: "When this permanent dies, you get seven {TK}.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: demonicTouristOutlast},
			{Grant: demonicTouristTickets},
		}},
	})
}
