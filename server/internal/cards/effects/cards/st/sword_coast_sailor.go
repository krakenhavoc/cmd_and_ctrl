package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sword Coast Sailor — Legendary Enchantment — Background {1}{U}
// (EDHREC rank 5966):
//
//	"Commander creatures you own have "Whenever this creature attacks
//	 a player, if no opponent has more life than that player, this
//	 creature can't be blocked this turn.""
//
// An ADR 0093 grant to each commander creature you own, under anyone's
// control (choose_a_background.go). The intervening if is read as the
// attack is declared and as the trigger resolves (CR 603.4); it
// resolves in the declare attackers step, before blockers. "Can't be
// blocked" is a restriction carried on the creature for the turn
// (Artful Dodge's RestrictUntilEOT), so nothing can block it.
//
// No simplification.
const swordCoastSailorGrant = "sword-coast-sailor/unblockable"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "71248bf6-a6a2-440e-9313-a7abce05a84b",
		Name:         "Sword Coast Sailor",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: swordCoastSailorGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisAttacksAPlayerNoOpponentRicher("Sword Coast Sailor — can't be blocked this turn", func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.RestrictUntilEOT{
						Target:       ctx.Source(),
						Restrictions: game.CantBeBlocked,
						Label:        "Sword Coast Sailor — can't be blocked",
					}.Apply(ctx)
				}),
			},
			Text: "Whenever this creature attacks a player, if no opponent has more life than that player, this creature can't be blocked this turn.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(swordCoastSailorGrant)},
	})
}
