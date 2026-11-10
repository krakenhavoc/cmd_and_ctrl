package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wanderwine Distracter — Creature — Merfolk Wizard {3}{U}, 4/3 (Lorwyn
// Eclipsed, #4458):
//
//	"Whenever this creature becomes tapped, target creature an opponent
//	 controls gets -3/-0 until end of turn."
//
// "Becomes tapped" is the tap event and an attack that taps it (#4534). The
// target is chosen as the trigger goes on the stack and re-judged as it
// resolves (CR 608.2b); the shrink is a turn-long -3/-0 on that creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6c0f6673-bc24-4c84-b4e7-736283ffd60e",
		Name:         "Wanderwine Distracter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
					"Wanderwine Distracter — target creature an opponent controls gets -3/-0 until end of turn",
					wanderwineDistracterShrink),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

func wanderwineDistracterShrink(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return BoostUntilEOT{Target: id, Power: -3,
		Label: "Wanderwine Distracter — -3/-0 until end of turn"}.Apply(ctx)
}
