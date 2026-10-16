package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brimaz, Blight of Oreskos — Legendary Creature — Phyrexian Cat {2}{W}{B},
// 3/4 (March of the Machine):
//
//	"Whenever you cast a Phyrexian creature or artifact creature spell,
//	 incubate X, where X is that spell's mana value.
//	 At the beginning of each end step, if a Phyrexian died under your
//	 control this turn, proliferate."
//
// The second ability is every player's end step (it does not read the
// event's actor) with an intervening if (CR 603.4): checked when the end
// step begins and again as the trigger resolves, off
// TurnTally.DiedSubtypes (#3592), the per-controller tally of the
// creature types that died this turn. "Under your control" is the
// controller the creature had as it left (CR 603.10a), a token that died
// counts, and a changeling is a Phyrexian. The check names the trigger's
// controller, so Brimaz changing hands before the end step asks the new
// controller, as printed.
//
// "Died" is a creature going from the battlefield to a graveyard, which
// is every Phyrexian that exists as a card or token with a creature
// type; a Phyrexian that is not a creature does not occur in the card
// pool. No simplification.
func init() {
	const label = "Brimaz, Blight of Oreskos — proliferate"
	Register(Spec{
		OracleID:     "a6e7cbfc-6d5a-4a67-9dad-d512d3c97bda",
		Name:         "Brimaz, Blight of Oreskos",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(And(Creature(), Or(HasSubtype("Phyrexian"), Artifact())),
				"Brimaz, Blight of Oreskos — incubate X, where X is that spell's mana value",
				incubateThatSpellsManaValue),
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15EndStepBegan(ev) && source != nil && g.DiedWithSubtypeThisTurn(source.Controller, source.TextSubtype("Phyrexian")) > 0
			}, label, func(g *game.Game, item *game.StackItem) error {
				// "Phyrexian" is a word of the source's text (CR 612.2, #3984).
				if g.DiedWithSubtypeThisTurn(item.Controller, g.TextReaderForEffect(item).TextSubtype("Phyrexian")) == 0 {
					return nil
				}
				return Proliferate{}.Apply(NewContext(g, item))
			}),
		},
	})
}
