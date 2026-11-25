package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Song of Freyalise — Enchantment — Saga {1}{G} (Dominaria):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Until your next turn, creatures you control gain "{T}: Add one
//	 mana of any color."
//	 III — Put a +1/+1 counter on each creature you control. Those creatures
//	 gain vigilance, trample, and indestructible until end of turn."
//
// I and II lock the set of creatures you control as the chapter resolves
// (CR 611.2c) and grant each a mana ability (ADR 0093) until your next turn
// begins; a creature that arrives later gets nothing. III puts a counter on
// each creature you control and gives those same creatures the three keywords
// until end of turn.
//
// No simplification.
func init() {
	const name = "Song of Freyalise"
	grant := func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		return effects.GrantAbilitiesFor{
			Match:    effects.And(effects.Creature(), effects.YouControl()),
			Keys:     []string{songOfFreyaliseGrant},
			Duration: effects.DurationUntilYourNextTurn(ctx, item.Controller),
			Label:    name + " — creatures you control gain a mana ability",
		}.Apply(ctx)
	}
	effects.Register(effects.Spec{
		OracleID:     "9fe1f3ef-2c30-4ccc-9a67-5349e98e2502",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Grants:       []effects.AbilityGrant{effects.AnyColorManaGrant(songOfFreyaliseGrant)},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "until your next turn, creatures you control gain \"{T}: Add one mana of any color.\""), grant),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "until your next turn, creatures you control gain \"{T}: Add one mana of any color.\""), grant),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "put a +1/+1 counter on each creature you control; those creatures gain vigilance, trample, and indestructible until end of turn"),
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.B11PutCounterOnEachCreatureYouControl(g, item); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.YouControl()),
						Keywords: []string{"vigilance", "trample", "indestructible"},
						Label:    name + " — vigilance, trample, and indestructible",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// songOfFreyaliseGrant names the granted ability's bundle: an on-disk identity.
const songOfFreyaliseGrant = "song-of-freyalise/tap-any-color"
