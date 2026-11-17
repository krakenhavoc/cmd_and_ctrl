package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avatar of Growth — Creature — Elemental Avatar {4}{G}{G}, 4/4 (Game
// Night, #5523):
//
//	"This spell costs {1} less to cast for each opponent you have.
//	 Trample
//	 When this creature enters, each player searches their library for
//	 up to two basic land cards, puts them onto the battlefield, then
//	 shuffles."
//
// The discount is undaunted's count as a self cost modifier (CR 601.2f,
// OpponentsOfCaster): the caster's opponents still in the game as it is
// cast. It reduces generic mana only, so the floor is {G}{G}; at a full
// four-player table it costs {1}{G}{G}.
//
// The search is Field of Ruin's, two basics deep: one search per player
// still in the game, each addressed to its own player over their own
// library, so the prompts can all be open at once. A search for a basic
// land names a quality, so each player may find fewer than two, or none
// (CR 701.23b), and still shuffles. The lands enter untapped, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "07db0374-3297-49c3-886d-a6bb42f7bb18",
		Name:            "Avatar of Growth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(OpponentsOfCaster(), "This spell costs {1} less to cast for each opponent you have."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Avatar of Growth — each player searches their library for up to two basic land cards and puts them onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, p := range g.Seats {
						if p == nil || p.Eliminated {
							continue
						}
						if err := (SearchLibrary{
							Player:    p.ID,
							Predicate: IsBasicLand,
							Dest:      game.ZoneBattlefield,
							Limit:     2,
							Shuffle:   true,
							Reason:    "Avatar of Growth — search for up to two basic lands",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
