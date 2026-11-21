package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infernal Captor — Creature — Devil Rogue {3}{R}, 3/3 (Modern Horizons
// 3):
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, gain control of target
//	 artifact or creature until end of turn. Untap that permanent. It
//	 gains haste until end of turn."
//
// Exploit is the shared constructor (exploit.go): the payoff is a
// reflexive trigger raised only when a creature really went, its target
// chosen then. The three steps are Act of Treason's, in printed order
// (threatenAndGrant), over an artifact or a creature; a target that is
// gone by resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f0dec6c-797f-4dc3-9994-066c00cbf1a4",
		Name:         "Infernal Captor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Exploit("Infernal Captor", infernalCaptorStealBody),
		},
	})
}

// infernalCaptorStealBody is "When this creature exploits a creature,
// gain control of target artifact or creature until end of turn. Untap
// that permanent. It gains haste until end of turn." An on-disk key:
// never renamed, never reused.
var infernalCaptorStealBody = game.ReflexiveBody("infernal-captor/steal", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.ThreatenAndGrant(ctx, id, "Infernal Captor", "haste")
}), effects.ConstTargets(func() *game.TargetSpec {
	return effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature()))
}))
