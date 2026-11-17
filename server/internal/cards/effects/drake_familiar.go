package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drake Familiar — Creature — Drake {1}{U}, 2/1 (Ravnica: City of Guilds,
// #5600):
//
//	"Flying
//	 When this creature enters, sacrifice it unless you return an
//	 enchantment to its owner's hand."
//
// Quickling's trigger over any enchantment on the battlefield, whoever
// controls it: asked as the trigger resolves, and with no enchantment
// anywhere the Drake is sacrificed at once. Naming none, naming one that
// has left, or a return that a replacement sends elsewhere sacrifices the
// Drake. A Drake that has left or come back as a new object is not
// sacrificed (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d144825-594a-454b-9276-fe92e87c19c0",
		Name:            "Drake Familiar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Drake Familiar — sacrifice it unless you return an enchantment to its owner's hand", drakeFamiliarEnters),
		},
	})
}

func drakeFamiliarEnters(g *game.Game, item *game.StackItem) error {
	var enchantments []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsEnchantment() {
			enchantments = append(enchantments, c.InstanceID)
		}
	}
	return sacrificeUnlessYouReturnOne(NewContext(g, item), item, enchantments,
		"Drake Familiar — return an enchantment to its owner's hand (or sacrifice Drake Familiar)",
		drakeFamiliarPicked)
}

// drakeFamiliarPicked returns the chosen enchantment, or sacrifices the
// Drake. An on-disk key: never renamed, never reused.
var drakeFamiliarPicked = askedPick("choose-cards/drake-familiar-return", func(ctx *Context, r game.CardsChosen) error {
	return returnPickedOrSacrifice(ctx, r, drakeFamiliarBounced)
})

// drakeFamiliarBounced sacrifices the Drake when the enchantment did not
// reach a hand. An on-disk key: never renamed, never reused.
var drakeFamiliarBounced = CardTail("drake-familiar/after-bounce", sacrificeSourceUnlessBounced)
