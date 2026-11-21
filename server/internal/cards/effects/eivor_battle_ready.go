package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eivor, Battle-Ready — Legendary Creature — Human Assassin Warrior
// {3}{R}{W}, 5/5 (Assassin's Creed, #5948):
//
//	"Vigilance
//	 Haste
//	 Whenever Eivor attacks, it deals damage equal to the number of Equipment
//	 you control to each opponent."
//
// The number is counted as the trigger resolves (Equipment you control, of
// any kind, attached or not) and Eivor deals the damage to each opponent as the
// object it was (CR 608.2h), so one that has left still deals it. With no
// Equipment nothing is dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1b4a07bd-5345-4520-bd05-b5173e08135c",
		Name:            "Eivor, Battle-Ready",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Eivor, Battle-Ready — damage equal to the number of Equipment you control to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					equipment := HasSubtype("Equipment")
					n := 0
					for _, c := range g.Battlefield.Cards {
						if c.Controller == item.Controller && equipment(g, item.Controller, c) {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return damageToEachOpponent(g, item, n)
				}),
		},
	})
}
