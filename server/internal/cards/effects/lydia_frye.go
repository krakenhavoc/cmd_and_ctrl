package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lydia Frye — Legendary Creature — Human Assassin {2}{U/B}, 3/2 (Assassin's
// Creed, #5948):
//
//	"Lydia Frye can't be blocked by creatures with power 3 or greater.
//	 At the beginning of your end step, surveil X, where X is the number of
//	 tapped Assassins you control."
//
// The evasion is April O'Neil's block rule (the blocker's power read live as
// blockers are declared). X is counted as the trigger resolves: tapped
// Assassins you control, Lydia herself included when she is tapped, each
// once. With none, X is 0 and nothing is surveilled (CR 701.25: surveil 0
// does nothing).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0315f5e7-2867-428e-8e43-66dda0db57a0",
		Name:         "Lydia Frye",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerGE(3), "creatures with power 3 or greater"),
		},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Lydia Frye — surveil X, where X is the number of tapped Assassins you control",
				func(g *game.Game, item *game.StackItem) error {
					assassin := HasSubtype("Assassin")
					x := 0
					for _, c := range g.Battlefield.Cards {
						if c.Controller == item.Controller && c.Tapped && assassin(g, item.Controller, c) {
							x++
						}
					}
					if x == 0 {
						return nil
					}
					return Surveil{Player: item.Controller, N: x}.Apply(NewContext(g, item))
				}),
		},
	})
}
