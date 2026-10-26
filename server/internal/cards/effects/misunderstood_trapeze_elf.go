package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const misunderstoodTrapezeElfPump = "misunderstood-trapeze-elf/cast-pump"

// Misunderstood Trapeze Elf — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Misunderstood", "Trapeze", "Elf"; three art stickers.
//	{TK}{TK} — Whenever you cast a spell, this creature gets +X/+X until end of turn, where X is the amount of generic mana in that spell's mana cost.
//	{TK}{TK}{TK} — Hexproof
//	{TK}{TK} — 4/2
//	{TK}{TK}{TK}{TK} — 6/6
//
// The first sticker gives a trigger on each spell its controller casts: X is the generic mana in that spell's mana cost, counted from the cost as printed on the spell (a face-down spell has none, CR 708.2a; {X} is not generic mana). The second gives hexproof.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "658252ed-7f52-4a31-834b-c39cee1d5e00",
		Name:         "Misunderstood Trapeze Elf",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: misunderstoodTrapezeElfPump,
				Triggered: []game.TriggeredAbility{
					WheneverYouCast(nil, "Misunderstood Trapeze Elf sticker — it gets +X/+X until end of turn, X the generic mana in that spell's mana cost",
						func(g *game.Game, item *game.StackItem) error {
							spell, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
							if !ok || spell.FaceDown {
								return nil
							}
							cost, err := game.ParseCost(spell.ManaCost)
							if err != nil || cost.Generic <= 0 {
								return nil
							}
							return BoostUntilEOT{Target: item.SourceCardID, Power: cost.Generic, Toughness: cost.Generic, Label: "Misunderstood Trapeze Elf sticker"}.Apply(NewContext(g, item))
						}),
				},
				Text: "Whenever you cast a spell, this creature gets +X/+X until end of turn, where X is the amount of generic mana in that spell's mana cost.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: misunderstoodTrapezeElfPump},
			{Keywords: []string{"hexproof"}},
		}},
	})
}
