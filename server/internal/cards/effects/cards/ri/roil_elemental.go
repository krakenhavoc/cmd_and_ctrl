package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roil Elemental — Creature — Elemental {3}{U}{U}{U}, 3/2:
//
//	"Flying
//	 Landfall — Whenever a land you control enters, you may gain control of
//	 target creature for as long as you control this creature."
//
// A landfall trigger with a target chosen as it goes on the stack (CR 603.3d)
// and the "you may" asked as it resolves (CR 603.5). The steal is Dragonlord
// Silumgar's: CR 611.2b's "for as long as you control this creature", so the
// creature goes home when the Elemental leaves and when somebody else takes the
// Elemental. An Elemental that has already left, or changed hands, by the time
// the trigger resolves takes nothing: the duration's condition is false as the
// effect would begin, so it never begins.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7e0d2965-bbcd-4214-b1b6-c0d2601ce078",
		Name:            "Roil Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.Targeting(
					effects.Landfall("Roil Elemental — you may gain control of target creature for as long as you control Roil Elemental (landfall)", roilElementalSteal),
					effects.TargetCreature("target creature")),
				"Roil Elemental — gain control of target creature for as long as you control Roil Elemental?"),
		},
	})
}

func roilElementalSteal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	d, ok := effects.DurationWhileYouControlSource(ctx, item.SourceCardID, item.Controller)
	if !ok {
		return nil
	}
	return effects.GainControl{
		Target:     id,
		Controller: item.Controller,
		Duration:   d,
		Label:      "Roil Elemental — control for as long as you control Roil Elemental",
	}.Apply(ctx)
}
