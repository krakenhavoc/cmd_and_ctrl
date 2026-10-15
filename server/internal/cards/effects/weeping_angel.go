package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weeping Angel — Artifact Creature — Alien Angel {1}{U}{B}, 2/2:
//
//	"Flash
//	 First strike, vigilance
//	 Whenever an opponent casts a creature spell, this creature isn't a
//	 creature until end of turn.
//	 If this creature would deal combat damage to a creature, prevent
//	 that damage and that creature's owner shuffles it into their
//	 library."
//
// The trigger is a layer-4 RemoveTypesMod("Creature") until end of
// turn. The layer pass drops the creature types with it (CR 205.1a,
// #2041), so while it isn't a creature it is an Artifact with no
// subtype, as the ruling says: it "continues to be an artifact but will
// stop being both an Alien and an Angel until it becomes a creature
// again". It is removed from combat as it stops being a creature (CR
// 506.4).
//
// The prevention is an ADR 0108 §8 static (#1906) applied once per
// creature it would deal combat damage to, so blocked by two creatures
// it shuffles both away. The shuffle-in is the ordinary tuck (a
// commander's owner is offered the command zone, CR 903.9), and the
// owner shuffles either way, as Chaos Warp does. Damage that can't be
// prevented is dealt and the creature is still shuffled in (CR 615.12).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1a89c623-c9c3-443c-bbd3-834767b6bdb8",
		Name:            "Weeping Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "first strike", "vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, AnOpponentCast(Creature()),
				"Weeping Angel — isn't a creature until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
						return nil
					}
					return untilEndOfTurn(ctx, item.SourceCardID, nil,
						"Weeping Angel — isn't a creature until end of turn", game.RemoveTypesMod("Creature"))
				}),
		},
		Replacements: []game.ReplacementEffect{
			PreventDamageDealtTo(PreventionStatic{
				To:     ToACreature,
				Damage: CombatDamage,
				From:   FromThis(),
				Then:   weepingAngelShuffleBody,
				Label:  "Weeping Angel — prevent the combat damage; that creature's owner shuffles it into their library",
			}),
		},
	})
}

// The additional effect: "that creature's owner shuffles it into their
// library", on the creature the damage would have been dealt to while
// it is still that object.
var weepingAngelShuffleBody = game.DelayedBody("weeping-angel/owner-shuffles-it-in", weepingAngelShuffle)

func weepingAngelShuffle(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	it, ok := followUpRecipient(g, item)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(it)
	if !ok {
		return nil
	}
	owner := c.Owner
	return g.TuckToLibraryThenForEffect(it, game.TuckOptions{}, func(g *game.Game, _ bool) error {
		return g.ShuffleLibraryForEffect(owner)
	})
}
