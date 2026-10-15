package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// SOLDIER Military Program — Enchantment {2}{W} (#3656):
//
//	"At the beginning of combat on your turn, choose one. If you
//	 control a commander, you may choose both instead.
//	 • Create a 1/1 white Soldier creature token.
//	 • Put a +1/+1 counter on each of up to two Soldiers you control."
//
// A modal beginning-of-combat trigger (Tiller Engine's Modes slot) with
// Akroma's Will's conditional count: a commander permanent you control
// lets the controller take both bullets. Neither bullet targets, so the
// second one asks for its Soldiers as it resolves: a choose-cards
// prompt with a floor of zero, over the Soldiers still on the
// battlefield. With both modes the token comes first and so may be
// chosen, as the bullets read in order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e2034ee-adc5-4a24-9605-c9722bf813c1",
		Name:         "SOLDIER Military Program",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventStepBegan},
			Key:       "SOLDIER Military Program — choose one; with a commander, you may choose both",
			AppliesTo: StepBegan(game.StepBeginCombat, true),
			Effect:    func(*game.Game, *game.StackItem) error { return nil },
			Modes: ChooseOne(
				ModeDoing("Create a 1/1 white Soldier creature token.", nil,
					func(_ *game.StackItem, ctx *Context, _ int) error {
						return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 white Soldier"), N: 1}.Apply(ctx)
					}),
				ModeDoing("Put a +1/+1 counter on each of up to two Soldiers you control.", nil,
					soldierMilitaryProgramCounters),
			).OrUpToIf(2, YouControlACommander),
		}},
	})
}

// soldierMilitaryProgramCounters is the second bullet: pick up to two
// Soldiers the controller controls and put a +1/+1 counter on each.
func soldierMilitaryProgramCounters(item *game.StackItem, ctx *Context, _ int) error {
	g, controller := ctx.Game, ctx.Controller()
	var soldiers []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.HasSubtype("Soldier") {
			soldiers = append(soldiers, c.InstanceID)
		}
	}
	if len(soldiers) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   item.SourceCardID,
		Question: "SOLDIER Military Program — put a +1/+1 counter on each of up to two Soldiers you control",
		Cards:    soldiers,
		Min:      0,
		Max:      2,
		Zone:     game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			for _, id := range picked {
				if c, ok := g.LookupCardForEffect(id); !ok || !g.Battlefield.Contains(id) || c.Controller != controller || !c.HasSubtype("Soldier") {
					continue
				}
				if err := g.AddCounterForEffect(id, "+1/+1", 1); err != nil {
					return err
				}
			}
			return nil
		},
	})
	return nil
}
