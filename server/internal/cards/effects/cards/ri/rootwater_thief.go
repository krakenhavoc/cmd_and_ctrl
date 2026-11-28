package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rootwater Thief — Creature — Merfolk Rogue {1}{U}, 1/2 (Nemesis, S82 OLD,
// #5618):
//
//	"{U}: This creature gains flying until end of turn.
//	 Whenever this creature deals combat damage to a player, you may pay {2}. If
//	 you do, search that player's library for a card and exile it, then the
//	 player shuffles."
//
// Roofstalker Wight's flying grant, and Intet's pay-as-it-resolves trigger: the
// damaged player is the damage event's target, their library is the one
// searched (by the Thief's controller, Praetor's Grasp's shape) and shuffled,
// and the card goes to exile face up.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68abaca6-a86a-46e3-b5b8-9860bb6425d2",
		Name:         "Rootwater Thief",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{U}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Rootwater Thief — flying until end of turn"),
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Rootwater Thief — you may pay {2} to search that player's library for a card and exile it", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}",
					Question: "Rootwater Thief — pay {2} to search that player's library for a card and exile it?",
					Then:     rootwaterThiefPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// rootwaterThiefPaid is the paid half, keyed. An on-disk key.
var rootwaterThiefPaid = effects.PayUnlessBranches("pay-unless/s82-old-b678-rootwater-thief-search",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		victim := ctx.Trigger().Event.Target
		if ctx.Game.PlayerByIDForEffect(victim) == nil {
			return nil
		}
		return effects.SearchLibrary{
			Player:       ctx.Controller(),
			LibraryOwner: victim,
			Limit:        1,
			Dest:         game.ZoneExile,
			Shuffle:      true,
			Reason:       "Rootwater Thief — search that player's library for a card and exile it",
			Source:       ctx.Source(),
		}.Apply(ctx)
	}, nil)
