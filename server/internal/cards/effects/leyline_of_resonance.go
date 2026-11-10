package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leyline of Resonance — Enchantment {2}{R}{R} (Duskmourn: House of Horror,
// #4466):
//
//	"If this card is in your opening hand, you may begin the game with it
//	 on the battlefield.
//	 Whenever you cast an instant or sorcery spell that targets only a
//	 single creature you control, copy that spell. You may choose new
//	 targets for the copy."
//
// The opening-hand clause is Spec.OpeningHand (ADR 0133, CR 103.6a), as on
// Leyline of Hope. The trigger is Aziza, Mage Tower Captain's copy with
// Mirrorwing Dragon's "targets only" condition (CR 115.9c): every target of
// the spell, in each instance of the word "target", is one creature, which
// is a creature you control on the battlefield as the spell is cast (a
// spell aimed at the same creature twice still qualifies). The copy is made
// from the spell's last-known information if it has already left the stack
// (CR 608.2h) and is controlled by you with the choice of new targets
// (CR 707.10c). It is created on the stack, not cast, so Leyline does not
// trigger on it again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f69574dd-097f-4908-9cab-344b0eb39c62",
		Name:         "Leyline of Resonance",
		Completeness: CompletenessFull,
		OpeningHand:  BeginTheGameOnTheBattlefield(),
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, AllOf(YouCast(Or(Instant(), Sorcery())), b24SpellTargetsOnlyOneOfYourCreatures),
				"Leyline of Resonance — copy that spell; you may choose new targets for the copy",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return CopySpell{
						StackID:          item.Trigger.Event.CardID,
						Controller:       item.Controller,
						ChooseNewTargets: true,
						FromLastKnown:    true,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
