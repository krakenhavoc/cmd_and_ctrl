package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaima, the Fractured Calm — Legendary Creature — Spirit {2}{R}{G}, 3/3
// (Kamigawa: Neon Dynasty):
//
//	"At the beginning of your end step, goad each creature your opponents
//	 control that's enchanted by an Aura you control. Put a +1/+1 counter on
//	 Kaima for each creature goaded this way. (Until your next turn, the
//	 goaded creatures attack each combat if able and attack a player other
//	 than you if able.)"
//
// The set is fixed as the ability resolves: each creature an opponent
// controls with an Aura Kaima's controller controls attached. Each is goaded
// by the controller (Havoc Eater's goadAndScheduleClear, CR 701.15a), and
// the counters are one for each of those still on the battlefield, put on
// Kaima if it is still the same object there (a creature that was already
// goaded is goaded again, so it counts).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d8f9b58-d011-4ed2-99ae-c626243fab26",
		Name:         "Kaima, the Fractured Calm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Kaima, the Fractured Calm — goad each creature your opponents control that's enchanted by an Aura you control; put a +1/+1 counter on Kaima for each creature goaded this way",
				kaimaEndStep),
		},
	})
}

func kaimaEndStep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var goaded []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller || !c.IsCreature() {
			continue
		}
		for _, a := range g.BattlefieldCardsForEffect() {
			if a.IsAura() && a.Controller == item.Controller && a.IsAttachedTo(c.InstanceID) {
				goaded = append(goaded, c.InstanceID)
				break
			}
		}
	}
	if len(goaded) == 0 {
		return nil
	}
	if err := effects.GoadAndScheduleClear(ctx, goaded); err != nil {
		return err
	}
	src := item.SourceCardID
	if ctx.IsNewSourceObject(src) || !effects.OnBattlefield(g, src) {
		return nil
	}
	return effects.AddCounter{Target: src, Kind: game.CounterPlusOne, N: len(goaded)}.Apply(ctx)
}
