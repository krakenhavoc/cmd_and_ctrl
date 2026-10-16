package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Springheart Nantuko — Enchantment Creature — Insect Monk {1}{G}, 1/1
// (EDHREC rank 707):
//
//	"Bestow {1}{G}
//	 Enchanted creature gets +1/+1.
//	 Landfall — Whenever a land you control enters, you may pay {1}{G}
//	 if this permanent is attached to a creature you control. If you do,
//	 create a token that's a copy of that creature. If you didn't create
//	 a token this way, create a 1/1 green Insect creature token."
//
// Bestow is ADR 0141's alternative cost, and the pump is PumpAttached.
//
// The landfall trigger asks its question as it resolves. If the
// Nantuko is attached to a creature you control then, the offer is a
// may-pay with a consequence on both answers (QueueMayPayElseForEffect):
// paying creates a token copy of that creature (CreateTokenCopy, its
// copiable values, CR 707.2); declining, or a "yes" that can't be
// funded, creates the Insect. Unattached, or attached to a creature you
// don't control, there is nothing to pay for and the Insect is created
// straight away.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "8a3ad2ef-8bcb-40c0-85de-f03328c2b644",
		Name:             "Springheart Nantuko",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Bestow("{1}{G}")},
		Static:           []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			LandfallPaying("Springheart Nantuko — copy the enchanted creature for {1}{G}, or create a 1/1 Insect",
				game.LandfallPayoff{Tokens: 1}, springheartLandfall),
		},
	})
}

// springheartLandfall is the trigger's resolution.
func springheartLandfall(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	controller := ctx.Controller()
	insect := func(g *game.Game) error {
		return CreateToken{Template: TokenCard("1/1 green Insect"), N: 1}.Apply(NewContext(g, item))
	}
	var host uuid.UUID
	if !sourceIsNewObject(g, item) {
		if h := attachedHostFor(g, item.SourceCardID); h != nil && h.IsCreature() && h.Controller == controller {
			host = h.InstanceID
		}
	}
	if host == uuid.Nil {
		return insect(g)
	}
	return g.QueueMayPayElseForEffect(controller, item.SourceCardID, "{1}{G}",
		"Springheart Nantuko — pay {1}{G} to create a token copy of the enchanted creature? (If you don't, create a 1/1 green Insect.)",
		func(g *game.Game) error {
			if _, ok := TokenCopyTemplate(g, host); !ok {
				return insect(g)
			}
			return CreateTokenCopy{Controller: controller, Copy: host, N: 1}.Apply(NewContext(g, item))
		},
		insect)
}
