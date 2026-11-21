package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Korvold, Gleeful Glutton — Legendary Creature — Dragon Noble {5}{B}{R}{G},
// 4/4: "This spell costs {1} less to cast for each card type among permanents
// you've sacrificed this turn. Flying, trample, haste. Whenever Korvold deals
// combat damage to a player, put X +1/+1 counters on Korvold and draw X
// cards, where X is the number of permanent types among cards in your
// graveyard."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "99e89098-bc55-4297-9407-ab9256b609da",
		Name:            "Korvold, Gleeful Glutton",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", "haste"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int {
				return effects.Cmdr2S52CardTypesSacrificedThisTurn(q.Game, q.Controller)
			}, "This spell costs {1} less to cast for each card type among permanents you've sacrificed this turn."),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Korvold — put X +1/+1 counters on it and draw X cards", effects.Cmdr2S52KorvoldGrow),
		},
	})
}
