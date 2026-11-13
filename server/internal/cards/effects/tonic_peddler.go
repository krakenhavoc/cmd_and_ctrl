package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tonic Peddler — Creature — Human Spellshaper {1}{W}, 1/1
// (Mercadian Masques, #5619):
//
//	"{W}, {T}, Discard a card: Target player gains 3 life."
//
// The spellshaper cost (Kris Mage's: mana, {T} and DiscardACard, the
// card discarded as the ability is activated, CR 602.2b) on Kenrith's
// "target player gains life" at 3. Any player may be the target, an
// opponent included; one who has left the game by resolution gains
// nothing (CR 608.2b). The {T} is a creature's, so the Peddler waits out
// summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9039f59a-915b-4b16-a5c4-1a5fbfb0b639",
		Name:         "Tonic Peddler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}, Discard a card: Target player gains 3 life.",
			Purpose: ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 3}),
			Cost:    Plus(ManaCost("{W}"), TapCost(), DiscardACard()),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				p, ok := firstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				return GainLife{Player: p, Amount: 3}.Apply(ctx)
			},
		}},
	})
}
