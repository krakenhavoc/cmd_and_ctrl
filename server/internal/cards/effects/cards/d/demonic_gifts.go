package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Demonic Gifts — Instant {1}{B} (Jumpstart 2022, #5948):
//
//	"Until end of turn, target creature gets +2/+0 and gains "When this
//	 creature dies, return it to the battlefield under its owner's control.""
//
// Fake Your Own Death's duration grant (ADR 0093 PR 4): the boost and the
// granted trigger are one effect at one timestamp (GrantAbilitiesFor.Also).
// The granted trigger is the creature's own and returns it untapped; a token
// or a card that has left the graveyard first does not come back (CR 111.7,
// CR 400.7).
//
// No simplification.
const demonicGiftsReturn = "demonic-gifts/return"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "162bcc9d-6edf-4b3b-83e6-2c933e2a5967",
		Name:         "Demonic Gifts",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Grants: []effects.AbilityGrant{{
			Key: demonicGiftsReturn,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisDies("Demonic Gifts — return it to the battlefield", effects.GapB170ReturnThisToBattlefield),
			},
			Text: "When this creature dies, return it to the battlefield under its owner's control.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.GrantAbilitiesFor{
				Target: target,
				Keys:   []string{demonicGiftsReturn},
				Also:   []game.Mod{game.ModifyPTMod(2, 0)},
				Label:  "Demonic Gifts — +2/+0 and it returns when it dies",
			}.Apply(ctx)
		},
	})
}
