package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eldrazi Mimic — Creature — Eldrazi {2}, 2/1 (Oath of the Gatewatch, #5545):
//
//	"Whenever another colorless creature you control enters, you may have this
//	 creature's base power and toughness become equal to that creature's power
//	 and toughness until end of turn."
//
// The trigger is Glaring Fleshraker's (another colorless creature entering
// under your control) with the "may" asked as it resolves. The numbers are the
// entering creature's power and toughness as the ability resolves, or as it last
// existed if it has left (CR 608.2h), and set the Mimic's BASE values (layer 7b)
// until end of turn, so counters and bonuses on the Mimic still apply.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cddb3ec4-01a2-4f4d-a690-3486c73c3df8",
		Name:         "Eldrazi Mimic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && c.IsColorless()
			}, "Eldrazi Mimic — you may have its base power and toughness become equal to that creature's until end of turn", eldraziMimicCopyStats),
				"Eldrazi Mimic — have its base power and toughness become equal to the entering creature's until end of turn?"),
		},
	})
}

func eldraziMimicCopyStats(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.OnBattlefield(g, ctx.Source()) {
		return nil
	}
	entered := ctx.Trigger().Event.CardID
	return effects.ScopedEffectFor{
		Target: ctx.Source(),
		Mods: []game.Mod{
			game.SetBasePowerMod(effects.DepartedCreaturePower(g, entered)),
			game.SetBaseToughnessMod(effects.DepartedCreatureToughness(g, entered)),
		},
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Eldrazi Mimic — base power and toughness become equal to the entering creature's",
	}.Apply(ctx)
}
