package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stormwing Entity — Creature — Elemental {3}{U}{U}, 3/3 (Core Set 2021,
// #5511):
//
//	"This spell costs {2}{U} less to cast if you've cast an instant or sorcery
//	 spell this turn.
//	 Flying
//	 Prowess
//	 When this creature enters, scry 2."
//
// The reduction takes {2}{U} off the cost as the spell is priced (CR 601.2f)
// when the caster has already cast an instant or sorcery this turn: the
// coloured symbol takes one blue off and the generic part the rest, and nothing
// goes below {0}. The scry is the enters trigger.
//
// No simplification.
func init() {
	unit, err := game.ParseCost("{2}{U}")
	if err != nil {
		panic(err)
	}
	reduction := effects.CostsLess(1, "This spell costs {2}{U} less to cast if you've cast an instant or sorcery spell this turn.",
		func(q game.CostQuery) bool {
			return q.Game != nil && effects.SanarCastAnInstantOrSorceryThisTurn(q.Game, q.Controller, uuid.Nil)
		})
	reduction.Unit = &unit
	effects.Register(effects.Spec{
		OracleID:          "ea43fd04-e9fa-42fb-ba20-22dcc21c6c43",
		Name:              "Stormwing Entity",
		Completeness:      effects.CompletenessFull,
		PrintedKeywords:   []string{"flying", "prowess"},
		Purpose:           game.Purpose{Scry: 2},
		SelfCostModifiers: []game.CostModifier{reduction},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Stormwing Entity — scry 2", effects.Do(effects.Scry{N: 2})),
		},
	})
}
