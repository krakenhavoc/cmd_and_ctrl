package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Primordial Plasm — Creature — Ooze 4/4, the BACK FACE of Invasion of
// Muraganda (face 1):
//
//	"At the beginning of combat on your turn, another target creature
//	 gets +2/+2 and loses all abilities until end of turn."
//
// Registered under "<oracle_id>#1". Both halves are one until-end-of-turn
// effect on the same creature: the +2/+2 stays, the abilities go, and the
// target may be an opponent's creature or one of yours.
//
// No simplification.
func init() {
	const name = "Primordial Plasm"
	trig := effects.AtBeginningOfYourCombat(name+" — another target creature gets +2/+2 and loses all abilities until end of turn",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetCard {
				return nil
			}
			t := legal[0]
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: name + " — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			return effects.UntilEndOfTurn(ctx, t.ID, nil, name+" — loses all abilities", game.LoseAllAbilitiesMod())
		})
	trig.Targets = effects.Another(effects.TargetCreature("another target creature"))
	effects.Register(effects.Spec{
		OracleID:     effects.InvasionOfMuragandaOracleID + "#1",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trig},
	})
}
