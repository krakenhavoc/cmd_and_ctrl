package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_r42_helpers.go — helpers S82 GAP roll-up 42 shares between batches
// (#5948) so TestNoNewExactClonesInTheCatalog holds.

// R42RavenousDraw is the draw half of ravenous (Warhammer 40,000 Commander):
// "If X is 5 or more, draw a card when this creature enters." X is the value
// announced for the spell, read off the permanent as the trigger is built.
// Ravener (batch 193), Exocrine (batch 194) and Zoanthrope (batch 195) share it.
func R42RavenousDraw(name string) game.TriggeredAbility {
	label := name + " — ravenous: draw a card"
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventETB},
		AppliesTo: Self,
		Key:       label,
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Amount = source.CastX()
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			if item.Params.Amount < 5 {
				return nil
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
		},
	}
}

// R42RegisterTappedDual registers a land that "enters tapped" and taps for
// either of two colours (the Guildgates' shape): one Spec for the lands of
// batches 182 and 183 that print exactly those two lines.
func R42RegisterTappedDual(oracleID, name, a, b string) {
	Register(Spec{
		OracleID:      oracleID,
		Name:          name,
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{DualManaAbility(a, b)},
	})
}
