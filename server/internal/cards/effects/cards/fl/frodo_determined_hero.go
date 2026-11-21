package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const frodoLabel = "Frodo, Determined Hero — you may attach target Equipment you control with mana value 2 or 3 to Frodo"

// Frodo, Determined Hero — Legendary Creature — Halfling Warrior {1}{W}, 2/2
// (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Whenever Frodo enters or attacks, you may attach target Equipment you
//	 control with mana value 2 or 3 to Frodo.
//	 During your turn, prevent all damage that would be dealt to Frodo."
//
// One "enters or attacks" watch (Iron Hills Stalwart's attach of a targeted
// Equipment) whose target clause admits only Equipment of mana value 2 or
// 3, and whose "may" is asked as it resolves (CR 603.5). The prevention is
// Gideon Blackblade's: it applies only while Frodo's controller is the
// active player, and prevents damage from every source.
//
// No simplification.
func init() {
	attach := effects.Optional(effects.Targeting(
		effects.OnAny([]game.EventKind{game.EventETB, game.EventAttack}, effects.Self, frodoLabel, frodoAttach),
		effects.TargetPermanent("target Equipment you control with mana value 2 or 3",
			effects.HasSubtype("Equipment"), effects.YouControl(), effects.ManaValueGE(2), effects.ManaValueLE(3))),
		"Frodo, Determined Hero — attach the Equipment to Frodo?")
	effects.Register(effects.Spec{
		OracleID:     "ae50be07-c8bc-45db-b47d-701c5d9ee633",
		Name:         "Frodo, Determined Hero",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.GapB128PreventDamageToThisOnYourTurn("Frodo, Determined Hero — prevent all damage that would be dealt to him during your turn"),
		},
		Triggered: []game.TriggeredAbility{attach},
	})
}

func frodoAttach(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	equip, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok || !effects.OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	return g.AttachForEffect(equip, game.TargetRef{Kind: game.TargetCard, ID: item.SourceCardID})
}
