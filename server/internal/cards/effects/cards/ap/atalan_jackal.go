package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Atalan Jackal — Creature — Human Tyranid Scout {1}{R}{G}, 2/2:
//
//	"Trample, haste
//	 Skilled Outrider — Whenever this creature deals combat damage to a
//	 player, you may search your library for a basic land card, put it onto
//	 the battlefield tapped, then shuffle."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8be11473-6edf-4f45-b541-11858655fb5f",
		Name:            "Atalan Jackal",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WheneverThisDealsCombatDamageToAPlayer("Atalan Jackal — you may search for a basic land card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Optional:      true,
						Shuffle:       true,
						TappedOnEntry: true,
						Source:        item.SourceCardID,
						Reason:        "Atalan Jackal — search for a basic land card",
					}.Apply(effects.NewContext(g, item))
				}), game.Purpose{Lands: 1}),
		},
	})
}
