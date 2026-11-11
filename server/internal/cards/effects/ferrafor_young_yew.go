package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ferrafor, Young Yew — Legendary Creature — Treefolk Druid {6}{G}, 4/7
// (#5368):
//
//	"When Ferrafor enters, create a number of 1/1 green Saproling creature
//	 tokens equal to the number of counters among creatures target player
//	 controls.
//	 {T}: Double the number of each kind of counter on target creature."
//
// The count is every counter of every kind on the creatures the chosen
// player controls, read as the trigger resolves (CR 608.2h); a target player
// who has left the game makes none (CR 608.2b). The tap ability is Vorel of
// the Hull Clade's doubling, aimed at a creature: each kind is read once and
// added again through AddCounter, so Hardened Scales and Doubling Season
// apply. A creature source, so the tap waits out summoning sickness.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a46f9136-b904-403e-b263-e99c9a2fe2ec",
		Name:         "Ferrafor, Young Yew",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Ferrafor — create a number of 1/1 green Saproling tokens equal to the number of counters among creatures target player controls",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						player, ok := firstLegalPlayerTarget(ctx)
						if !ok {
							return nil
						}
						n := 0
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.Controller != player || !c.IsCreature() {
								continue
							}
							for _, v := range c.Counters {
								if v > 0 {
									n += v
								}
							}
						}
						if n <= 0 {
							return nil
						}
						return CreateToken{Template: TokenCard("1/1 green Saproling"), N: n}.Apply(ctx)
					}),
				TargetPlayer("target player")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Double the number of each kind of counter on target creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  b30DoubleCountersOnFirstLegalTarget,
		}},
	})
}
