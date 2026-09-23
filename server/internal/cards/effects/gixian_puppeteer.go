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
// GRAVEYARD, where the Puppeteer itself now is — so "another" matters.
// The clause is built per trigger through AnotherTarget
// (TriggeredAbility.TargetsFrom, which is handed the source), so the
// picker excludes THIS Puppeteer by instance rather than by name: a
// second Gixian Puppeteer in the same graveyard — a Clone or a token
// copy, or simply a second copy that died earlier — is a legal
// target, as printed, and the Puppeteer whose trigger it is never is.
// The resolution re-check (CR 608.2b) runs the same clause.
//
// Mana value is read off the card in the graveyard, where it has no
// layer cache and the printed cost is the only reading there is —
// which is also the correct one (CR 202.3b: a card outside the stack
// and battlefield has its printed mana value, X counting as zero).
func init() {
	Register(Spec{
		OracleID:     "9d6a9a37-a245-4503-baeb-9488553798ab",
		Name:         "Gixian Puppeteer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b40IsYourSecondDrawThisTurn(ev, source, g)
			}, "Gixian Puppeteer — each opponent loses 2, you gain 2", b40DrainEachOpponent(2)),
			gixianPuppeteerDiesTrigger(),
		},
	})
}

// gixianPuppeteerDiesTrigger is the reanimation half, split out so
// TargetsFrom can be set directly — Targeting only takes a static
// *game.TargetSpec, and this clause needs the trigger's own source
// (see the card comment).
func gixianPuppeteerDiesTrigger() game.TriggeredAbility {
	t := WhenThisDies("Gixian Puppeteer — reanimate a creature card with mana value 3 or less", func(g *game.Game, item *game.StackItem) error {
		if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		return ReturnFromGraveyard{
			Target: item.Targets[0].ID,
			Dest:   game.ZoneBattlefield,
		}.Apply(NewContext(g, item))
	})
	t.TargetsFrom = AnotherTarget(func(other CardPredicate) *game.TargetSpec {
		return TargetCardInGraveyard("another target creature card with mana value 3 or less from your graveyard",
			YouOwn(), Creature(), ManaValueLE(3), other)
	})
	return t
}
