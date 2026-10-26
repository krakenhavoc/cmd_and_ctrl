package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const mysticDoomSandwichBlocked = "mystic-doom-sandwich/must-be-blocked"

// Mystic Doom Sandwich — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Mystic", "Doom", "Sandwich"; three art stickers.
//	{TK}{TK} — Lifelink
//	{TK}{TK}{TK} — This creature must be blocked if able. Whenever this creature becomes blocked, it gets +1/+1 until end of turn for each creature blocking it.
//	{TK}{TK} — 1/4
//	{TK}{TK}{TK}{TK}{TK} — 6/8
//
// The first sticker gives lifelink. The second gives two abilities in one bundle: the static "must be blocked if able" (a block requirement, CR 509.1c) and a trigger when the creature becomes blocked, which pumps it by the number of creatures blocking it as the ability resolves.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c490157a-cff5-417b-8f83-15f84398e34f",
		Name:         "Mystic Doom Sandwich",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key:    mysticDoomSandwichBlocked,
				Static: []game.StaticAbility{MustBeBlocked()},
				Triggered: []game.TriggeredAbility{
					On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
						return ev.Kind == game.EventBecomesBlocked && ev.Target == source.InstanceID
					}, "Mystic Doom Sandwich sticker — it gets +1/+1 until end of turn for each creature blocking it",
						func(g *game.Game, item *game.StackItem) error {
							n := len(blockersOf(g, item.SourceCardID))
							if n <= 0 {
								return nil
							}
							return BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n, Label: "Mystic Doom Sandwich sticker"}.Apply(NewContext(g, item))
						}),
				},
				Text: "This creature must be blocked if able. Whenever this creature becomes blocked, it gets +1/+1 until end of turn for each creature blocking it.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Keywords: []string{"lifelink"}},
			{Grant: mysticDoomSandwichBlocked},
		}},
	})
}
