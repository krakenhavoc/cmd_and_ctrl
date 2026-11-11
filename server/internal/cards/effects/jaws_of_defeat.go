package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jaws of Defeat — Enchantment {3}{B} (Tarkir: Dragonstorm Commander):
//
//	"Whenever a creature you control enters, target opponent loses life equal
//	 to the difference between that creature's power and its toughness."
//
// Every creature entering under your control triggers it, tokens and the
// creature itself included (CreatureEnteredUnderYourControl), each with its
// own target opponent. The difference is the larger of power and toughness
// minus the smaller (the ruling), read as the ability resolves from the
// creature's current values, or its last-known ones if it has left (CR 608.2h).
// A negative power is used as it is (the difference between -1 and 3 is 4),
// not clamped to zero. An equal power and toughness loses nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46327e52-b04a-4155-bda6-9559e46fb9ed",
		Name:         "Jaws of Defeat",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, CreatureEnteredUnderYourControl,
				"Jaws of Defeat — target opponent loses life equal to the difference between that creature's power and its toughness",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := NewContext(g, item)
					power, toughness, ok := cmdrB13PowerAndToughness(g, item.Trigger.Event.CardID)
					if !ok {
						return nil
					}
					diff := power - toughness
					if diff < 0 {
						diff = -diff
					}
					if diff == 0 {
						return nil
					}
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return g.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -diff)
						}
					}
					return nil
				}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
