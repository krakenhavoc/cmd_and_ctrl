package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doomgape — Creature — Elemental {4}{B/G}{B/G}{B/G}, 10/10:
//
//	"Trample
//	 At the beginning of your upkeep, sacrifice a creature. You gain
//	 life equal to that creature's toughness."
//
// Consuming Vapors' sacrifice-and-read body run by the controller on
// themselves. The sacrifice is mandatory and may be Doomgape itself;
// the life is the sacrificed creature's last-known toughness.
func init() {
	Register(Spec{
		OracleID:        "ab94d8f7-f811-436a-9a22-9a12bd929e98",
		Name:            "Doomgape",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Doomgape — sacrifice a creature, gain life equal to its toughness",
				func(g *game.Game, item *game.StackItem) error {
					me := item.Controller
					return g.PlayerSacrificesThenForEffect(item.SourceCardID, me,
						sacrificeSpec("a creature", Creature()),
						"Doomgape — sacrifice a creature", 1,
						func(g *game.Game, sacrificed game.PromptedSacrifices) error {
							ids := sacrificed.By(me)
							if len(ids) == 0 {
								return nil
							}
							t := departedCreatureToughness(g, ids[0])
							if t <= 0 {
								return nil
							}
							return GainLife{Player: me, Amount: t}.Apply(NewContext(g, item))
						})
				}),
		},
	})
}
