package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const squidFireKnightDice = "squid-fire-knight/next-roll-plus-one"

// Squid Fire Knight — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Squid", "Fire", "Knight"; three art stickers.
//	{TK}{TK} — {T}: The next time target player would roll one or more dice this turn, instead they roll that many dice plus one, then you choose one of those rolls to ignore.
//	{TK}{TK}{TK} — Protection from odd mana values
//	{TK}{TK} — 4/1
//	{TK}{TK}{TK}{TK} — 6/6
//
// The first sticker gives an activated ability whose effect is a one-use
// replacement of the target player's next dice roll this turn
// (game.RollExtraDieNextTimeForEffect, CR 614.1a, 706): they roll one more
// die of the same kind, then the ability's controller sees the dice and
// chooses one to ignore (CR 706.6: it never happened, so nothing triggers
// on it). Every card's roll goes through the one dice door, so the rest of
// any dice-rolling card waits for that answer; a card that ignores its own
// lowest roll (Berserker's Frenzy) does that after. When two players'
// shields replace one roll, the roller chooses which applies first (CR
// 616.1). The second sticker gives protection from odd mana values (CR
// 702.16, 202.3).
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5bd19c93-4b1b-45d7-9764-1327fffe94ca",
		Name:         "Squid Fire Knight",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: squidFireKnightDice,
				Activated: []ActivatedAbility{{
					Label:   "{T}: The next time target player would roll one or more dice this turn, instead they roll that many dice plus one, then you choose one of those rolls to ignore.",
					Cost:    TapCost(),
					Targets: TargetPlayer("target player"),
					Effect: func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetPlayer {
								g.RollExtraDieNextTimeForEffect(item.SourceCardID, item.Controller, t.ID, "Squid Fire Knight sticker")
								return nil
							}
						}
						return nil
					},
				}},
				Text: "{T}: The next time target player would roll one or more dice this turn, instead they roll that many dice plus one, then you choose one of those rolls to ignore.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: squidFireKnightDice},
			{Keywords: []string{"protection from odd mana values"}},
		}},
	})
}
