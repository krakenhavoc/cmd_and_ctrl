package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vengeful Pharaoh — Creature — Zombie {2}{B}{B}{B}, 5/4 (Magic 2012, S82
// older sets batch 451, #5573):
//
//	"Deathtouch
//	 Whenever combat damage is dealt to you or a planeswalker you control,
//	 if this card is in your graveyard, destroy target attacking creature,
//	 then put this card on top of your library."
//
// A graveyard-zone trigger (InGraveyard: "you" is the card's owner while it
// sits there, CR 108.4), once per batch of simultaneous combat damage
// (CR 603.2c). The intervening "if this card is in your graveyard" is
// rechecked as it resolves (CR 603.4): a Pharaoh that left in response
// neither destroys nor goes to the library.
//
// No simplification.
func init() {
	pharaoh := effects.OncePerBatch(effects.InGraveyard(effects.On(game.EventDealDamage, effects.S82OldB451DamageToYouOrYourPlaneswalker,
		"Vengeful Pharaoh — destroy target attacking creature, then put this card on top of your library",
		effects.S82OldB451PharaohRises)))
	pharaoh.Targets = effects.TargetCreature("target attacking creature", effects.AttackingCreature())
	effects.Register(effects.Spec{
		OracleID:        "7e465cfe-d2ee-4420-87f4-945728f58f87",
		Name:            "Vengeful Pharaoh",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered:       []game.TriggeredAbility{pharaoh},
	})
}
