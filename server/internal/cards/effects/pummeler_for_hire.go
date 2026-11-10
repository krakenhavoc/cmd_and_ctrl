package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pummeler for Hire — Creature — Giant Mercenary {4}{G} (Lorwyn Eclipsed, #4458):
//
//	"Reach, vigilance
//	 Ward {2}
//	 When this creature enters, you gain X life, where X is the greatest
//	 power among Giants you control."
//
// The Pummeler is a Giant itself and is counted as the ability resolves.
func init() {
	Register(Spec{
		OracleID:        "ffa23c3a-76a8-414f-aa3b-fc09aabd78c5",
		Name:            "Pummeler for Hire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Pummeler for Hire — ward {2}"),
			WhenThisEnters("Pummeler for Hire — you gain X life, X the greatest power among Giants you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					giant := ctx.TextSubtype("Giant")
					best := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.HasSubtype(giant) && c.CurrentPower() > best {
							best = c.CurrentPower()
						}
					}
					return GainLife{Amount: best}.Apply(ctx)
				}),
		},
	})
}
