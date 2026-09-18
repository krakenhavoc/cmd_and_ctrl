package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gixian Puppeteer — Creature — Phyrexian Warlock {3}{B}, 4/3
// (EDHREC rank 4166):
//
//	"Whenever you draw your second card each turn, each opponent
//	 loses 2 life and you gain 2 life.
//	 When this creature dies, return another target creature card with
//	 mana value 3 or less from your graveyard to the battlefield."
//
// A four-mana 4/3 that is two cards: a drain engine for any deck that
// draws twice a turn, and a Gravedigger-that-reanimates when it dies.
// In a Phyrexian Arena or Faerie Mastermind deck the drain fires every
// single turn, which at four players is a six-point swing a turn for
// one card.
//
// "Your SECOND card each turn" is the only real rules work, and two
// readings matter:
//
//   - "each turn", not "each of your turns" — a draw on an opponent's
//     turn counts, so a Windfall or a cycled card on their turn can be
//     the second.
//   - "the second", exactly: the third and later draws of a turn do
//     not fire it again. The tally is the event log rather than the
//     turn counter, because a trigger's condition runs while the
//     event is being emitted and must not depend on whether the
//     counter has been bumped yet (b40IsYourSecondDrawThisTurn).
//
// The drain is life LOSS, not damage: nothing prevents it, and no
// damage doubler touches it.
//
// The dies half is a targeted reanimation and it targets FROM THE
// GRAVEYARD, where the Puppeteer itself now is — so "another" matters
// and is the catalog's by-name exclusion (b03NotNamed), the same
// posture Benevolent Hydra and Dour Port-Mage take for "another" on a
// declared target clause. The consequence is stated rather than
// hidden: a SECOND Gixian Puppeteer in the same graveyard cannot be
// the target, which the printed card allows. Weaker, never stronger.
//
// Mana value is read off the card in the graveyard, where it has no
// layer cache and the printed cost is the only reading there is —
// which is also the correct one (CR 202.3b: a card outside the stack
// and battlefield has its printed mana value, X counting as zero).
func init() {
	Register(Spec{
		OracleID:     "9d6a9a37-a245-4503-baeb-9488553798ab",
		Name:         "Gixian Puppeteer",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The dies trigger can't reanimate a second Gixian Puppeteer out of your graveyard, though the printed card can."},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b40IsYourSecondDrawThisTurn(ev, source, g)
			}, "Gixian Puppeteer — each opponent loses 2, you gain 2", b40DrainEachOpponent(2)),
			Targeting(
				WhenThisDies("Gixian Puppeteer — reanimate a creature card with mana value 3 or less", func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					return ReturnFromGraveyard{
						Target: item.Targets[0].ID,
						Dest:   game.ZoneBattlefield,
					}.Apply(NewContext(g, item))
				}),
				TargetCardInGraveyard("another target creature card with mana value 3 or less from your graveyard",
					YouOwn(), Creature(), ManaValueLE(3), b03NotNamed("Gixian Puppeteer")),
			),
		},
	})
}
