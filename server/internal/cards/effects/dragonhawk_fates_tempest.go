package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonhawk, Fate's Tempest — Legendary Creature — Bird Dragon {3}{R}{R},
// 5/5 (Bloomburrow, #4467):
//
//	"Flying
//	 Whenever Dragonhawk enters or attacks, exile the top X cards of your
//	 library, where X is the number of creatures you control with power 4
//	 or greater. You may play those cards until your next end step. At the
//	 beginning of your next end step, Dragonhawk deals 2 damage to each
//	 opponent for each of those cards that are still exiled."
//
// X is counted once, as the ability resolves, through the layered view
// (Dragonhawk itself, a 5/5, counts). The exile is the impulse exile
// (ExileTopNUntilYourNextEndStep: the window closes as that end step
// begins, and a land is played, spending the land drop). The delayed
// trigger is scheduled for those cards only (the ruling), controlled by
// the ability's controller and bound to their own end step (CR 603.7), and
// carries them as data. It counts the cards that are still in exile when
// it resolves and deals 2 damage for each, as one damage event, whether or
// not Dragonhawk is still on the battlefield (the damage is dealt by
// Dragonhawk as it last existed).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f7f435bc-1716-4412-9505-a1f2f2d83d08",
		Name:            "Dragonhawk, Fate's Tempest",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Dragonhawk, Fate's Tempest — exile the top X cards of your library (X is the number of creatures you control with power 4 or greater); you may play them until your next end step; at the beginning of your next end step, 2 damage to each opponent for each still exiled",
				dragonhawkExile),
		},
	})
}

// dragonhawkExile is the trigger's body.
func dragonhawkExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	controller := ctx.Controller()
	x := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.PowerForComparison() >= 4 {
			x++
		}
	}
	if x <= 0 {
		return nil
	}
	exiled, err := g.ExileTopWithPermissionForEffect(controller, controller, x, game.CastPermission{
		Duration: g.UntilYourNextEndStepDuration(controller),
	})
	if err != nil || len(exiled) == 0 {
		return err
	}
	return ScheduleDelayedTrigger{
		At:                 game.StepEnd,
		ControllerTurnOnly: true,
		Controller:         controller,
		Label:              "Dragonhawk, Fate's Tempest — 2 damage to each opponent for each of those cards still exiled",
		Cards:              exiled,
		Body:               dragonhawkDamageBody,
	}.Apply(ctx)
}

// dragonhawkDamageBody is "2 damage to each opponent for each of those
// cards that are still exiled". An on-disk key.
var dragonhawkDamageBody = game.SimpleDelayedBody("dragonhawk-fates-tempest/damage-per-exiled", func(g *game.Game, item *game.StackItem) error {
	n := 0
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z != nil && z.Kind == game.ZoneExile {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return DamageEachOpponent{N: 2 * n}.Apply(NewContext(g, item))
})
