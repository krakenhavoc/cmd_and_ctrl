package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bill Ferny, Bree Swindler — Legendary Creature — Human Rogue {1}{U}, 2/1
// (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Whenever Bill Ferny becomes blocked, choose one —
//	 • Create a Treasure token.
//	 • Target opponent gains control of target Horse you control. If they
//	   do, remove Bill Ferny from combat and create three Treasure tokens."
//
// A "becomes blocked" modal trigger (Grenzo's modal shape). The second
// bullet has two targets, a player and a Horse; control passes with no
// stated duration (CR 611.2a), and only if the Horse is then that player's
// does Bill Ferny leave combat (CR 506.4) and his controller get the three
// Treasures. A Horse that left in response, or an opponent who left,
// changes nothing.
//
// No simplification.
func init() {
	t := effects.On(game.EventBecomesBlocked, effects.Self, "Bill Ferny, Bree Swindler — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("Create a Treasure token.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("Target opponent gains control of target Horse you control. If they do, remove Bill Ferny from combat and create three Treasure tokens.",
			effects.Clauses(
				effects.TargetPlayer("target opponent", effects.Opponent()),
				effects.TargetCreature("target Horse you control", effects.YouControl(), effects.HasSubtype("Horse")),
			),
			billFernyHorse),
	)
	effects.Register(effects.Spec{
		OracleID:     "ebbe5d82-a7a4-4f4b-b038-095021d9a6a8",
		Name:         "Bill Ferny, Bree Swindler",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered:    []game.TriggeredAbility{t},
	})
}

func billFernyHorse(item *game.StackItem, ctx *effects.Context, occ int) error {
	var opp, horse game.TargetRef
	for _, t := range ctx.ModeTargets(occ) {
		if !ctx.IsTargetLegal(t) {
			continue
		}
		switch t.Kind {
		case game.TargetPlayer:
			opp = t
		case game.TargetCard:
			horse = t
		}
	}
	if opp.Kind != game.TargetPlayer || horse.Kind != game.TargetCard {
		return nil
	}
	if err := (effects.GainControl{Target: horse.ID, Controller: opp.ID, Duration: game.IndefiniteDuration(),
		Label: "Bill Ferny, Bree Swindler — gain control (no stated duration)"}).Apply(ctx); err != nil {
		return err
	}
	c, ok := ctx.Game.LookupCardForEffect(horse.ID)
	if !ok || c.Controller != opp.ID {
		return nil
	}
	if effects.OnBattlefield(ctx.Game, item.SourceCardID) {
		if err := ctx.Game.RemoveFromCombatForEffect(item.SourceCardID); err != nil && err != game.ErrNotInCombat {
			return err
		}
	}
	return effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: 3}.Apply(ctx)
}
