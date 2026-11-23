package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invasion of Ergamon // Truga Cliffcharger — March of the Machine (#5477).
//
// Front, Battle — Siege 5, {R}{G}: "When this Siege enters, create a Treasure
// token. Then you may discard a card. If you do, draw a card." Back, Truga
// Cliffcharger, Creature — Rhino 3/4: "Trample. When this creature enters,
// you may discard a card. If you do, search your library for a land or battle
// card, reveal it, put it into your hand, then shuffle."
//
// The discard is chosen at resolution ("up to one"), and the draw / search
// follow only when a card really was discarded, through keyed continuations.
//
// No simplification.
const InvasionOfErgamonOracleID = "fe84933e-d98f-4cbc-b272-387a80894e31"

const trugaCliffchargerLabel = "Truga Cliffcharger — you may discard a card; if you do, search for a land or battle card"

func init() {
	effects.Register(effects.Spec{
		OracleID:     InvasionOfErgamonOracleID,
		Name:         "Invasion of Ergamon",
		Completeness: effects.CompletenessFull,
		Battle:       &effects.BattleSpec{Defense: 5, Subtype: effects.BattleSubtypeSiege},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Invasion of Ergamon — create a Treasure token, then you may discard a card and draw a card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.CreateToken{Template: effects.TreasureToken(), N: 1}).Apply(ctx); err != nil {
						return err
					}
					return effects.B39MayDiscardThenDraw(1, false, "Invasion of Ergamon — you may discard a card; if you do, draw a card",
						effects.DrawOneIfDiscarded)(ctx)
				}),
			effects.DefeatedTrigger("Invasion of Ergamon — defeated: exile it, then cast Truga Cliffcharger", effects.SiegeDefeated()),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        InvasionOfErgamonOracleID + "#1",
		Name:            "Truga Cliffcharger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(trugaCliffchargerLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player:   ctx.Controller(),
					Source:   ctx.Source(),
					N:        1,
					UpTo:     true,
					Question: trugaCliffchargerLabel,
				}, trugaCliffchargerSearch, nil)
			}),
		},
	})
}

var trugaCliffchargerSearch = game.RegisterDiscardThen("choose-cards/truga-cliffcharger-search", func(g *game.Game, r game.DiscardsDone) error {
	if r.Discarded.Count() == 0 {
		return nil
	}
	return effects.SearchLibrary{
		Player:    r.Player,
		Predicate: func(c game.Card) bool { return c.IsLand() || c.IsBattle() },
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Truga Cliffcharger — a land or battle card, to your hand",
		Source:    r.Source,
	}.Apply(effects.KeyedContext(g, r.Player, r.Source))
})
