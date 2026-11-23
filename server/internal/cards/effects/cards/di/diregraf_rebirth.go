package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diregraf Rebirth — Sorcery {3}{B}{G} (Midnight Hunt, #5497):
//
//	"This spell costs {1} less to cast for each creature that died this turn.
//	 Return target creature card from your graveyard to the battlefield.
//	 Flashback {5}{B}{G}"
//
// Zombify's reanimation with Dread Return's graveyard casting. The discount
// is Bone Devourer's table-wide count of creatures that died this turn
// (b11CreaturesDiedThisTurn, anyone's, tokens included), read at CR 601.2f
// for every way the spell is cast, so the flashback cost shrinks too. It
// spends generic mana only, so the spell never costs less than {B}{G}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "1920254f-0f65-4426-aabc-f3807476b67b",
		Name:          "Diregraf Rebirth",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetCreatureInYourGraveyard(),
		Purpose:       effects.ForTargets(effects.Reanimates(0)),
		CastableZones: []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{
			effects.Flashback("{5}{B}{G}"),
		},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int {
				if q.Game == nil {
					return 0
				}
				return effects.B11CreaturesDiedThisTurn(q.Game)
			}, "This spell costs {1} less to cast for each creature that died this turn."),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			_, _ = effects.ReanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
