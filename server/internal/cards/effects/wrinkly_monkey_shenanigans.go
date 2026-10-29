package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	wrinklyMonkeyShenanigansBlock  = "wrinkly-monkey-shenanigans/cant-block-this"
	wrinklyMonkeyShenanigansMorbid = "wrinkly-monkey-shenanigans/morbid-end-step"
)

// Wrinkly Monkey Shenanigans — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Wrinkly", "Monkey", "Shenanigans"; three art stickers.
//	{TK}{TK} — {1}: Target creature can't block this creature this turn.
//	{TK}{TK}{TK}{TK} — Morbid — At the beginning of each end step, if a creature died this turn and this permanent is a creature, put a -1/-1 counter on this permanent and draw a card.
//	{TK}{TK} — 2/4
//	{TK}{TK}{TK} — 7/2
//
// The first sticker gives an ability that stops a target creature blocking this creature this turn. The second gives morbid (CR 207.2c): at the beginning of each end step, if a creature died this turn and this permanent is a creature (an intervening "if", CR 603.4, checked as it triggers and as it resolves), a -1/-1 counter goes on it and its controller draws a card.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f7016cec-bb17-4f56-bf8e-3d8c83e6e41b",
		Name:         "Wrinkly Monkey Shenanigans",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: wrinklyMonkeyShenanigansBlock,
				Activated: []ActivatedAbility{{
					Label:   "{1}: Target creature can't block this creature this turn.",
					Cost:    ManaCost("{1}"),
					Targets: TargetCreature("target creature"),
					Effect: func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := stickerFirstLegalTarget(ctx)
						if !ok {
							return nil
						}
						return CantBlockThisCreatureThisTurn{Blocker: id, Attacker: item.SourceCardID, Label: "Wrinkly Monkey Shenanigans sticker — can't block it"}.Apply(ctx)
					},
				}},
				Text: "{1}: Target creature can't block this creature this turn.",
			},
			{
				Key: wrinklyMonkeyShenanigansMorbid,
				Triggered: []game.TriggeredAbility{
					On(game.EventBeginEndStep,
						func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
							return source.IsCreature() && b11CreaturesDiedThisTurn(g) > 0
						},
						"Wrinkly Monkey Shenanigans sticker — morbid: put a -1/-1 counter on this permanent and draw a card",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							c, ok := g.LookupCardForEffect(item.SourceCardID)
							if !ok || !c.IsCreature() || b11CreaturesDiedThisTurn(g) == 0 || ctx.isNewSourceObject(item.SourceCardID) {
								return nil
							}
							if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterMinusOne, N: 1}).Apply(ctx); err != nil {
								return err
							}
							return DrawCards{N: 1}.Apply(ctx)
						}),
				},
				Text: "Morbid — At the beginning of each end step, if a creature died this turn and this permanent is a creature, put a -1/-1 counter on this permanent and draw a card.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: wrinklyMonkeyShenanigansBlock},
			{Grant: wrinklyMonkeyShenanigansMorbid},
		}},
	})
}
