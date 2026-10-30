package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// U.S.Agent, John Walker — {3}{W/B} Legendary Creature — Human Soldier Hero
// 3/2 (Marvel Super Heroes, #4455):
//
//	"When U.S.Agent enters, create a colorless Equipment artifact token
//	 named Sturdy Shield with "Equipped creature gets +1/+2" and equip {2}.
//	 Attach it to U.S.Agent."
//
// Iron Hills Blacksmith's Axe, with the attach step: the token is created
// through the keyed creation, and the shared attach-to-host tail attaches
// each created token to the host riding the tail (here U.S.Agent). Should
// U.S.Agent have left the battlefield by then, the attach is skipped and the
// shield stays unattached (CR 701.3b). The Sturdy Shield is a token template
// of its own in token_catalog.go; its slug is an on-disk identity.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2b6fa8b6-4865-4eb6-974b-00f6b16b6f0f",
		Name:         "U.S.Agent, John Walker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("U.S.Agent, John Walker — create a Sturdy Shield Equipment token and attach it to U.S.Agent",
				usAgentCreateShield),
		},
	})
}

// SturdyShieldEquipmentToken is the Sturdy Shield as PRINTED.
func SturdyShieldEquipmentToken() game.Card { return tokenFromCatalog(printedSturdyShieldToken) }

// printedSturdyShieldToken is the catalog's entry for the Sturdy Shield:
// a colorless Equipment artifact token, "Equipped creature gets +1/+2" and
// equip {2}. Its slug is an on-disk identity and is never renamed.
func printedSturdyShieldToken() tokenTemplate {
	return tokenTemplate{
		Slug: "sturdy-shield-equipment",
		Card: game.Card{
			Name:     "Sturdy Shield",
			TypeLine: "Token Artifact — Equipment",
		},
		Text:      "Equipped creature gets +1/+2.\nEquip {2}",
		Static:    []game.StaticAbility{PumpAttached(1, 2)},
		Activated: activatedShapes([]ActivatedAbility{EquipAbility("{2}")}),
	}
}

// usAgentCreateShield creates the token, then attaches it to the ability's
// source.
func usAgentCreateShield(g *game.Game, item *game.StackItem) error {
	if sourceIsNewObject(g, item) || !onBattlefield(g, item.SourceCardID) {
		return CreateToken{Template: SturdyShieldEquipmentToken(), N: 1}.Apply(NewContext(g, item))
	}
	ctx := NewContext(g, item)
	tail := ctx.Tail(roleAttachToHost)
	tail.Carry = []uuid.UUID{item.SourceCardID}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{{
			Template: ctx.TextToken(SturdyShieldEquipmentToken()),
			Count:    1,
		}},
	}, tail)
}
