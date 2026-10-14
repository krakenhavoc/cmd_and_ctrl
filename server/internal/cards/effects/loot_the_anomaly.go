package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loot, the Anomaly — Legendary Creature — Beast Horror {2}{B}, -2/4:
//
//	"If Loot's power is negative, he assigns combat damage as though his
//	 power were positive.
//	 Threshold — Sacrifice another creature or planeswalker: Loot gets
//	 -2/-0 until end of turn. Activate only if there are seven or more
//	 cards in your graveyard."
//
// "As though his power were positive" is Characteristic.DamageAsPositivePower:
// a layer 6 static on Loot sets the flag on Loot alone, and the combat
// damage pass (combatDamagePower, once per combatant per step) answers
// with the absolute value of his unclamped power instead of CR 510.1a's
// zero. Every other creature keeps the clamp, and a Loot that has lost
// his abilities (CR 613.1f) loses the flag with them, because the static
// is his. Trample, deathtouch, lifelink and commander damage all read the
// same number, so a -3/4 Loot with lifelink gains 3.
//
// The -2/-0 is an ordinary until-end-of-turn boost on Loot, resolved from
// the stack; the sacrifice is a cost, so it is paid at announce, and the
// threshold (an ability word, CR 207.2c) is an activation condition here.
// "Another" excludes Loot himself, by object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0142e7d-7b0b-4290-a56c-5fa950fdccf8",
		Name:         "Loot, the Anomaly",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "If Loot's power is negative, he assigns combat damage as though his power were positive.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.DamageAsPositivePower = true
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "Threshold — Sacrifice another creature or planeswalker: Loot gets -2/-0 until end of turn. Activate only if there are seven or more cards in your graveyard.",
			// A sacrifice at will, in the cost (AnswerSacOutlet); the
			// -2/-0 itself answers nothing an opponent casts.
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: game.AbilityCost{
				SacrificeOther: Another(sacrificeSpec("another creature or planeswalker",
					Or(Creature(), Planeswalker()))),
			},
			Condition: GraveyardAtLeast(7, nil),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{
					Target:    item.SourceCardID,
					Power:     -2,
					Toughness: 0,
					Label:     "Loot, the Anomaly — -2/-0",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
