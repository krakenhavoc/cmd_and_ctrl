package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zuko's Exile — Instant — Lesson {5} (EDHREC rank 4534):
//
//	"Exile target artifact, creature, or enchantment. Its controller
//	 creates a Clue token. (It's an artifact with "{2}, Sacrifice this
//	 token: Draw a card.")"
//
// Five COLOURLESS mana for unconditional exile at instant speed. The
// rate is bad and the colour is the point: a deck with no removal in
// its identity — mono-green, mono-blue, an artifact deck that wants
// to keep its own colours open — can still answer a commander, and
// exile answers the ones that come back.
//
// The Clue is the compensation that makes it an evenly-traded card
// rather than a Vindicate. It goes to the EXILED PERMANENT'S
// CONTROLLER, not to you — so pointing this at your own permanent to
// dodge a removal spell is a real (and sometimes correct) line that
// also draws you a card.
//
// The controller is read BEFORE the exile, which is the only place
// that information exists once the permanent is gone. A target that
// left in response fizzles the spell entirely, Clue included
// (CR 608.2b): there is no controller to give one to. The Clue is
// created only if the exile actually landed — ExileTarget's `Then`
// reports whether it did, so a permanent saved by a replacement
// effect produces no Clue.
//
// The Lesson subtype is flavour here (no Learn card in the catalog
// fetches it yet) and rides the type line from Scryfall.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5cf18398-9552-415e-a61f-2c149e35ec3d",
		Name:         "Zuko's Exile",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target artifact, creature, or enchantment",
			effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(item.Targets[0].ID)
			if !ok {
				return nil
			}
			owner := victim.Controller
			return effects.ExileTarget{
				Target:   item.Targets[0].ID,
				ThenTail: effects.WithCarry(ctx.ItemTail(zukosExileExiled), owner),
			}.Apply(ctx)
		},
	})
}

// zukosExileExiled is the clause after the exile (#4253 slice 9).
var zukosExileExiled = effects.CardTail("zukos-exile/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	owner := effects.TailCarry0(o)
	exiled := len(o.IDs) == 1

	if !exiled {
		return nil
	}
	return effects.CreateToken{Controller: owner, Template: effects.ClueToken(), N: 1}.Apply(ctx)
})
