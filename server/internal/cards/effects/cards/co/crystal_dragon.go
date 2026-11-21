package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crystal Dragon // Rob the Hoard — an Adventure creature (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	Crystal Dragon {4}{W}{W}, Creature — Dragon, 4/4
//	  "Flying, vigilance"
//	Rob the Hoard {1}{W}, Sorcery — Adventure
//	  "Return target artifact, enchantment, or legendary card from your
//	   graveyard to your hand."
//
// Back for Seconds' return to hand over a card in the caster's own graveyard
// that is an artifact, an enchantment or legendary, judged as the spell is
// announced and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	const oracleID = "56b0cb64-b886-413e-bb20-392bf12d3926"
	effects.Register(effects.Spec{
		OracleID:        oracleID,
		Name:            "Crystal Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Rob the Hoard",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("target artifact, enchantment, or legendary card from your graveyard",
			effects.Or(effects.Artifact(), effects.Enchantment(), effects.Legendary()), effects.YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
