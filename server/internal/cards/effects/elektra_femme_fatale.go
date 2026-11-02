package effects

import (
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
	Register(Spec{
		OracleID:     "75f2545d-cd31-4404-953f-1110f646228b",
		Name:         "Elektra, Femme Fatale",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(
				WhenThisEnters("Elektra, Femme Fatale — you may have her deal 2 damage to you; when you do, she deals 4 damage to target creature", elektraEnters),
				"Elektra, Femme Fatale — have her deal 2 damage to you?"),
		},
	})
}

func elektraEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 2}).Apply(ctx); err != nil {
		return err
	}
	return ReflexiveTrigger{
		Label:  "Elektra, Femme Fatale — 4 damage to target creature",
		Body:   elektraDamageBody,
		Params: game.EffectParams{Amount: 4},
	}.Apply(ctx)
}
