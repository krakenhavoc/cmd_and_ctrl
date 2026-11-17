package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Death-Rattle Oni — Creature — Demon Spirit {6}{B}, 5/4 (March of the
// Machine: The Aftermath, #5476):
//
//	"Flash
//	 This spell costs {2} less to cast for each creature that died this
//	 turn.
//	 When this creature enters, destroy all other creatures that were dealt
//	 damage this turn."
//
// The discount reads the table-wide count of creatures that died this turn
// (b11CreaturesDiedThisTurn, Mahadi's tally) when the cost is determined
// (CR 601.2f); it takes generic mana only, so the Oni never costs less than
// {B}. The enters trigger destroys every other creature the engine records
// as dealt damage this turn (DealtDamageThisTurn): damage dealt, after
// prevention, and a creature that entered after the damage is a new object
// that was never dealt it (CR 400.7). It reads as a board wipe, so it
// declares its Sweep (partial: it spares undamaged creatures).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bea40c67-73ca-4b5e-b902-5de07f21585c",
		Name:            "Death-Rattle Oni",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(func(q game.CostQuery) int { return 2 * b11CreaturesDiedThisTurn(q.Game) },
				"This spell costs {2} less to cast for each creature that died this turn."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Death-Rattle Oni — destroy all other creatures that were dealt damage this turn",
				func(g *game.Game, item *game.StackItem) error {
					return DestroyAllMatching{
						Match: And(Creature(), OtherThan(item.SourceCardID), DealtDamageThisTurn()),
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
