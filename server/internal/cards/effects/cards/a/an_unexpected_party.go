package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// An Unexpected Party // At the Door — an Adventure card (CR 715,
// The Hobbit, #4454), oracle 38503d34…:
//
//	An Unexpected Party — Enchantment {2}{W}{W}
//	  "As this enchantment enters, choose a creature type.
//	   Creatures you control of the chosen type get +2/+2."
//	At the Door — Sorcery — Adventure {X}{2}{W}
//	  "Create X 2/2 red Dwarf creature tokens. (Then exile this card.
//	   You may cast the enchantment later from exile.)"
//
// The enchantment is Radiant Destiny's chosen-type anthem (the choice is
// made as it enters, CR 614.12, and read live by the layer 7c static);
// "creatures you control", so it pumps no other player's. At the Door
// reads X as announced; X = 0 makes nothing.
//
// No simplification.

// anUnexpectedPartyOracleID is shared by both faces of the card.
const AnUnexpectedPartyOracleID = "38503d34-f4f9-4eeb-9b1e-4fec1df921b3"

func init() {
	effects.Register(effects.Spec{
		OracleID:     AnUnexpectedPartyOracleID,
		Name:         "An Unexpected Party",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("An Unexpected Party"),
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Chosen: true, YoursOnly: true}, 2, 2),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     AnUnexpectedPartyOracleID + "#1",
		Name:         "At the Door",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("2/2 red Dwarf"), N: ctx.X()}.Apply(ctx)
		},
	})
}
