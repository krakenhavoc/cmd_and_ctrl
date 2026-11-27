package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Djinn of Infinite Deceits — Creature — Djinn {4}{U}{U}, 2/7 (Commander 2013,
// S82 OLD b385):
//
//	"Flying
//	 {T}: Exchange control of two target nonlegendary creatures. You can't
//	 activate this ability during combat."
//
// Switcheroo's exchange (CR 701.12, all or nothing: nothing changes hands if
// either target is illegal as the ability resolves; no duration). "During
// combat" is any step of the combat phase, the beginning of combat to the end
// of combat, on any player's turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f9de4cea-27c4-4343-8a7a-09b8f346c3b5",
		Name:            "Djinn of Infinite Deceits",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Exchange control of two target nonlegendary creatures. You can't activate this ability during combat.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("two target nonlegendary creatures", effects.Not(effects.Legendary())).WithCount(2, 2),
			Condition: func(g *game.Game, _, _ uuid.UUID) bool {
				return game.PhaseOf(g.Turn.Step) != game.PhaseCombat
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				var legal []game.TargetRef
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						legal = append(legal, t)
					}
				}
				if len(legal) != 2 {
					return nil
				}
				return effects.ExchangeControl{A: legal[0].ID, B: legal[1].ID, Label: "Djinn of Infinite Deceits — exchange control"}.Apply(ctx)
			},
		}},
	})
}
