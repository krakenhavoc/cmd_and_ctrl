package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vesuvan Duplimancy — Enchantment {3}{U} (Dominaria United):
//
//	"Whenever you cast a spell that targets only a single artifact or
//	 creature you control, create a token that's a copy of that artifact or
//	 creature, except it's not legendary."
//
// "Targets only a single" is CR 115.9c (every target is the same permanent,
// still on the battlefield, which you controlled as the spell was cast).
// Only your own spells count. The token is created under your control as the
// trigger resolves from the permanent's copiable values (last-known
// information if it has left, CR 608.2h), and "not legendary" is part of the
// copy (CR 707.9b). The token was not cast, so it triggers nothing that
// watches casts, and a noncreature, nonartifact target does not qualify.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6058fab1-9878-42b7-93be-b0a5aa99bd6d",
		Name:         "Vesuvan Duplimancy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast,
				dmuB17CastSpellTargetsOnly(true, func(target game.Card, source *game.Card) bool {
					return (target.IsArtifact() || target.IsCreature()) && target.Controller == source.Controller
				}),
				"Vesuvan Duplimancy — create a token that's a copy of that artifact or creature, except it's not legendary",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					target, ok := g.StackItemSoleTargetForEffect(item.Trigger.Event.CardID)
					if !ok || target.Kind != game.TargetCard {
						return nil
					}
					return CreateTokenCopy{
						Controller: item.Controller,
						Copy:       target.ID,
						N:          1,
						Except: func(t *game.Card) {
							t.TypeLine = removeLegendaryFromTypeLine(t.TypeLine)
						},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
