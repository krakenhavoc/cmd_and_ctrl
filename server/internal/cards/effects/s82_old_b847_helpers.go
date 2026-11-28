package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b847_helpers.go — shared pieces of Ice Age batch 847 (#5636, S82).

// OldB847TalismanSpec is the whole Spec of one of the Ice Age Talismans:
//
//	"Whenever a player casts a <color> spell, you may pay {3}. If you do, untap
//	 target permanent."
//
// `letter` is the color's mana letter ("G") and `word` its name ("green"). The
// color is read through the Talisman's text (CR 612.2), as Wooden Sphere's is.
// The trigger watches every player's casts and chooses its target as it goes on
// the stack (CR 603.3d: no legal permanent, no trigger); the Talisman's
// controller is asked whether to pay {3} as it resolves (Brass Man's MayPay),
// and a payment untaps the target if it is still a legal one (CR 608.2b).
func OldB847TalismanSpec(oracleID, name, letter, word string) Spec {
	label := name + " — you may pay {3}; if you do, untap target permanent"
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventCast, OldB608CastASpellOfColor(letter), label,
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{3}",
						Question: name + " — pay {3} to untap target permanent?",
						Then:     oldB847TalismanUntap,
					}.Apply(NewContext(g, item))
				}), TargetPermanent("target permanent")),
		},
	}
}

// oldB847TalismanUntap is the paid branch: untap the permanent the trigger
// targeted, if it is still a legal target. An on-disk key: never renamed,
// never reused.
var oldB847TalismanUntap = PayUnlessBranches("pay-unless/old-b847-talisman-untap",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return UntapTarget{Target: t.ID}.Apply(ctx)
			}
		}
		return nil
	}, nil)
