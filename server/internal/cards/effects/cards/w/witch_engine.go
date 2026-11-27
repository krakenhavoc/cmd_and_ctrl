package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witch Engine — Creature — Horror {5}{B}, 4/4 (Urza's Saga, S82 OLD batch 732,
// #5624):
//
//	"Swampwalk
//	 {T}: Add {B}{B}{B}{B}. Target opponent gains control of this creature.
//	 (Activate only as an instant.)"
//
// It targets, so it is not a mana ability (CR 605.1a): it goes on the stack, and
// the {B}{B}{B}{B} arrive when it resolves (it can be responded to, and the
// mana is not available to pay for anything in response). If the opponent is
// no longer a legal target the whole ability does nothing. On resolution its
// controller adds the mana, then the chosen opponent gains control of the
// Engine for good, as long as it is still the object that was tapped (CR 400.7);
// the new controller gets it with summoning sickness, so can't tap it until
// their next turn begins (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "48adccf6-adb4-49b2-b013-4dc5e0fc7bcb",
		Name:            "Witch Engine",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Add {B}{B}{B}{B}. Target opponent gains control of this creature.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.AddMana{Produced: "{B}{B}{B}{B}"}).Apply(ctx); err != nil {
					return err
				}
				if effects.SourceIsNewObject(g, item) {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						g.GainControlForEffect(item.SourceCardID, item.SourceCardID, t.ID, game.IndefiniteDuration(),
							"Witch Engine — target opponent gains control of this creature")
					}
				}
				return nil
			},
		}},
	})
}
