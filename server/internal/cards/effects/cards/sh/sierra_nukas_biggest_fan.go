package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sierra, Nuka's Biggest Fan — Legendary Creature — Human Citizen {3}{W},
// 3/4 (Fallout):
//
//	"The Nuka-Cola Challenge — Whenever one or more creatures you control
//	 deal combat damage to a player, put a quest counter on Sierra and create
//	 a Food token.
//	 Whenever you sacrifice a Food, target creature you control gets +X/+X
//	 until end of turn, where X is the number of quest counters on Sierra."
//
// The first ability is Rapacious Guest's batched trigger (one per damaged
// player, whatever the number of creatures: CR 603.2c). The second is its
// Food read (b31YouSacrificedAFood) with a target; X is the number of quest
// counters on Sierra as it resolves, or as she last had them (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4fb15b6-02bb-4923-953a-69784ca780b8",
		Name:         "Sierra, Nuka's Biggest Fan",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Sierra, Nuka's Biggest Fan — The Nuka-Cola Challenge: put a quest counter on Sierra and create a Food token",
				effects.Do(effects.CounterOnThis{Kind: "quest", N: 1}, effects.CreateToken{Template: effects.FoodToken(), N: 1})),
			effects.Targeting(effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B31YouSacrificedAFood(ev, source, g)
			}, "Sierra, Nuka's Biggest Fan — target creature you control gets +X/+X until end of turn, where X is the number of quest counters on Sierra",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok {
						return nil
					}
					x := 0
					if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && effects.OnBattlefield(g, item.SourceCardID) {
						x = c.Counters["quest"]
					} else if info, ok := g.LastKnownPermanentForEffect(item.SourceCardID); ok {
						x = info.Counters["quest"]
					}
					return effects.BoostUntilEOT{Target: t.ID, Power: x, Toughness: x, Label: "Sierra, Nuka's Biggest Fan — +X/+X"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
