package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Return to Action — Instant {1}{B} (Kamigawa: Neon Dynasty, S82 #5493):
//
//	"Until end of turn, target creature gets +1/+0 and gains lifelink
//	 and "When this creature dies, return it to the battlefield tapped
//	 under its owner's control.""
//
// Fake Your Own Death's duration grant (ADR 0093 PR 4): the +1/+0, the
// lifelink and the granted trigger are one effect at one timestamp. The
// granted trigger is the creature's own; a token that dies does not come
// back (CR 111.7) and a card that left the graveyard first is not the
// object that died (CR 400.7).
//
// No simplification.
const returnToActionReturn = "return-to-action/return"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c002b4d-bb6c-45a5-88e1-1df2958dff9f",
		Name:         "Return to Action",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Grants: []effects.AbilityGrant{{
			Key: returnToActionReturn,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisDies("Return to Action — return it tapped",
					effects.ReturnThisCreatureFromGraveyard(nil, nil)),
			},
			Text: "When this creature dies, return it to the battlefield tapped under its owner's control.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.GrantAbilitiesFor{
				Target: target,
				Keys:   []string{returnToActionReturn},
				Also:   []game.Mod{game.ModifyPTMod(1, 0), game.AddKeywordsMod("lifelink")},
				Label:  "Return to Action — +1/+0, lifelink and it returns when it dies",
			}.Apply(ctx)
		},
	})
}
