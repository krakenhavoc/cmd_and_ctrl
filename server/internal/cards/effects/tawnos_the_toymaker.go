package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tawnos, the Toymaker — Legendary Creature — Human Artificer {3}{G}{U},
// 3/5 (The Brothers' War):
//
//	"Whenever you cast a Beast or Bird creature spell, you may copy it,
//	 except the copy is an artifact in addition to its other types. (The
//	 copy becomes a token.)"
//
// One optional cast trigger. "Copy it" doesn't target, so the copy is made
// from last-known information when the spell has already left the stack
// (CR 608.2h, 707.10); the copy is controlled by Tawnos's controller
// (CR 707.10) and its exception adds the artifact card type to its
// copiable values (CR 707.10a, 707.9b), which the token it becomes keeps.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48cce741-cfbd-4ed9-a7ac-38880867a286",
		Name:         "Tawnos, the Toymaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, YouCast(And(Creature(), AnySubtype("Beast", "Bird"))),
				"Tawnos, the Toymaker — you may copy the spell, except the copy is an artifact in addition to its other types",
				func(g *game.Game, item *game.StackItem) error {
					return CopySpell{
						StackID:       item.Trigger.Event.CardID,
						Controller:    item.Controller,
						FromLastKnown: true,
						Except:        func(v *game.PrintedValues) { v.AddCardType("Artifact") },
					}.Apply(NewContext(g, item))
				}),
				"Tawnos, the Toymaker — copy that spell?"),
		},
	})
}
