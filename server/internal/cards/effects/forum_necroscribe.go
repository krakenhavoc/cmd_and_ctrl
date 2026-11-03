package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forum Necroscribe — Creature — Troll Warlock {5}{B}, 5/4 (Secrets of
// Strixhaven, #4456):
//
//	"Ward—Discard a card.
//	 Repartee — Whenever you cast an instant or sorcery spell that targets a
//	 creature, return target creature card from your graveyard to the
//	 battlefield."
//
// Ward—Discard a card is Ward(WardDiscard(1)) (ward.go, Graveyard
// Trespasser's shape).
//
// Repartee is an ability word (CR 207.2c); the ability is one trigger on your
// instant and sorcery casts that target a creature, read from the targets the
// spell was cast with (informedInkwrightCast, the heroic reading): a target
// that is a creature on the battlefield as the spell is cast. It resolves
// before the spell. Its own target is a creature card in YOUR graveyard,
// chosen as it goes on the stack, so with none there is no trigger. The
// return is Karmic Guide's: onto the battlefield under your control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50c97054-77c9-4446-9a41-e4f8877b887d",
		Name:         "Forum Necroscribe",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardDiscard(1), "Forum Necroscribe — ward, discard a card"),
			Targeting(On(game.EventCast, informedInkwrightCast,
				"Forum Necroscribe — Repartee: return target creature card from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					reanimateSingleTarget(NewContext(g, item), item.Controller)
					return nil
				}),
				targetCreatureInYourGraveyard()),
		},
	})
}
