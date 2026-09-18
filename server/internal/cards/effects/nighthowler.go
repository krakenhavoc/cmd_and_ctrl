package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nighthowler — Enchantment Creature — Horror {1}{B}{B}, 0/0 (EDHREC
// rank 4337):
//
//	"Bestow {2}{B}{B} (If you cast this card for its bestow cost, it's
//	 an Aura spell with enchant creature. It becomes a creature again
//	 if it's not attached.)
//	 This creature and enchanted creature each get +X/+X, where X is
//	 the number of creature cards in all graveyards."
//
// A three-mana creature that is a 6/6 by the mid-game in a format
// where every graveyard is filling up, and grows every time anything
// dies anywhere. It counts ALL graveyards, not just yours — that is
// the whole reason it is a Commander card rather than a Limited one.
//
// The bonus is a live read on every layer recompute, at layer 7c
// (modify), so it stacks additively with anthems and counters and is
// applied on top of the printed 0/0. A Nighthowler with nothing in any
// graveyard is a 0/0 and dies to the state-based-action sweep the
// moment it lands — the printed behaviour, and the reason the card is
// unplayable on turn three of an empty board.
//
// Declared simplification, weaker than printed (#259): BESTOW is not
// implemented. The Nighthowler can only be cast as an ordinary
// creature spell for {1}{B}{B}; it can never be cast for {2}{B}{B} as
// an Aura, so the "and enchanted creature" half of the bonus never
// applies to anything and the card cannot survive its host's death by
// becoming a creature again. What ships is the creature half, exactly
// as printed.
func init() {
	Register(Spec{
		OracleID:     "57b3f7fc-1812-4134-a645-6cef48a8aa71",
		Name:         "Nighthowler",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Bestow is not available — the Nighthowler can only be cast as a creature for {1}{B}{B}, never as an Aura for {2}{B}{B}, so it never pumps another creature.",
		},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := b41CreatureCardsInAllGraveyards(g)
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
