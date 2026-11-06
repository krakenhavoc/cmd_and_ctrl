package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0259a9ed-0558-4183-b737-871ed5573c32",
		Name:         "Gran-Gran",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Gran-Gran — draw a card, then discard a card", b35LootOne),
				game.Purpose{Draws: 1, Discards: 1}),
		},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Noncreature spells you cast cost {1} less to cast as long as there are three or more Lesson cards in your graveyard.",
				YourSpell(), NoncreatureSpell(), avB4ThreeLessonsInYourGraveyard()),
		},
	})
}
