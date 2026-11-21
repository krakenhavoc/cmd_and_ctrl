package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Space Amoeba — Creature — Germ {4}{U}, 3/3 (Star Trek Commander,
// #4974):
//
//	"Flying
//	 Whenever this creature attacks, target creature defending player
//	 controls gets -3/-0 until end of turn, then create a token that's a
//	 copy of this creature."
//
// The target is chosen off the Amoeba's own attack
// (TargetOfDefendingPlayer), so attacking a planeswalker or battle names
// that permanent's controller or protector as the defending player. With
// no creature to target the trigger is removed (CR 603.3d), and a target
// gone illegal by resolution makes the whole ability do nothing (CR
// 608.2b): no token either. Otherwise the token is a copy of the Amoeba,
// by its last-known copiable values if it has left (CR 608.2h,
// CreateTokenCopy). The token enters neither tapped nor attacking.
//
// No simplification.
func init() {
	attack := effects.WheneverThisAttacks("Giant Space Amoeba — target creature defending player controls gets -3/-0, then create a token copy",
		giantSpaceAmoebaAttack)
	attack.TargetsFrom = effects.TargetOfDefendingPlayer("creature", "target creature defending player controls", effects.Creature())
	effects.Register(effects.Spec{
		OracleID:        "ff9f3d23-a12f-428b-8001-90dafac0c6d7",
		Name:            "Giant Space Amoeba",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{attack},
	})
}

func giantSpaceAmoebaAttack(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.BoostUntilEOT{Target: id, Power: -3, Label: "Giant Space Amoeba — -3/-0 until end of turn"}).Apply(ctx); err != nil {
		return err
	}
	return effects.CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(ctx)
}
