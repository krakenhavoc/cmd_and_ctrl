package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Outlaw Stitcher — Creature — Human Warlock {3}{U}, 1/4 (Outlaws of
// Thunder Junction, #4468):
//
//	"When this creature enters, create a 2/2 blue and black Zombie
//	 Rogue creature token, then put two +1/+1 counters on that token for
//	 each spell you've cast this turn other than the first.
//	 Plot {4}{U} (You may pay {4}{U} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// The token is made through CreateTokensKeyedForEffect (Artistic
// Process), so "that token" is the one that landed even when the
// creation waits on a replacement-ordering prompt. The count is your
// per-turn cast tally (CastTally.Total) read as the ability resolves:
// the Stitcher itself counts when it was cast, so casting it as your
// second spell gives the token two counters, and putting it onto the
// battlefield without casting it counts only the spells you did cast.
// Plot is Beastbond Outcaster's special action (CR 702.170a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "9fda1239-97ca-41f0-8bfa-040484769396",
		Name:           "Outlaw Stitcher",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{4}{U}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Outlaw Stitcher — create a 2/2 Zombie Rogue, then put two +1/+1 counters on it for each spell you've cast this turn other than the first",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return g.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller: item.Controller,
						Source:     item.SourceCardID,
						Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("2/2 blue and black Zombie Rogue")), Count: 1}},
					}, ctx.ItemTail(outlawStitcherCounters))
				}),
		},
	})
}

// outlawStitcherCounters is "then put two +1/+1 counters on that token
// for each spell you've cast this turn other than the first".
var outlawStitcherCounters = CardTail("outlaw-stitcher/token-counters", func(ctx *Context, o game.TailOutcome) error {
	others := ctx.Game.CastTallyFor(ctx.Controller()).Total - 1
	if others <= 0 {
		return nil
	}
	for _, id := range o.IDs {
		if !onBattlefield(ctx.Game, id) {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2 * others}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
