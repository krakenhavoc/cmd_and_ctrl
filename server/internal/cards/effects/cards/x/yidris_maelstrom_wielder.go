package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yidris, Maelstrom Wielder — Legendary Creature — Ogre Wizard {U}{B}{R}{G},
// 5/4 (Secret Lair Drop):
//
//	"Trample
//	 Whenever Yidris deals combat damage to a player, as you cast spells from
//	 your hand this turn, they gain cascade. (When you cast the spell, exile
//	 cards from the top of your library until you exile a nonland card that
//	 costs less. You may cast it without paying its mana cost. Put the exiled
//	 cards on the bottom in a random order.)"
//
// The combat damage trigger sets up, for the rest of the turn, a repeating
// delayed trigger (CR 603.7) on every spell its controller casts from their
// hand, which cascades off that spell's mana value as it was cast (X counted,
// CR 202.3e): the cascade ability the spell "gains" is carried out by that
// trigger, which goes on the stack above the spell as cascade's own does.
// Spells cast from anywhere but the hand get nothing, and the trigger ends
// with the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9efe8aff-9a7d-4397-b5fd-c1a0fad7c15f",
		Name:            "Yidris, Maelstrom Wielder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Yidris, Maelstrom Wielder — spells you cast from your hand this turn gain cascade",
				func(g *game.Game, item *game.StackItem) error {
					d := effects.WhenYouNextCastFromHand("Yidris, Maelstrom Wielder — cascade", game.CastFilter{}, yidrisCascadeBody)
					d.Repeats = true
					return d.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// yidrisCascadeBody cascades off the spell the delayed trigger fired on: the
// spell rides the item as its payload, and its mana value is read as it was
// cast. An on-disk key: never renamed, never reused.
var yidrisCascadeBody = game.SimpleDelayedBody("cascade/the-spell-you-just-cast-from-hand", func(g *game.Game, item *game.StackItem) error {
	cast := effects.NewContext(g, item).PayloadCards()
	if len(cast) == 0 {
		return nil
	}
	spell, ok := g.LookupCardForEffect(cast[0])
	if !ok {
		return nil
	}
	mv, _ := g.ManaValueForEffect(spell)
	return g.CascadeForEffect(item.Controller, item.SourceCardID, mv)
})
