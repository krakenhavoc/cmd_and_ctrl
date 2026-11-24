package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Belbe, Corrupted Observer — Legendary Creature — Phyrexian Zombie Elf {B}{G},
// 2/2 (Commander Legends):
//
//	"At the beginning of each postcombat main phase, the active player adds
//	 {C}{C} for each of your opponents who lost life this turn. (Damage causes
//	 loss of life.)"
//
// One trigger on every player's postcombat main phase (an added second combat's
// extra main phase counts: "each postcombat main phase"). The count is read as
// the trigger resolves: Belbe's controller's opponents still in the game whose
// life lost this turn is above zero (the turn tally counts damage too, CR
// 120.3a). The mana goes to the active player, not to Belbe's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1c303d8-6d77-4bc6-82d1-2a409a6f2cfd",
		Name:         "Belbe, Corrupted Observer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepPostcombatMain,
				"Belbe, Corrupted Observer — the active player adds {C}{C} for each of your opponents who lost life this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, p := range g.Seats {
						if p.ID != item.Controller && !p.Eliminated && g.TurnTallyFor(p.ID).LifeLost > 0 {
							n++
						}
					}
					if n == 0 || g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
						return nil
					}
					return effects.AddMana{Player: g.Seats[g.Turn.ActiveSeat].ID, Amount: effects.ManaN(2*n, "C")}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
