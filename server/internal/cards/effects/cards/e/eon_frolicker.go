package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eon Frolicker — Creature — Elemental Otter {2}{U}{U}, 5/5 (Commander 2020,
// S82 OLD, #5513):
//
//	"Flying
//	 When this creature enters, if you cast it, target opponent takes an
//	 extra turn after this one. Until your next turn, you and planeswalkers
//	 you control gain protection from that player."
//
// "If you cast it" is the intervening if of Rocco, Cabaretti Caterer (CR
// 603.4): the permanent came from a spell. The extra turn is Beacon of
// Tomorrows' (CR 500.7). The protection is Guardian Archon's player keyword,
// protection from that one player (CR 702.16), lasting until your next turn
// (CR 611.2b): which, with the extra turn in between, is your next real turn.
// The planeswalkers are the ones you control as the trigger resolves, locked
// then (CR 611.2c), and gain the same protection for the same duration.
//
// No simplification.
func init() {
	const label = "Eon Frolicker — target opponent takes an extra turn after this one; until your next turn, you and planeswalkers you control gain protection from that player"
	effects.Register(effects.Spec{
		OracleID:        "c0a18a53-d520-4795-9b3b-0f93bcf35848",
		Name:            "Eon Frolicker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.AllOf(effects.Self, effects.GeologicalAppraiserWasCast), label,
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						if err := (effects.TakeExtraTurn{Player: t.ID, N: 1}).Apply(ctx); err != nil {
							return err
						}
						quality := game.ProtectionFromPlayer(t.ID)
						d := effects.DurationUntilYourNextTurn(ctx, ctx.Controller())
						if err := (effects.GainPlayerKeyword{
							Player:   ctx.Controller(),
							Keyword:  quality,
							Label:    "Eon Frolicker — you gain protection from that player",
							Duration: d,
						}).Apply(ctx); err != nil {
							return err
						}
						return effects.ScopedEffectFor{
							Match:    effects.And(effects.Planeswalker(), effects.YouControl()),
							Mods:     []game.Mod{game.AddKeywordsMod(quality)},
							Duration: d,
							Label:    "Eon Frolicker — planeswalkers you control gain protection from that player",
						}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
