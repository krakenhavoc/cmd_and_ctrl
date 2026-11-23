package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ambitious Dragonborn — {3}{G} Creature — Dragon Barbarian 0/0
//
//	"This creature enters with X +1/+1 counters on it, where X is the
//	 greatest power among creatures you control and creature cards in your
//	 graveyard."
//
// X is read as it enters (CR 614.1c, 122.6): the board and graveyard before
// the Dragonborn is on the battlefield, so it never counts itself, and the
// counters are there as it enters, so the 0/0 never dies for want of them
// when X is positive.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "934094d6-c897-4ae2-89cd-08d106600d3c",
		Name:         "Ambitious Dragonborn",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			RowLabel:  "Enters with X +1/+1 counters on it, where X is the greatest power among creatures you control and creature cards in your graveyard.",
			EntryOnly: true,
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield && src != nil && ev.CardID == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				if x := effects.GapB215GreatestPowerAmongCreaturesAndGraveyardCards(g, src.Controller); x > 0 {
					ev.AddCounterAtETB(game.CounterPlusOne, x)
				}
				return nil
			},
		}},
	})
}
