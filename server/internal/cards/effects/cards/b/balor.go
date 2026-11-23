package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Balor — Creature — Demon {3}{R}{R}, 5/5 (Commander Legends: Battle for
// Baldur's Gate):
//
//	"Flying
//	 Whenever this creature attacks or dies, choose one or more. Each mode must
//	 target a different player.
//	 • Target opponent draws three cards, then discards three cards at random.
//	 • Target opponent sacrifices a nontoken artifact of their choice.
//	 • This creature deals damage to target opponent equal to the number of cards in
//	   their hand."
//
// One ability with two conditions (Gorex's attack-or-dies shape) that is a modal
// trigger (CR 700.2): one to three bullets chosen as it goes on the stack (CR
// 603.3c), every "target opponent" clause Distinct as Vindictive Lich's, so no two
// chosen bullets can name the same player. The first bullet draws then discards,
// as two steps so the discard waits for a prompt the draw asks (CR 608.2c); the
// discard is at random (CR 701.9b). The second is an edict over nontoken artifacts
// (the opponent chooses, CR 701.21a). The third deals damage from this creature,
// last-known information if it has died (CR 608.2h), equal to the cards in the
// target's hand as it resolves.
//
// No simplification.
func init() {
	targetOpp := func() *game.TargetSpec {
		return effects.Distinct(effects.TargetPlayer("target opponent", effects.Opponent()))
	}
	trigger := effects.OnAny([]game.EventKind{game.EventAttack, game.EventLTB},
		func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return effects.AttackDeclared(ev, source) || effects.CardDied(ev, source)
		},
		"Balor — choose one or more; each mode must target a different player",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseN("Choose one or more", 1, 3,
		effects.ModeSteps("Target opponent draws three cards, then discards three cards at random.", targetOpp(),
			effects.TmtB6ModeTargetDraws(3),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return ctx.Game.DiscardRandomKeyedForEffect(t.ID, 3, game.KeyedTail{})
			})),
		effects.ModeDoing("Target opponent sacrifices a nontoken artifact of their choice.", targetOpp(),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				ctx.Game.PlayerSacrificesForEffect(ctx.Source(), t.ID,
					effects.SacrificeSpec("a nontoken artifact", effects.And(effects.Artifact(), effects.Not(effects.IsTokenPredicate()))),
					"Balor — sacrifice a nontoken artifact")
				return nil
			}),
		effects.ModeDoing("This creature deals damage to target opponent equal to the number of cards in their hand.", targetOpp(), balorDamage),
	)
	effects.Register(effects.Spec{
		OracleID:        "6f931c1e-f1d2-4dc7-a653-c29adfc8fb2c",
		Name:            "Balor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}

// balorDamage is the third bullet: this creature deals damage to the bullet's target
// opponent equal to the cards in their hand as it resolves. A named function, its
// target read its own, so it is never mistaken for a gate on the second bullet's
// sacrifice.
func balorDamage(item *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	p := ctx.Game.PlayerByIDForEffect(t.ID)
	if p == nil || p.Hand == nil {
		return nil
	}
	return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: p.Hand.Size()}.Apply(ctx)
}
