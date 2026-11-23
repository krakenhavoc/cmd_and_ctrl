package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gran-Gran — Legendary Creature — Human Peasant Ally {U}, 1/2 (Avatar:
// The Last Airbender, #4459):
//
//	"Whenever Gran-Gran becomes tapped, draw a card, then discard a card.
//	 Noncreature spells you cast cost {1} less to cast as long as there
//	 are three or more Lesson cards in your graveyard."
//
// "Becomes tapped" is ThisBecameTapped over the tap and the attack
// events (Hemmer, Fixing What Is Broken's shape), so any tap loots: an
// attack, a cost, an opponent's tap effect. The tax cut is Lyse Hext's
// generic {1} off noncreature spells you cast, gated on the Lesson count
// in your graveyard as each spell is priced (CR 601.2f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0259a9ed-0558-4183-b737-871ed5573c32",
		Name:         "Gran-Gran",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Gran-Gran — draw a card, then discard a card", effects.B35LootOne),
				game.Purpose{Draws: 1, Discards: 1}),
		},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Noncreature spells you cast cost {1} less to cast as long as there are three or more Lesson cards in your graveyard.",
				effects.YourSpell(), effects.NoncreatureSpell(), effects.AvB4ThreeLessonsInYourGraveyard()),
		},
	})
}
