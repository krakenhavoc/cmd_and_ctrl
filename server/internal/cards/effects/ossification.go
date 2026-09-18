package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ossification — Enchantment — Aura {1}{W} (EDHREC rank 4295):
//
//	"Enchant basic land you control
//	 When this Aura enters, exile target creature or planeswalker an
//	 opponent controls until this Aura leaves the battlefield."
//
// Oblivion Ring for two mana, with the drawback moved from the mana
// cost to the deckbuilding: it needs a basic land, and a Commander
// mana base full of duals and fetch targets may not have one spare.
// The reward is that it is one mana cheaper than every other card
// printed with this text.
//
// "Until this Aura leaves the battlefield" is TWO abilities, not one —
// an entry trigger that exiles, and a leave trigger that returns
// whatever that exile took. The card is written that way because the
// rules are: CR 610.3, the exile is a one-shot, and the return is its
// own triggered ability with its own record.
//
// That record is the event log, read back through b27ExiledWith on the
// entry trigger's stack label — which is why the label is a const
// shared by both halves rather than two string literals. A card that
// something else moved out of exile in the meantime is not pulled back
// out of wherever it went, because the newer move closes the older
// record.
//
// The exiled permanent returns under its OWNER's control (CR 610.3),
// not under the Aura controller's: destroying an Ossification on your
// own Ravenous Chupacabra gives it back to you, and destroying one on
// an opponent's commander gives it back to them.
//
// The enchant clause is a REAL restriction, re-run every turn by the
// CR 704.5m legality check: a basic land that stops being one — or
// stops being yours — takes the Aura to the graveyard, which returns
// the exiled permanent. That is the card's honest weakness.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e29bfd62-286f-4982-813f-7086573c333b",
		Name:         "Ossification",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant basic land you control", b41BasicLandYouControl()),
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Targets: TargetPermanent("target creature or planeswalker an opponent controls",
					b41CreatureOrPlaneswalkerAnOpponentControls()),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					// b27ExileChosenTarget is Duplicant's body: exile
					// whatever the pick stamped into the item, if it is
					// still legal (CR 608.2b).
					return game.NewTriggeredItem(source, b41OssificationExileLabel, b27ExileChosenTarget)
				},
			},
			On(game.EventLTB, Self, "Ossification — return the exiled card",
				b41ReturnCardsExiledWithToTheBattlefield(b41OssificationExileLabel)),
		},
	})
}
