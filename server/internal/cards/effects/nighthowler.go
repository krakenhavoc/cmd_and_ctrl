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
// applied on top of the printed 0/0. A Nighthowler cast as a creature
// with nothing in any graveyard is a 0/0 and dies to the state-based-
// action sweep the moment it lands — the printed behaviour.
//
// Bestowed (ADR 0141, #2862) it is an Aura on a creature, the bonus
// goes to the host (SelfCreatureOrAttached), and when the host dies it
// becomes unattached and stays as a creature whose X already counts
// the host (CR 702.103f).
func init() {
	Register(Spec{
		OracleID:     "57b3f7fc-1812-4134-a645-6cef48a8aa71",
		Name:         "Nighthowler",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Bestow("{2}{B}{B}"),
		},
		Static: []game.StaticAbility{
			PumpSelfCreatureOrAttachedPer(1, 1, func(g *game.Game, _ *game.Card) int {
				return b41CreatureCardsInAllGraveyards(g)
			}),
		},
	})
}
