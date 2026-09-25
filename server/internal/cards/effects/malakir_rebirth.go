package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malakir Rebirth // Malakir Mire — modal double-faced card. This file
// is the FRONT face, Instant {B}:
//
//	"Choose target creature. You lose 2 life. Until end of turn, that
//	 creature gains "When this creature dies, return it to the
//	 battlefield tapped under its owner's control.""
//
// The back face, Malakir Mire, is registered with the MDFC land cycle
// in mdfc_lands.go under "<oracle>#1".
//
// A duration grant (ADR 0093 PR 4, #1584) — Feign Death without the
// counter, plus the life. The life is lost only if the spell resolves:
// with its one target gone it does nothing at all (CR 608.2b).
//
// No simplification.
const malakirRebirthReturn = "malakir-rebirth/return"

func init() {
	Register(Spec{
		OracleID:     "a731e87b-8d99-4b64-8ee3-8e540d652366",
		Name:         "Malakir Rebirth",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Grants: []AbilityGrant{{
			Key: malakirRebirthReturn,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Malakir Rebirth — return it tapped",
					returnThisCreatureFromGraveyard(nil, nil)),
			},
			Text: "When this creature dies, return it to the battlefield tapped under its owner's control.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2); err != nil {
				return err
			}
			return GrantAbilitiesFor{
				Target: target,
				Keys:   []string{malakirRebirthReturn},
				Label:  "Malakir Rebirth — until end of turn, it returns when it dies",
			}.Apply(ctx)
		},
	})
}
