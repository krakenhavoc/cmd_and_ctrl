package effects

// Eldrazi Guacamole Tightrope — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Eldrazi", "Guacamole", "Tightrope"; three art stickers.
//	{TK}{TK} — Haste
//	{TK}{TK}{TK}{TK}{TK} — You may cast this card from your graveyard by paying 2 life in addition to paying its other costs.
//	{TK}{TK} — 1/4
//	{TK}{TK}{TK} — 5/3
//
// The first sticker gives haste. The second gives a static ability that functions from the graveyard (CR 123.7: an ability sticker gives its ability to a card in any zone, and a sticker stays on a card that goes to a graveyard, CR 123.5): "You may cast this card from your graveyard by paying 2 life in addition to paying its other costs." It is a cast permission derived from the sticker on every query (StickerAbilityDef.GraveyardCastLife), claimed as an offer that charges the card's printed mana cost and 2 life; timing, targets and the rest of the cost are the spell's own, and a land has nothing to cast.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6614c21f-080e-441c-9ebb-c7369c1f52e1",
		Name:         "Eldrazi Guacamole Tightrope",
		Completeness: CompletenessFull,
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Keywords: []string{"haste"}},
			{GraveyardCastLife: 2},
		}},
	})
}
