package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "6c0f6673-bc24-4c84-b4e7-736283ffd60e",
		Name:         "Wanderwine Distracter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
					"Wanderwine Distracter — target creature an opponent controls gets -3/-0 until end of turn",
					wanderwineDistracterShrink),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}

func wanderwineDistracterShrink(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.BoostUntilEOT{Target: id, Power: -3,
		Label: "Wanderwine Distracter — -3/-0 until end of turn"}.Apply(ctx)
}
