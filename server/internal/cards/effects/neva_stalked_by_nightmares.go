package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Neva, Stalked by Nightmares — Legendary Creature — Human Noble {2}{W}{B},
// 2/2 (Wilds of Eldraine, #4471):
//
//	"Menace
//	 When Neva enters, return target creature or enchantment card from your
//	 graveyard to your hand.
//	 Whenever an enchantment you control is put into a graveyard from the
//	 battlefield, put a +1/+1 counter on Neva, then scry 1."
//
// The enters trigger is Sibsig Muckdraggers' graveyard return with the
// pick widened to a creature or an enchantment card you own. The second
// trigger is Ashiok's Reaper's (socB13EnchantmentYouControlWasPutIntoAGraveyard):
// the enchantment is read as it last existed on the battlefield, so an
// Aura that falls off with its creature and a Role token both count. The
// counter goes on Neva only while she is still the permanent the trigger
// came from (CR 400.7); the scry happens either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ba5895e1-f12b-43c7-b47c-376588d9d2d2",
		Name:            "Neva, Stalked by Nightmares",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Neva — return target creature or enchantment card to your hand", returnTargetCardToHand),
				TargetCardInGraveyard("target creature or enchantment card from your graveyard", YouOwn(), Or(Creature(), Enchantment())),
			),
			On(game.EventLTB, socB13EnchantmentYouControlWasPutIntoAGraveyard,
				"Neva — put a +1/+1 counter on Neva, then scry 1",
				nevaCounterThenScry),
		},
	})
}

// nevaCounterThenScry is "put a +1/+1 counter on Neva, then scry 1".
func nevaCounterThenScry(g *game.Game, item *game.StackItem) error {
	if err := putCounterOnSelf(g, item); err != nil {
		return err
	}
	return Scry{N: 1}.Apply(NewContext(g, item))
}
