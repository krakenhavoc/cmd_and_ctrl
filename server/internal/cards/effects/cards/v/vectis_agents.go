package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vectis Agents — Artifact Creature — Human Rogue {3}{U}{B}, 4/3
// (Conflux, #5587):
//
//	"{U}{B}: This creature gets -2/-0 until end of turn and can't be
//	 blocked this turn."
//
// Ant-Man, Reformed Rogue's blue trigger body as an activated row at
// -2/-0: both halves land on the source only while it is still the
// permanent the ability came from (CR 400.7), and end in the cleanup
// step (CR 514.2). Activating it twice is -4/-0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd9bb758-a763-46bf-9e56-4be26e271741",
		Name:         "Vectis Agents",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{B}: This creature gets -2/-0 until end of turn and can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{U}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: -2, Label: "Vectis Agents — -2/-0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: "Vectis Agents — can't be blocked"}.Apply(ctx)
			},
		}},
	})
}
