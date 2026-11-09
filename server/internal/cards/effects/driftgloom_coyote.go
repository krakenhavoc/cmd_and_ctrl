package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Driftgloom Coyote — Creature — Elemental Coyote {3}{W}{W}, 5/3
// (Bloomburrow, #4467):
//
//	"When this creature enters, exile target creature an opponent controls
//	 until this creature leaves the battlefield. If that creature had power
//	 2 or less, put a +1/+1 counter on this creature."
//
// The exile is Oblivion Ring's (CR 610.3). The power is read off the target
// as it was on the battlefield, before the exile (CR 608.2h), and rides the
// exile's keyed tail; the counter is put on only if the creature was
// exiled this way and the Coyote is still on the battlefield.
//
// No simplification.
func init() {
	const label = "Driftgloom Coyote — exile target creature an opponent controls until this creature leaves the battlefield; if it had power 2 or less, put a +1/+1 counter on this creature"
	Register(Spec{
		OracleID:     "f76a86f3-568a-40b3-8e2c-f84db4141293",
		Name:         "Driftgloom Coyote",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					n := 0
					if c, ok := g.LookupCardForEffect(t.ID); ok && c.CurrentPower() <= 2 {
						n = 1
					}
					return ExileUntil{
						Target:     t.ID,
						ThisLeaves: true,
						Label:      "Driftgloom Coyote — the exiled creature returns when Driftgloom Coyote leaves the battlefield",
						ThenTail:   ctx.TailN(driftgloomCoyoteCounter, n),
					}.Apply(ctx)
				}
				return nil
			}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
