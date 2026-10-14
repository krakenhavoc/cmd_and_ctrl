package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fight Rigging — Enchantment {2}{G}:
//
//	"Hideaway 5 (When this enchantment enters, look at the top five
//	 cards of your library, exile one face down, then put the rest on
//	 the bottom in a random order.)
//	 At the beginning of combat on your turn, put a +1/+1 counter on
//	 target creature you control. Then if you control a creature with
//	 power 7 or greater, you may play the exiled card without paying its
//	 mana cost."
//
// Rabble Rousing's shape (ADR 0091): the hidden card is the one THIS
// Fight Rigging hid, captured as the object as the trigger is built
// (CR 607.2a), and the free play is the same grant (PlayHiddenCard). The
// power check is made after the counter lands, so the counter itself
// can reach 7; it is the counter placement's continuation, because a
// placement can pause on a CR 616 prompt. With the target gone the
// counter is not put, but the check still happens (CR 608.2b).
//
// No simplification.
const fightRiggingLabel = "Fight Rigging — put a +1/+1 counter on target creature you control"

func init() {
	Register(Spec{
		OracleID:     "117d89b2-b4c0-4421-a631-37810c0e0969",
		Name:         "Fight Rigging",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Hideaway("Fight Rigging", 5),
			{
				Watches:   []game.EventKind{game.EventStepBegan},
				AppliesTo: StepBegan(game.StepBeginCombat, true),
				Key:       fightRiggingLabel,
				Targets:   TargetCreature("target creature you control", YouControl()),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, fightRiggingLabel)
					item.Params.Object = game.ObjectRef(game.ObjectRefOf(*source))
					return item
				},
				Effect: fightRiggingResolve,
			},
		},
	})
}

func fightRiggingResolve(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	ref := game.PermissionCardRef(item.Params.Object)
	offer := func(g *game.Game, _ int) error {
		if fightRiggingHasASevenPowerCreature(g, controller) {
			grantHiddenPlay(g, controller, ref, "Fight Rigging — play the exiled card without paying its mana cost")
		}
		return nil
	}
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return offer(g, 0)
	}
	return g.AddCounterByThenForEffect(controller, id, game.CounterPlusOne, 1, offer)
}

func fightRiggingHasASevenPowerCreature(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.CurrentPower() >= 7 {
			return true
		}
	}
	return false
}
