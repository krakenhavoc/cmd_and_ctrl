package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Champions of the Perfect — Creature — Elf Warrior {3}{G}, 6/6
// (Lorwyn Eclipsed):
//
//	"As an additional cost to cast this spell, behold an Elf and exile it.
//	 (Exile an Elf you control or an Elf card from your hand.)
//	 Whenever you cast a creature spell, draw a card.
//	 When this creature leaves the battlefield, return the exiled card to
//	 its owner's hand."
//
// Champion of the Clachan's cost and leave trigger (#4687,
// BeholdAndExileCost): one Elf, a permanent you control or a card in your
// hand (a changeling counts), exiled with the spell on the stack and linked
// to the permanent the spell becomes (CR 607.2q). The cast trigger is Beast
// Whisperer's: your own creature spells, read on the stack, and the
// Champions themselves are not on the battlefield when their own cast
// fires.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "9cb7f23f-322d-4a87-b6c6-15b9d446aef3",
		Name:           "Champions of the Perfect",
		Completeness:   CompletenessFull,
		AdditionalCost: BeholdAndExileCost("an", "Elf"),
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastACreatureSpell, "Champions of the Perfect — draw a card", Do(DrawCards{N: 1})),
			ReturnTheExiledCardToHand("Champions of the Perfect"),
		},
	})
}
