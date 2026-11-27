package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vhati il-Dal — Legendary Creature — Human Warrior {2}{B}{G}, 3/3 (Tempest, S82
// OLD b752, #5628):
//
//	"{T}: Until end of turn, target creature has base power 1 or base
//	 toughness 1."
//
// The choice between the two is made as the ability resolves (CR 608.2d), by the
// ability's controller, after the target is chosen: the prompt's two buttons are
// the two printed alternatives. Either way it is one layer 7b set (CR 613.4b) that
// ends at the cleanup step (CR 514.2) on the creature as it is now (CR 611.2c);
// counters and pumps still apply on top, and the other characteristic is left as
// it was. A target that left or stopped being a creature does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1759547a-52b3-4bae-817a-658a6fd047e3",
		Name:         "Vhati il-Dal",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Until end of turn, target creature has base power 1 or base toughness 1.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.MayChoice{
					Question: "Vhati il-Dal — base power 1 or base toughness 1 until end of turn?",
					YesLabel: "Base power 1",
					NoLabel:  "Base toughness 1",
					Then:     vhatiIlDalAnswer,
					Carry:    []uuid.UUID{id},
				}.Apply(ctx)
			},
		}},
	})
}

// vhatiIlDalAnswer is the choice: "yes" is base power 1, "no" is base toughness 1.
// Keyed (#4291).
var vhatiIlDalAnswer = effects.MayAnswer("confirm/vhati-il-dal-base-pt", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	id := effects.CarryAt(r, 0)
	mod, label := game.SetBaseToughnessMod(1), "Vhati il-Dal — base toughness 1 until end of turn"
	if r.Accepted {
		mod, label = game.SetBasePowerMod(1), "Vhati il-Dal — base power 1 until end of turn"
	}
	return effects.ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{mod},
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    label,
	}.Apply(ctx)
})
