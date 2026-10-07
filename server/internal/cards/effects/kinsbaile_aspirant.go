package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kinsbaile Aspirant — Creature — Kithkin Citizen {W}, 2/1:
//
//	"As an additional cost to cast this spell, behold a Kithkin or pay {2}. (To behold a Kithkin, choose a Kithkin you control or reveal a Kithkin card from your hand.)
//	 Whenever another creature you control enters, this creature gets +1/+1 until end of turn."
//
// The additional cost is the either/or branch cost of ADR 0100 §2
// with the behold branch added by its 2026-10-07 amendment
// (BeholdOrPay): the caster announces the branch, names the
// card on reveal_ids, and either shows it to the table or pays {2}
// more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "845d80ee-1cd5-437a-90a5-40dcc657d3aa",
		Name:           "Kinsbaile Aspirant",
		Completeness:   CompletenessFull,
		AdditionalCost: BeholdOrPay("a", "Kithkin", "{2}"),
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Kinsbaile Aspirant — +1/+1 until end of turn", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{
					Target:    ctx.Source(),
					Power:     1,
					Toughness: 1,
					Label:     "Kinsbaile Aspirant — +1/+1 until end of turn",
				}.Apply(ctx)
			}),
		},
	})
}
