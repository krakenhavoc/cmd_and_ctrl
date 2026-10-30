package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stature, Young Avenger — Legendary Creature — Human Hero {1}{R/G},
// 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Reach
//	 Whenever you cast a noncreature spell, Stature has base power and
//	 toughness 4/4 until end of turn."
//
// Base power and toughness are set in layer 7b, so counters and pumps
// still apply on top of the 4/4. A second cast the same turn changes
// nothing. Stature that left the battlefield (or came back as a new
// object) is not affected.
//
// No simplification.
func init() {
	const label = "Stature, Young Avenger — base power and toughness 4/4 until end of turn"
	Register(Spec{
		OracleID:        "b1eee67f-30cf-4afb-8080-84ac282effc4",
		Name:            "Stature, Young Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), label, func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return untilEndOfTurn(NewContext(g, item), item.SourceCardID, nil, label,
					game.SetBasePowerMod(4), game.SetBaseToughnessMod(4))
			}),
		},
	})
}
