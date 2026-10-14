package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gatta and Luzzu — Legendary Creature — Human Soldier {2}{W}, 1/1:
//
//	"Flash
//	 When Gatta and Luzzu enters, choose target creature you control. If
//	 damage would be dealt to that creature this turn, prevent that
//	 damage and put that many +1/+1 counters on it."
//
// Brace for Impact's turn-long shield pinned to the target (ADR 0108 §7),
// with Panther Habit's follow-up: every instance of damage to the
// creature this turn is prevented and that many +1/+1 counters go on it
// (CR 615.5). Damage that can't be prevented is dealt and gives nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c01d6fde-1bd2-4412-932a-9bc97d4e5aef",
		Name:            "Gatta and Luzzu",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Gatta and Luzzu — prevent damage to target creature you control this turn and put that many +1/+1 counters on it",
				func(g *game.Game, item *game.StackItem) error {
					return PreventDamageFromSource{
						Protect: ShieldTheTarget,
						Then:    thatManyCountersOnItBody,
					}.Apply(NewContext(g, item))
				}),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
