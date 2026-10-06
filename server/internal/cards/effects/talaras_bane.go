package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Talara's Bane — Sorcery {1}{B}:
//
//	"Target opponent reveals their hand. You choose a green or white
//	 creature card from it. You gain life equal to that creature card's
//	 toughness, then that player discards that card."
//
// The revealed-hand pick with a clause that reads the chosen card
// (#2115, ADR 0116's 2026-10-05 amendment). The whole table sees the
// hand (CR 701.20a) and only a green or white creature card may be
// chosen. The life gain is printed before the discard (CR 608.2c), so
// it is a keyed continuation that runs while the card is still in the
// hand and then discards it. "That creature card's toughness" is the
// card's in the hand, as its 2008-08-01 ruling says, with its own
// characteristic-defining ability applied, which works in every zone
// (CR 113.6a, 604.3): a Tarmogoyf in a hand has its graveyard-counted
// toughness, not a 0. Each candidate's toughness is read as the pick
// is raised, while Talara's Bane is still on the stack (CR 608.2h): the
// continuation runs when the pick is answered, by which time the spell
// is in its owner's graveyard and would add a sorcery to Tarmogoyf's
// count. Nothing else can change in between; the pick stops the table.
//
// A hand with no green or white creature card is revealed and nothing
// is gained or discarded (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2a573cc3-0a2f-4f06-bbac-c762adb9bae4",
		Name:         "Talara's Bane",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChooseFromRevealedHand{
				Player: TargetedPlayer(ctx),
				Filter: And(Creature(), Or(OfColor("G"), OfColor("W"))),
				Label:  "green or white creature card",
				Then:   talarasBaneGainToughness,
				Measure: func(g *game.Game, c game.Card) int {
					return g.ToughnessAnywhereForEffect(c)
				},
			}.Apply(ctx)
		},
	})
}

// talarasBaneGainToughness is "You gain life equal to that creature
// card's toughness, then that player discards that card." It runs
// before the discard, and the discard is pick.Done.
var talarasBaneGainToughness = RevealedPickFirst("revealed-pick/talaras-bane-gain-toughness",
	func(ctx *Context, pick game.RevealedPick) error {
		gain := 0
		for _, t := range pick.Measures {
			gain += max(t, 0)
		}
		if gain == 0 {
			return pick.Done(ctx.Game)
		}
		return ctx.Game.ChangePlayerLifeThenForEffect(pick.Source, ctx.Controller(), gain,
			func(g *game.Game, _ int) error { return pick.Done(g) })
	})
