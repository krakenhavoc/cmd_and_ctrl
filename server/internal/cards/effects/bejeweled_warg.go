package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bejeweled Warg — Creature — Wolf {1}{G}, 3/2 (The Hobbit, HOB):
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, choose one —
//	 • Put a +1/+1 counter on target Wolf you control.
//	 • Create a Treasure token."
//
// Felidar Retreat's shape on a combat-damage trigger: the trigger does
// nothing itself and each bullet is a ModeDoing body. Only the counter
// bullet targets; its Wolf (the Warg included) is picked as the trigger goes
// on the stack and re-checked at resolution (CR 608.2b).
//
// No simplification.
func init() {
	damage := WheneverThisDealsCombatDamageToAPlayer("Bejeweled Warg — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	damage.Modes = ChooseOne(
		ModeDoing("Put a +1/+1 counter on target Wolf you control.",
			TargetCreature("target Wolf you control", YouControl(), OfCreatureType("Wolf")),
			PlusOneCounterOnTheModesTarget),
		ModeDoing("Create a Treasure token.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Controller: ctx.Controller(), Template: TreasureToken(), N: 1}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:        "051ff7e0-dd00-4467-8796-a5d1c21934ed",
		Name:            "Bejeweled Warg",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{damage},
	})
}
