package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// My Precious // Allure of Power — an Adventure card (CR 715, The Hobbit,
// #4454), oracle 2e728381…:
//
//	My Precious — Legendary Artifact — Equipment {3}
//	  "Equipped creature has hexproof and can't be blocked.
//	   Equip—{2}, Pay 2 life."
//	Allure of Power — Instant — Adventure {1}{B}
//	  "As an additional cost to cast this spell, sacrifice a creature.
//	   Draw two cards. (Then exile this card. You may cast the artifact
//	   later from exile.)"
//
// Whispersilk Cloak's grant (hexproof for shroud) with Inventor's Axe's
// equip that pays something besides mana: the life is an ordinary cost
// component, paid at activation (CR 602.2b). Allure of Power is Village
// Rites on the Adventure half; the sacrifice is a cost, so it is paid
// before the spell is on the stack and a counterspell does not return the
// creature.
//
// No simplification.

// myPreciousOracleID is shared by both faces of the card.
const MyPreciousOracleID = "2e728381-6db0-4c66-883d-82d718fef833"

func init() {
	effects.Register(effects.Spec{
		OracleID:     MyPreciousOracleID,
		Name:         "My Precious",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.GrantToAttached("hexproof"),
			effects.RestrictAttached(game.CantBeBlocked),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipPayingAbility("Equip—{2}, Pay 2 life", effects.Plus(effects.ManaCost("{2}"), effects.PayLife(2))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:       MyPreciousOracleID + "#1",
		Name:           "Allure of Power",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Draws: 2},
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
