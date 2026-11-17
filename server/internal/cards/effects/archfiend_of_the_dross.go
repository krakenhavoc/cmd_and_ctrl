package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archfiend of the Dross — Creature — Phyrexian Demon {2}{B}{B}, 6/6:
//
//	"Flying
//	 This creature enters with four oil counters on it.
//	 At the beginning of your upkeep, remove an oil counter from this
//	 creature. Then if it has no oil counters on it, you lose the game.
//	 Whenever a creature an opponent controls dies, its controller loses
//	 2 life."
//
// Entry counters are the CR 614.1c self-replacement; the upkeep trigger
// removes a counter as it resolves and reads "no oil counters" afterwards,
// so the fourth upkeep ends the game for its controller. "An opponent"
// is relative to the Archfiend's controller and the dead creature's
// controller is read as it last existed (CR 603.10a). The life loss is
// life loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9188031e-1c24-4981-a3d2-461a2bf917b6",
		Name:            "Archfiend of the Dross",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{s82OneB12EntersWithOil(4, "Archfiend of the Dross enters with four oil counters on it")},
		Triggered: []game.TriggeredAbility{
			s82OneB12OilUpkeep("Archfiend of the Dross — remove an oil counter; if none, you lose the game",
				func(g *game.Game, item *game.StackItem) error {
					return LoseTheGame{Player: item.Controller}.Apply(NewContext(g, item))
				}),
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					dead, ok := diedCreature(ev, g)
					return ok && leftUnderControlOf(ev, dead) != source.Controller
				},
				Key: "Archfiend of the Dross — its controller loses 2 life",
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					dead, ok := diedCreature(ev, g)
					if !ok {
						return nil
					}
					item := game.NewTriggeredItem(source, "Archfiend of the Dross — its controller loses 2 life")
					item.Params.Player = leftUnderControlOf(ev, dead)
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Params.Player, -2)
				},
			},
		},
	})
}
