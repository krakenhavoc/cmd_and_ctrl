package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	coolFluffyLoxodonDraw    = "cool-fluffy-loxodon/leaves-draw"
	coolFluffyLoxodonAnimate = "cool-fluffy-loxodon/becomes-13-13"
)

// Cool Fluffy Loxodon — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Cool", "Fluffy", "Loxodon"; three art stickers.
//	{TK}{TK} — When this permanent leaves the battlefield, draw a card.
//	{TK}{TK}{TK}{TK}{TK} — Whenever a creature enters under your control, this permanent becomes a 13/13 Eldrazi creature in addition to its other types until end of turn.
//	{TK}{TK} — 4/2
//	{TK}{TK}{TK}{TK} — 5/6
//
// The first sticker gives a leaves-the-battlefield draw (the trigger is judged on the permanent as it last existed, CR 603.10a). The second animates the permanent for the turn: it becomes a 13/13 Eldrazi creature in addition to its other types (CR 205.1b), a layer 4 and 7b effect that ends at cleanup, as each creature enters under its controller.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba2bb276-b7ef-46ec-9618-2cf4d60c70a6",
		Name:         "Cool Fluffy Loxodon",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: coolFluffyLoxodonDraw,
				Triggered: []game.TriggeredAbility{
					WhenThisLeaves("Cool Fluffy Loxodon sticker — draw a card", Do(DrawCards{N: 1})),
				},
				Text: "When this permanent leaves the battlefield, draw a card.",
			},
			{
				Key: coolFluffyLoxodonAnimate,
				Triggered: []game.TriggeredAbility{
					On(game.EventETB, CreatureEnteredUnderYourControl,
						"Cool Fluffy Loxodon sticker — it becomes a 13/13 Eldrazi creature until end of turn",
						func(g *game.Game, item *game.StackItem) error {
							return BecomeCreatureUntilEOT{
								Target: item.SourceCardID, Types: []string{"Creature"}, Subtypes: []string{"Eldrazi"},
								SetPower: 13, SetToughness: 13,
								Label: "Cool Fluffy Loxodon sticker — 13/13 Eldrazi creature",
							}.Apply(NewContext(g, item))
						}),
				},
				Text: "Whenever a creature enters under your control, this permanent becomes a 13/13 Eldrazi creature in addition to its other types until end of turn.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: coolFluffyLoxodonDraw},
			{Grant: coolFluffyLoxodonAnimate},
		}},
	})
}
