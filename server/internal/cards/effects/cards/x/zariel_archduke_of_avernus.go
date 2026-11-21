package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zariel, Archduke of Avernus — Legendary Planeswalker — Zariel {2}{R}{R},
// loyalty 4 (Adventures in the Forgotten Realms, S82 batch 15, #5499):
//
//	"+1: Creatures you control get +1/+0 and gain haste until end of turn.
//	 0: Create a 1/1 red Devil creature token with 'When this token dies, it
//	    deals 1 damage to any target.'
//	 −6: You get an emblem with 'At the end of the first combat phase on
//	     your turn, untap target creature you control. After this phase,
//	     there is an additional combat phase.'"
//
// The +1 locks the set of creatures as it resolves (CR 611.2c). The Devil is
// the printed token with its dies trigger (the Spiked Corridor template).
// The emblem's trigger fires at the beginning of the end of combat step of
// the first combat phase of your turn (CR 511.1, via IsFirstCombatPhase, so
// the added combat does not trigger it again), targeting a creature you
// control as it goes on the stack: with none, the trigger is removed and no
// extra combat is added, and a target that is illegal on resolution makes the
// whole ability do nothing (CR 608.2b). The added combat comes with no main
// phase. The emblem is the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "04ddb495-e9b0-47bf-b892-c7ae7b87c10c",
		Name:            "Zariel, Archduke of Avernus",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &effects.EmblemSpec{
			Label: "Zariel, Archduke of Avernus emblem",
			Text:  "At the end of the first combat phase on your turn, untap target creature you control. After this phase, there is an additional combat phase.",
			Triggered: []game.TriggeredAbility{
				effects.Targeting(
					effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepEndCombat, true), zarielFirstCombat),
						"Zariel, Archduke of Avernus emblem — untap target creature you control; additional combat phase",
						zarielEmblemUntapAndCombat),
					effects.TargetCreature("target creature you control", effects.YouControl())),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Creatures you control get +1/+0 and gain haste until end of turn.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					yours := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{Match: yours, Power: 1, Label: "Zariel, Archduke of Avernus — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"haste"}, Label: "Zariel, Archduke of Avernus — haste"}.Apply(ctx)
				},
			},
			{
				Label:  "0: Create a 1/1 red Devil creature token with \"When this token dies, it deals 1 damage to any target.\"",
				Cost:   effects.LoyaltyCost(0),
				Effect: effects.Do(effects.CreateToken{Template: effects.DevilToken(), N: 1}),
			},
			{
				Label:  "−6: You get an emblem with \"At the end of the first combat phase on your turn, untap target creature you control. After this phase, there is an additional combat phase.\"",
				Cost:   effects.LoyaltyCost(-6),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

// zarielFirstCombat is the first combat phase of the turn.
func zarielFirstCombat(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	return effects.IsFirstCombatPhase(g)
}

// zarielEmblemUntapAndCombat untaps the chosen creature and adds the combat.
func zarielEmblemUntapAndCombat(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	t, ok := effects.FirstLegalTarget(ctx)
	if !ok {
		return nil
	}
	if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
		return err
	}
	return effects.ExtraCombatAfterThisPhase().Apply(ctx)
}
