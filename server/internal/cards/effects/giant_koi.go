package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giant Koi — Creature — Fish {4}{U}{U}, 5/7 (Avatar: The Last
// Airbender, #4459):
//
//	"Waterbend {3}: This creature can't be blocked this turn. (While
//	 paying a waterbend cost, you can tap your artifacts and creatures
//	 to help. Each one pays for {1}.)
//	 Islandcycling {2} ({2}, Discard this card: Search your library for
//	 an Island card, reveal it, put it into your hand, then shuffle.)"
//
// WaterbendCost("{3}") is the whole cost (CR 701.67a); the effect is an
// until-end-of-turn can't-be-blocked restriction on this object only.
// Islandcycling is the shared Typecycling constructor (CR 702.29e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c93bb65a-a74a-499b-a538-65272079aab8",
		Name:         "Giant Koi",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "Waterbend {3}: This creature can't be blocked this turn.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    WaterbendCost("{3}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
						Label: "Giant Koi — can't be blocked"}.Apply(NewContext(g, item))
				},
			},
			Typecycling("Islandcycling", "{2}", "an Island card", IsLandWithSubtype("Island")),
		},
	})
}
