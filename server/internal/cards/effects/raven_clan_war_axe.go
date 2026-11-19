package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raven Clan War-Axe — Artifact — Equipment {1}{R}{W} (Assassin's Creed,
// #5948):
//
//	"When this Equipment enters, you may search your library and/or graveyard
//	 for a card named Eivor, Battle-Ready, reveal it, and put it into your
//	 hand. If you search your library this way, shuffle.
//	 Equipped creature gets +2/+0 and has trample.
//	 Equip {2}"
//
// Delivery Moogle's "library and/or graveyard" search with a name for the
// stated quality and a "you may" in front (Optional): the zones are asked
// first and the library is shown, shuffled and announced only if the player
// chose it (CR 701.23a). A search for a quality may find nothing (CR 701.23b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d306de95-a80e-4e85-b7d2-36c6ea769b15",
		Name:         "Raven Clan War-Axe",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			GrantToAttached("trample"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Raven Clan War-Axe — you may search your library and/or graveyard for a card named Eivor, Battle-Ready",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     func(c game.Card) bool { return c.Name == "Eivor, Battle-Ready" },
						Dest:          game.ZoneHand,
						Limit:         1,
						Reveal:        true,
						Shuffle:       true,
						AlsoGraveyard: true,
						Optional:      true,
						Reason:        "Raven Clan War-Axe — a card named Eivor, Battle-Ready",
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
