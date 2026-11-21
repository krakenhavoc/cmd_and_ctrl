package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elektra, Femme Fatale — Legendary Creature — Human Assassin Hero {3}{R},
// 3/2 (Marvel Super Heroes Commander, #4455):
//
//	"When Elektra enters, you may have her deal 2 damage to you. When you
//	 do, she deals 4 damage to target creature."
//
// The "you may" is the parent trigger's optional prompt. Once taken, Elektra
// deals the 2 damage to her controller (a real damage event, with Elektra as
// the source), and the reflexive trigger (CR 603.12) goes on the stack to
// choose its creature then and deal the 4 damage. The reflexive half is
// mandatory. The damage source is Elektra by her object id, so if she has
// left by then her last-known information is used (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75f2545d-cd31-4404-953f-1110f646228b",
		Name:         "Elektra, Femme Fatale",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.WhenThisEnters("Elektra, Femme Fatale — you may have her deal 2 damage to you; when you do, she deals 4 damage to target creature", elektraEnters),
				"Elektra, Femme Fatale — have her deal 2 damage to you?"),
		},
	})
}

func elektraEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 2}).Apply(ctx); err != nil {
		return err
	}
	return effects.ReflexiveTrigger{
		Label:  "Elektra, Femme Fatale — 4 damage to target creature",
		Body:   effects.ElektraDamageBody,
		Params: game.EffectParams{Amount: 4},
	}.Apply(ctx)
}
