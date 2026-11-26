package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Merfolk Trickster — Creature — Merfolk Wizard {U}{U}, 2/2 (Dominaria, S82
// OLD batch 214, #5530):
//
//	"Flash
//	 When this creature enters, tap target creature an opponent controls. It
//	 loses all abilities until end of turn."
//
// The tap and the loss of abilities are one resolution on the one target,
// re-checked as it resolves (CR 608.2b): a creature that has left or stopped
// being legal is neither tapped nor stripped. The loss is Barrow-Blade's
// layer 6 scoped effect until the end of the turn (CR 514.2), so an ability
// it is given afterwards still works.
//
// No simplification.
func init() {
	const label = "Merfolk Trickster — tap target creature an opponent controls; it loses all abilities until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "f712aed5-d6e0-45d3-ae20-518b765ce9da",
		Name:            "Merfolk Trickster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 || ts[0].Kind != game.TargetCard {
					return nil
				}
				if err := (effects.TapTarget{Target: ts[0].ID}).Apply(ctx); err != nil {
					return err
				}
				return effects.UntilEndOfTurn(ctx, ts[0].ID, nil, label, game.LoseAllAbilitiesMod())
			}), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
