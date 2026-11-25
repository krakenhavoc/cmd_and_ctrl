package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Adverse Conditions — Instant {3}{U} (Battle for Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Tap up to two target creatures. Those creatures don't untap during their
//	 controller's next untap step. Create a 1/1 colorless Eldrazi Scion
//	 creature token. It has "Sacrifice this token: Add {C}.""
//
// Will of the Naga's tap-and-freeze over up to two targets, then the Scion,
// which is made whether or not any target is still legal as long as one was
// legal on resolution (with every target gone the spell is countered on
// resolution, CR 608.2b, but "up to two" with none chosen still resolves).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7a5a81f7-1597-4972-8565-fd2a5667df99",
		Name:            "Adverse Conditions",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Targets:         effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := effects.B751TapFreezeTargets(ctx, uuid.Nil, "Adverse Conditions"); err != nil {
				return err
			}
			return effects.CreateToken{Template: effects.B28EldraziScionToken(), N: 1}.Apply(ctx)
		},
	})
}
