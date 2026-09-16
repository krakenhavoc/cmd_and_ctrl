package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ravenous Squirrel — Creature — Squirrel {B/G}, 1/1 (EDHREC rank
// 3030):
//
//	"Whenever you sacrifice an artifact or creature, put a +1/+1
//	 counter on this creature.
//	 {1}{B}{G}, Sacrifice an artifact or creature: You gain 1 life
//	 and draw a card."
//
// The one-drop aristocrat. The trigger is
// b28YouSacrificedArtifactOrCreature — the sacrifice event fires
// before the zone move, so the permanent is still on the battlefield
// to be typed — and grows the Squirrel if it is still there when the
// trigger resolves. The activated ability's sacrifice is a COST
// (SacrificeOther over artifacts and creatures, the Squirrel itself a
// legal choice as in paper), paid at announce, so it fires the first
// ability above the second: the counter lands, then the life and the
// card, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8fb48d65-406b-4231-83ea-9cf3bb8dff76",
		Name:         "Ravenous Squirrel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b28YouSacrificedArtifactOrCreature(ev, source, g)
			}, "Ravenous Squirrel — put a +1/+1 counter on it", putCounterOnSelf),
		},
		Activated: []ActivatedAbility{{
			Label: "{1}{B}{G}, Sacrifice an artifact or creature: You gain 1 life and draw a card.",
			Cost: Plus(ManaCost("{1}{B}{G}"), game.AbilityCost{
				SacrificeOther: sacrificeSpec("an artifact or creature", Or(Artifact(), Creature())),
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (GainLife{Player: item.Controller, Amount: 1}).Apply(ctx); err != nil {
					return err
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			},
		}},
	})
}
