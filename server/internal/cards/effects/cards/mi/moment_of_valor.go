package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moment of Valor — Instant {2}{W} (Wilds of Eldraine, #4471):
//
//	"Choose one —
//	 • Untap target creature. It gets +1/+0 and gains indestructible until
//	   end of turn.
//	 • Destroy target creature with power 4 or greater."
//
// The first bullet is Act of Heroism's shape: the untap, then ONE
// until-end-of-turn record carrying the +1/+0 (layer 7c) and the
// indestructible grant (layer 6), since the second sentence is one effect.
// The second is Spectacular Tactics' destroy bullet: a creature whose power
// has dropped below 4 in response is an illegal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfac8072-2bbc-43ca-b141-9cd0799840f7",
		Name:         "Moment of Valor",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Untap target creature. It gets +1/+0 and gains indestructible until end of turn.",
				effects.TargetCreature("target creature"),
				effects.RfCreatureBModeTargetDoes(momentOfValorUntapAndProtect)),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "indestructible"))),
			effects.ModeDoing("Destroy target creature with power 4 or greater.",
				effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
				effects.DestroyTheModesTarget),
		),
	})
}

// momentOfValorUntapAndProtect is the first bullet on its still-legal
// target.
func momentOfValorUntapAndProtect(ctx *effects.Context, id uuid.UUID) error {
	if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return effects.UntilEndOfTurn(ctx, id, nil, "Moment of Valor — +1/+0 and indestructible",
		game.ModifyPTMod(1, 0), game.AddKeywordsMod("indestructible"))
}
