package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnoll War Band — Creature — Gnoll {5}{R}, 5/5 (Commander Legends: Battle for
// Baldur's Gate):
//
//	"This spell costs {1} less to cast for each opponent who was dealt damage this
//	 turn.
//	 Menace
//	 Myriad (Whenever this creature attacks, for each opponent other than defending
//	 player, you may create a token copy that's tapped and attacking that player or a
//	 planeswalker they control. Exile the tokens at end of combat.)"
//
// The discount is Bloodsoaked Insight's shape over the turn tally of damage dealt to
// each player (combat or not, any source): one {1} per opponent of the caster dealt
// at least 1 damage, counted as the cost is worked out (CR 601.2f), generic mana only.
// Menace is printed and myriad is the shared trigger (CR 702.116).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ff8fc718-6218-48f3-981d-c4d2bc468493",
		Name:            "Gnoll War Band",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(
				func(q game.CostQuery) int { return effects.GapB219OpponentsDealtDamageThisTurn(q.Game, q.Controller) },
				"This spell costs {1} less to cast for each opponent who was dealt damage this turn."),
		},
		Triggered: []game.TriggeredAbility{effects.GrantedMyriad("Gnoll War Band")},
	})
}
