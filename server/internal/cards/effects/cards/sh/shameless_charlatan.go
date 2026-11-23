package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shameless Charlatan — Legendary Enchantment — Background {1}{U} (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Commander creatures you own have "{2}{U}: This creature becomes a copy of
//	 another target creature.""
//
// A granted activated ability (ADR 0093). The commander becomes a copy with no
// duration (CR 611.2a, CopyIndefinite: Unstable Shapeshifter's), of another
// target creature judged at announce and again at resolution; it copies the
// target's copiable values (CR 707.2), and "another" excludes the commander
// itself by object. Choose a Background is deck construction (ADR 0144).
//
// No simplification.
const shamelessCharlatanGrant = "shameless-charlatan/copy"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9b08567-398a-46ba-b0e8-ade65fec83c6",
		Name:         "Shameless Charlatan",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: shamelessCharlatanGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{2}{U}: This creature becomes a copy of another target creature.",
				Cost:    effects.ManaCost("{2}{U}"),
				Targets: effects.Another(effects.TargetCreature("another target creature")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					return effects.BecomeCopy{Targets: []uuid.UUID{ctx.Source()}, Of: ts[0].ID, Duration: effects.CopyIndefinite,
						Label: "Shameless Charlatan — becomes a copy"}.Apply(ctx)
				},
			}},
			Text: "{2}{U}: This creature becomes a copy of another target creature.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(shamelessCharlatanGrant)},
	})
}
