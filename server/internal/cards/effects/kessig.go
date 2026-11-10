package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kessig — Plane — Innistrad (#4870):
//
//	"Prevent all combat damage that would be dealt by non-Werewolf creatures.
//	 Whenever chaos ensues, each creature you control gets +2/+2, gains trample,
//	 and becomes a Werewolf in addition to its other types until end of turn."
//
// The prevention reads the damaging creature as it last existed (a creature
// with the Werewolf subtype, even beside others, is not "non-Werewolf", the
// ruling). The chaos ability affects the creatures the planar controller has
// as it resolves and no later arrival (the ruling, CR 611.2c).
//
// No simplification.
func init() {
	mine := And(Creature(), YouControl())
	Register(Spec{
		OracleID:     "b54d9a4a-1d55-4ceb-96fd-2b1d5e46de14",
		Name:         "Kessig",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			PreventAllDamageDealtTo(PreventionStatic{
				To: ToAnything, Damage: CombatDamage,
				From: func(_ *game.ReplacementEvent, _ *game.Game, lki *game.Characteristic, _ *game.Card) bool {
					return lki != nil && slices.Contains(lki.Types, "Creature") && !slices.Contains(lki.Subtypes, "Werewolf")
				},
				Label: "Kessig — prevent all combat damage that would be dealt by non-Werewolf creatures",
			}),
		},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Kessig — each creature you control gets +2/+2, gains trample, and becomes a Werewolf until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (BoostUntilEOT{Match: mine, Power: 2, Toughness: 2, Label: "Kessig — +2/+2"}).Apply(ctx); err != nil {
						return err
					}
					if err := (GrantKeywordUntilEOT{Match: mine, Keywords: []string{"trample"}, Label: "Kessig — trample"}).Apply(ctx); err != nil {
						return err
					}
					return untilEndOfTurn(ctx, uuid.Nil, mine, "Kessig — a Werewolf in addition to its other types", game.AddSubtypesMod("Werewolf"))
				}),
		},
	})
}
