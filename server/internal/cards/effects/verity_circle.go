package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Verity Circle — Enchantment {2}{U} (Innistrad: Midnight Hunt Commander
// reprint, S82):
//
//	"Whenever a creature an opponent controls becomes tapped, if it isn't
//	 being declared as an attacker, you may draw a card.
//	 {4}{U}: Tap target creature without flying."
//
// "Becomes tapped" for any reason but being declared as an attacker. An
// attack tap emits EventAttack, never EventTapCard, so watching only
// EventTapCard is exactly "not being declared as an attacker": tapping for
// mana, a cost, crewing or an effect all count, and a creature that
// attacks with vigilance was never tapped. "An opponent" is read against
// the Circle's controller as the tap happens. The draw is a "you may"
// (CR 603.5), asked before the trigger goes on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b80b910-c0e6-42e3-929c-2b4c85c850fe",
		Name:         "Verity Circle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OptionalDraw(On(game.EventTapCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventTapCard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.Controller != source.Controller
			}, "Verity Circle — draw a card", Do(DrawCards{N: 1})), "Verity Circle — draw a card?", 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{U}: Tap target creature without flying.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    ManaCost("{4}{U}"),
			Targets: TargetCreature("target creature without flying", WithoutKeyword("flying")),
			Effect:  tapChosenPermanent,
		}},
	})
}
