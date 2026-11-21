package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nalfeshnee — Creature — Beast Demon {5}{R}, 4/6 (Commander Legends: Battle for
// Baldur's Gate):
//
//	"Flying
//	 Whenever you cast a spell from exile, copy it. You may choose new targets for
//	 the copy. If it's a permanent spell, the copy gains haste and "At the beginning
//	 of the end step, sacrifice this permanent." (A copy of a permanent spell becomes
//	 a token.)"
//
// The trigger reads a cast from exile off the announce record (OldZone). The copy is
// put on the stack under Nalfeshnee's controller, not cast (CR 707.10), with new
// targets allowed (CR 707.10c). For a permanent spell, the except clause is Choreographed
// Sparks': haste and the quoted trigger are copiable values of the copy (CR 707.9a), so a
// copy that becomes a token (CR 608.3f) keeps them and is sacrificed at the next end step;
// an instant or sorcery copy gains neither.
//
// No simplification.
func init() {
	const grant = "nalfeshnee/sacrifice-this-permanent"
	effects.Register(effects.Spec{
		OracleID:        "97a85a3e-258e-49b7-8ebe-7efee808da7a",
		Name:            "Nalfeshnee",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []effects.AbilityGrant{{
			Key:  grant,
			Text: "At the beginning of the end step, sacrifice this permanent.",
			Triggered: []game.TriggeredAbility{
				effects.AtEachStep(game.StepEnd, "Nalfeshnee — sacrifice this permanent", func(g *game.Game, item *game.StackItem) error {
					return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.GapB220YouCastFromExile,
				"Nalfeshnee — copy that spell; you may choose new targets for the copy; a permanent spell's copy gains haste and a sacrifice trigger",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return effects.CopySpell{
						StackID:          item.Trigger.Event.CardID,
						Controller:       item.Controller,
						ChooseNewTargets: true,
						Except: func(v *game.PrintedValues) {
							if effects.GapB220CopyIsAPermanentSpell(v) {
								v.AddKeyword("haste")
								v.GrantAbility(grant)
							}
						},
					}.Apply(ctx)
				}),
		},
	})
}
