package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fateful Handoff — Sorcery {3}{B} (The Brothers' War):
//
//	"Draw cards equal to the mana value of target artifact or creature you
//	 control. An opponent gains control of that permanent."
//
// The mana value is read as the spell resolves, before the permanent
// moves (CR 608.2h); the draw happens first, then you choose the
// opponent (ChoosePlayer, Khârn the Betrayer's question) and they gain
// control of it for good. A target that is no longer legal makes the whole
// spell do nothing (CR 608.2b). Choosing is skipped when it is a
// two-player table's only opponent.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ecde43f-e418-437b-8b8f-efe120f96cb3",
		Name:         "Fateful Handoff",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or creature you control", effects.Or(effects.Artifact(), effects.Creature()), effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, mv, ok := effects.FirstLegalTargetCardManaValue(ctx)
			if !ok {
				return nil
			}
			if err := (effects.DrawCards{N: mv}).Apply(ctx); err != nil {
				return err
			}
			return effects.ChoosePlayer{
				Among:    effects.Opponents,
				Question: "Fateful Handoff — choose an opponent to gain control of that permanent",
				ThenKey:  fatefulHandoffChosen,
				Carry:    []uuid.UUID{id},
			}.Apply(ctx)
		},
	})
}

// fatefulHandoffChosen hands the permanent to the chosen opponent. An
// on-disk key. The permanent must still be there under the same
// controller as it was targeted.
var fatefulHandoffChosen = effects.ChoosePlayerThen("option-pick/fateful-handoff-opponent",
	func(ctx *effects.Context, chosen uuid.UUID, carry []uuid.UUID) error {
		if chosen == uuid.Nil || len(carry) == 0 || !effects.OnBattlefield(ctx.Game, carry[0]) {
			return nil
		}
		ctx.Game.GainControlForEffect(ctx.Source(), carry[0], chosen, game.IndefiniteDuration(),
			"Fateful Handoff — an opponent gains control of that permanent")
		return nil
	})
