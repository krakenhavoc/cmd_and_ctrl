package cardsg

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// General Tazri — Legendary Creature — Human Ally {4}{W}, 3/4 (Oath of the
// Gatewatch, #5545):
//
//	"When General Tazri enters, you may search your library for an Ally
//	 creature card, reveal it, put it into your hand, then shuffle.
//	 {W}{U}{B}{R}{G}: Ally creatures you control get +X/+X until end of turn,
//	 where X is the number of colors among those creatures."
//
// The search is Micromancer's optional, revealed tutor. The pump locks both
// the set of Allies and X as the ability resolves (CR 611.2c): X is the number
// of distinct colors among the Ally creatures you control then, General Tazri
// herself included (she is white), and a colorless Ally adds none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0f19cba-1339-4518-8320-d7b1dcaf2eb0",
		Name:         "General Tazri",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("General Tazri — search for an Ally creature card", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsCreature() && c.HasSubtype(ctx.TextSubtype("Ally")) },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "General Tazri — an Ally creature card, revealed, to hand",
				}.Apply(ctx)
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{U}{B}{R}{G}: Ally creatures you control get +X/+X until end of turn, where X is the number of colors among those creatures.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{W}{U}{B}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				allies := effects.And(effects.OfCreatureType("Ally"), effects.YouControl())
				colors := map[string]bool{}
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller != ctx.Controller() || !allies(g, ctx.Controller(), c) {
						continue
					}
					for _, col := range c.EffectiveColors() {
						colors[strings.ToUpper(col)] = true
					}
				}
				x := len(colors)
				if x == 0 {
					return nil
				}
				return effects.BoostUntilEOT{Match: allies, Power: x, Toughness: x,
					Label: "General Tazri — Ally creatures you control get +X/+X"}.Apply(ctx)
			},
		}},
	})
}
