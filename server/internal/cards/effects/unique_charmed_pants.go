package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	uniqueCharmedPantsMana = "unique-charmed-pants/tap-any-color"
	uniqueCharmedPantsPump = "unique-charmed-pants/attacks-pump"
)

// Unique Charmed Pants — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Unique", "Charmed", "Pants"; three art stickers.
//	{TK}{TK} — {T}: Add one mana of any color.
//	{TK}{TK}{TK} — Whenever this creature attacks, if it's not a Brushwagg, it gets +X/+0 until end of turn, where X is the number of supertypes, card types, and subtypes it has.
//	{TK}{TK} — 5/1
//	{TK}{TK}{TK}{TK} — 4/8
//
// The first sticker gives a mana ability of any one color. The second gives an attack trigger with an intervening "if" (CR 603.4: "if it's not a Brushwagg", checked as it triggers and as it resolves): X is the number of different supertypes, card types and subtypes the creature has as the ability resolves, a changeling having every creature type (CR 702.73a).
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "989346e5-3e76-40ce-8295-d267929d4fd5",
		Name:         "Unique Charmed Pants",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			TapForManaGrant(uniqueCharmedPantsMana, "{W|U|B|R|G}", "Add one mana of any color", "{T}: Add one mana of any color."),
			{
				Key: uniqueCharmedPantsPump,
				Triggered: []game.TriggeredAbility{{
					Watches: []game.EventKind{game.EventAttack},
					AppliesTo: func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
						return ThisAttacked(ev, source, lki, g) && !source.HasSubtype("Brushwagg")
					},
					Key: "Unique Charmed Pants sticker — if it's not a Brushwagg, it gets +X/+0 until end of turn, X the number of its supertypes, card types and subtypes",
					Effect: func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if ctx.isNewSourceObject(item.SourceCardID) {
							return nil
						}
						c, ok := g.LookupCardForEffect(item.SourceCardID)
						if !ok || c.HasSubtype("Brushwagg") {
							return nil
						}
						return BoostUntilEOT{Target: c.InstanceID, Power: supertypesTypesAndSubtypesOf(c), Label: "Unique Charmed Pants sticker"}.Apply(ctx)
					},
				}},
				Text: "Whenever this creature attacks, if it's not a Brushwagg, it gets +X/+0 until end of turn, where X is the number of supertypes, card types, and subtypes it has.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: uniqueCharmedPantsMana},
			{Grant: uniqueCharmedPantsPump},
		}},
	})
}
