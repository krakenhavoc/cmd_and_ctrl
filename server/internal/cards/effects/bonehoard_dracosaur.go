package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bonehoard Dracosaur — Creature — Dinosaur Dragon {3}{R}{R}, 5/5
// (EDHREC rank 1375):
//
//	"Flying, first strike
//	 At the beginning of your upkeep, exile the top two cards of your
//	 library. You may play them this turn. If you exiled a land card
//	 this way, create a 3/1 red Dinosaur creature token. If you exiled
//	 a nonland card this way, create a Treasure token."
//
// Two cards a turn plus a body or a Treasure for each kind, on a 5/5
// flying first striker. The exile is the impulse-exile primitive
// with a "play" grant (so the land can be the turn's drop), and the
// riders read the exiled cards where they now sit — the printed
// type line, which for "land card" / "nonland card" is exactly the
// question. Each rider fires at most once ("a land card", not "each
// land card"): two lands is one Dinosaur, one land and one spell is
// one of each, and an empty library exiles nothing and makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7ba550a5-81fc-42c2-8df8-ad455d938605",
		Name:            "Bonehoard Dracosaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Bonehoard Dracosaur — exile the top two cards, play them this turn", func(g *game.Game, item *game.StackItem) error {
				// The riders read what was EXILED, so they wait for
				// the continuation form: a commander among the two
				// cards pauses the batch on its owner's CR 903.9
				// prompt, and the fire-and-forget form's return would
				// be short or empty right when this needs it most
				// (#1587).
				return b12ImpulseExileForTurnThen(g, item, 2, func(g *game.Game, exiled []uuid.UUID) error {
					land, nonland := false, false
					for _, id := range exiled {
						c, ok := g.LookupCardForEffect(id)
						if !ok {
							continue
						}
						if c.IsLand() {
							land = true
						} else {
							nonland = true
						}
					}
					ctx := NewContext(g, item)
					if land {
						if err := (CreateToken{Controller: item.Controller, Template: TokenCard("3/1 red Dinosaur"), N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					if nonland {
						if err := (CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			}),
		},
	})
}
