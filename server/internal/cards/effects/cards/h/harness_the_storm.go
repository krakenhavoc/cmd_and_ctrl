package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harness the Storm — Enchantment {2}{R} (Shadows over Innistrad, #5543):
//
//	"Whenever you cast an instant or sorcery spell from your hand, you may
//	 cast target card with the same name as that spell from your graveyard.
//	 (You still pay its costs.)"
//
// The trigger fires on a cast from your hand (YouCastFromYourHand) and
// chooses its target as it goes on the stack: a card in your graveyard that
// shares a name with the spell that was cast (CR 201.2a), read off the spell
// as it was cast. As the ability resolves the card is offered as a paid cast
// from the graveyard (Sproutback Trudge's offer): every cost is paid, and the
// card goes back to the graveyard as usual when it resolves. With no card of
// that name in the graveyard there is no target and the trigger is not put
// on the stack.
//
// No simplification.
func init() {
	trigger := effects.On(game.EventCast, effects.YouCastFromYourHand(effects.Or(effects.Instant(), effects.Sorcery())),
		"Harness the Storm — you may cast target card with the same name as that spell from your graveyard", effects.S82OldB285HarnessOffer("Harness the Storm"))
	trigger.TargetsFrom = effects.S82OldB285SameNameInYourGraveyard
	effects.Register(effects.Spec{
		OracleID:     "38eaf0ea-8cdd-45f4-bfb9-0f8d0fcf863a",
		Name:         "Harness the Storm",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
