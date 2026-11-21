package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Two-Handed Axe // Sweeping Cleave — an adventure card (Commander Legends:
// Battle for Baldur's Gate, S82 GAP #5948):
//
//	Two-Handed Axe {2}{R}, Artifact — Equipment:
//	"Whenever equipped creature attacks, double its power until end of turn.
//	 Equip {1}{R}"
//
//	Sweeping Cleave {1}{R}, Instant — Adventure:
//	"Target creature you control gains double strike until end of turn."
//
// The Equipment is Baseball Bat's attack trigger over the creature it is
// attached to as the trigger resolves (CR 701.10b, DoublePowerUntilEOT: the
// doubling is read then, and an Axe that has fallen off by then doubles
// nothing); the adventure half is the "<oracle_id>#1" spec, and the engine
// exiles the resolved adventure with permission to cast the Equipment
// (CR 715.3d).
func init() {
	const oracleID = "b012cec9-2aff-4dfb-8db1-61e830c33926"
	effects.Register(effects.Spec{
		OracleID:     oracleID,
		Name:         "Two-Handed Axe",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack,
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttachedCreatureAttacked(ev, source)
				},
				"Two-Handed Axe — double equipped creature's power until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					host := effects.AttachedHostFor(g, item.SourceCardID)
					if host == nil {
						return nil
					}
					return effects.DoublePowerUntilEOT(effects.NewContext(g, item), host.InstanceID, "Two-Handed Axe — double power")
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}{R}")},
	})
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Sweeping Cleave",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "double strike")), Answers: game.AnswerCombatGrant},
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Sweeping Cleave — double strike"}.Apply(ctx)
		},
	})
}
