package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orochi Soul-Reaver — Creature — Snake Ninja Rogue {5}{B} 5/4 (Outlaws of
// Thunder Junction Commander, #5429):
//
//	"Ninjutsu {3}{B} ({3}{B}, Return an unblocked attacker you control to
//	 hand: Put this card onto the battlefield from your hand tapped and
//	 attacking.)
//	 Whenever one or more creatures you control deal combat damage to a
//	 player, create a Treasure token and manifest the top card of that
//	 player's library. (Put it onto the battlefield face down as a 2/2
//	 creature. Turn it face up any time for its mana cost if it's a
//	 creature card.)"
//
// Ninjutsu is the shared keyword (CR 702.49a). The trigger is one per
// damaged player per combat damage step, however many creatures connected
// (CR 603.2c): three creatures hitting two opponents are two triggers. "A
// player" is any player, so damage your creatures deal to you (redirected)
// counts too, and then the manifest is the top card of your own library.
// The manifest is the two-player door (ManifestTopOfLibrary,
// game/manifest_library.go): you control the face-down 2/2 and the
// damaged player still owns it. The Treasure is made first, as printed,
// and is made even when that library is empty.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6d9ab641-ba99-4b27-a84e-8c2ae57a2fc1",
		Name:         "Orochi Soul-Reaver",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Orochi Soul-Reaver — create a Treasure and manifest the top card of that player's library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (CreateToken{Template: TreasureToken(), N: 1}).Apply(ctx); err != nil {
						return err
					}
					if item.Trigger == nil {
						return nil
					}
					return ManifestTopOfLibrary{LibraryOwner: item.Trigger.Event.Target}.Apply(ctx)
				}),
		},
	})
}
