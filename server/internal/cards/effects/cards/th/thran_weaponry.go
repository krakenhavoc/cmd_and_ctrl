package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thran Weaponry — Artifact, {4}:
//
//	"Echo {4} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 You may choose not to untap this artifact during your untap step.
//	 {2}, {T}: All creatures get +2/+2 for as long as this artifact remains tapped."
//
// "All creatures get +2/+2" is a continuous effect from a resolving ability,
// so the creatures it affects are locked in as it resolves (CR 611.2c), and
// it lasts for as long as this artifact remains tapped (CR 611.2b). "You may
// choose not to untap" is Amber Prison's untap opt-out.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1af303ef-dd76-4684-97a1-355b7eee53e4",
		Name:         "Thran Weaponry",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{
			effects.MayChooseNotToUntapSelf("Thran Weaponry — you may choose not to untap this artifact"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: All creatures get +2/+2 for as long as this artifact remains tapped.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				d, ok := effects.DurationWhileThisRemainsTapped(ctx)
				if !ok {
					return nil
				}
				return effects.ScopedEffectFor{
					Match:    effects.Creature(),
					Mods:     []game.Mod{game.ModifyPTMod(2, 2)},
					Duration: d,
					Label:    "Thran Weaponry — +2/+2 while tapped",
				}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{effects.Echo("Thran Weaponry", "{4}")},
	})
}
