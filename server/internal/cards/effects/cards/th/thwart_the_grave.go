package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thwart the Grave — Sorcery {4}{B}{B} (Commander Legends: Battle for Baldur's
// Gate, S82 GAP #5948):
//
//	"This spell costs {1} less to cast for each creature in your party. (Your
//	 party consists of up to one each of Cleric, Rogue, Warrior, and Wizard.)
//	 Return target creature card and up to one target Cleric, Rogue, Warrior,
//	 or Wizard creature card from your graveyard to the battlefield."
//
// The party is counted as the spell is cast (CR 700.8, the matching Kabira
// Outrider uses, so a changeling or a creature with two party types counts
// once, for the best assignment). Two clauses: a creature card and, distinct
// from it, up to one creature card of a party type, both from the caster's
// graveyard. Each still legal as the spell resolves comes back together, so
// each sees the other enter (CR 603.6a); a target that left, or whose card
// type no longer fits, is skipped (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b545e26f-2034-409c-80cc-51b72e75e6e4",
		Name:         "Thwart the Grave",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.OldB094PartySizeForCost, "This spell costs {1} less to cast for each creature in your party."),
		},
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
			effects.Distinct(effects.TargetCardInGraveyard("up to one target Cleric, Rogue, Warrior, or Wizard creature card from your graveyard",
				effects.YouOwn(), effects.Creature(),
				effects.Or(effects.Subtype("Cleric"), effects.Subtype("Rogue"), effects.Subtype("Warrior"), effects.Subtype("Wizard"))).WithCount(0, 1)),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToBattlefieldTogether(ctx)
		},
	})
}
