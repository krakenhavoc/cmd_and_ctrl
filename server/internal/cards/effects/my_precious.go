package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const myPreciousOracleID = "2e728381-6db0-4c66-883d-82d718fef833"

func init() {
	Register(Spec{
		OracleID:     myPreciousOracleID,
		Name:         "My Precious",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("hexproof"),
			RestrictAttached(game.CantBeBlocked),
		},
		Activated: []ActivatedAbility{
			EquipPayingAbility("Equip—{2}, Pay 2 life", Plus(ManaCost("{2}"), PayLife(2))),
		},
	})
	Register(Spec{
		OracleID:       myPreciousOracleID + "#1",
		Name:           "Allure of Power",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Draws: 2},
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
