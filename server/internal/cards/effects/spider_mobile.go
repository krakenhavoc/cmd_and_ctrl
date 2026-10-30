package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Mobile — {3} Artifact — Vehicle 3/3 (Marvel Super Heroes
// Commander, #4455):
//
//	"Trample
//	 Whenever this Vehicle attacks or blocks, it gets +1/+1 until end of
//	 turn for each Spider you control.
//	 Crew 2"
//
// One trigger on the attack declaration and the block declaration (Getaway
// Car's shape; a block is once however many attackers it blocks, CR
// 509.3a). The Spiders are counted as the trigger resolves, so a Spider
// that arrived or left in response changes the bonus, and a Vehicle that has
// left the battlefield gets nothing. Crew is Dependable Quinjet's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3c322d99-2bc4-4618-aeb3-5583fce54f1b",
		Name:            "Spider-Mobile",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source) || selfBlocksOnce(ev, source)
				},
				"Spider-Mobile — it gets +1/+1 until end of turn for each Spider you control",
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := NewContext(g, item)
					spider := ctx.TextSubtype("Spider")
					n := countControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtype(spider) })
					return BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n,
						Label: "Spider-Mobile — +1/+1 for each Spider"}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Spider-Mobile"),
		}},
	})
}
