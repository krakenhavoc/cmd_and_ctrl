package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bishop of Binding — Creature — Vampire Cleric {3}{W}, 1/1 (Rivals of Ixalan,
// S82 OLD batch 218, #5531):
//
//	"When this creature enters, exile target creature an opponent controls
//	 until this creature leaves the battlefield.
//	 Whenever this creature attacks, target Vampire gets +X/+X until end of
//	 turn, where X is the power of the exiled card."
//
// The enters trigger is Banisher Priest's exile, linked to the Bishop (CR 607.2a:
// the card lands in exile with Card.ExiledWith naming this object). The attack
// trigger reads the power of the card exiled with this object as it resolves
// (the card's own printed power in exile); with nothing exiled, or the Bishop
// having left and returned as a new object (CR 400.7), X is 0. The Vampire
// target may be the Bishop itself or any Vampire.
//
// No simplification.
func init() {
	const exileLabel = "Bishop of Binding — exile target creature an opponent controls until this creature leaves the battlefield"
	const pumpLabel = "Bishop of Binding — target Vampire gets +X/+X until end of turn, where X is the power of the exiled card"
	effects.Register(effects.Spec{
		OracleID:     "c2dd6c78-e482-4a3c-ad9c-7dd71b3e34d2",
		Name:         "Bishop of Binding",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(exileLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.ExileUntil{
					Target:     id,
					ThisLeaves: true,
					Linked:     true,
					Label:      "Bishop of Binding — the exiled creature returns when Bishop of Binding leaves the battlefield",
				}.Apply(ctx)
			}), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
			effects.Targeting(effects.On(game.EventAttack, effects.Self, pumpLabel, bishopOfBindingPump),
				effects.TargetCreature("target Vampire", effects.HasSubtype("Vampire"))),
		},
	})
}

// bishopOfBindingPump is the attack trigger: +X/+X, X the power of the exiled card.
func bishopOfBindingPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	x := 0
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		x += c.Power
	}
	if x <= 0 {
		return nil
	}
	return effects.BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Bishop of Binding — +X/+X"}.Apply(ctx)
}
