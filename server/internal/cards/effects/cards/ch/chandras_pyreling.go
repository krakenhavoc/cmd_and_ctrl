package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chandra's Pyreling — Creature — Elemental Lizard {1}{R}, 1/3 (Jumpstart
// 2022, #5948):
//
//	"Whenever a source you control deals noncombat damage to an opponent,
//	 this creature gets +1/+0 and gains double strike until end of turn."
//
// Niv-Mizzet, Visionary's watch: one trigger per noncombat damage event to an
// opponent from a source its controller controls (a permanent, or the spell
// still on the stack), so two sources are two triggers and the boost stacks
// (+2/+0 in all). The creature gets the boost and the keyword only while it is
// still the permanent that was on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5291deb9-0b89-42c2-bde6-26f88931c3f4",
		Name:         "Chandra's Pyreling",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ASourceYouControlDealtNoncombatDamageToAnOpponent,
				"Chandra's Pyreling — it gets +1/+0 and gains double strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Chandra's Pyreling — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"double strike"}, Label: "Chandra's Pyreling — double strike"}.Apply(ctx)
				}),
		},
	})
}
