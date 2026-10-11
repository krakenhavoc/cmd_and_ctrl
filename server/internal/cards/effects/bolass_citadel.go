package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bolas's Citadel — Legendary Artifact {3}{B}{B}{B}:
//
//	"You may look at the top card of your library any time.
//	 You may play lands and cast spells from the top of your library.
//	 If you cast a spell this way, pay life equal to its mana value
//	 rather than pay its mana cost.
//	 {T}, Sacrifice ten nonland permanents: Each opponent loses 10 life."
//
// The card #765 is named for, and the one that needs all three halves
// of CR 401.5 at once: a permission over a POSITION rather than a
// card, the visibility that makes the position playable, and a price
// that is computed from whatever card happens to be there.
//
//   - The permission is standing and opens the top card only. The card
//     on top changes every draw, mill, shuffle and cast, so nothing is
//     stored: the engine asks the battlefield, then asks the library
//     what is on top, every time.
//
//   - "You may look at the top card of your library any time" is
//     LibraryTopOwner — private to its owner, unlike Courser of
//     Kruphix's revealed top. Without it the permission would open a
//     card you cannot see, which is not a play at all.
//
//   - "Pay life equal to its mana value rather than pay its mana cost"
//     is an alternative cost (CR 118.9) with a life component (CR
//     119.4), computed per card. It is a COST, so a player without the
//     life cannot claim it — a Citadel cast is refused rather than
//     resolving and killing you, which is what CR 119.4 means by
//     paying life down to exactly zero being legal and further being
//     not.
//
//   - "lands AND spells", so the permission is not cast-only and no
//     filter narrows it. A land played this way still spends the
//     turn's land drop (CR 305.2) and pays no life, because a land has
//     no mana cost to replace.
//
//   - "{T}, Sacrifice ten nonland permanents" is Plus(TapCost(),
//     SacrificeN(10, ..., Nonland())): a fixed-N clause over the
//     controller's own nonland permanents, paid whole at activation
//     (CR 602.2b) and refused with nothing paid when fewer than ten are
//     named. The Citadel is itself a nonland permanent, so it may be one
//     of the ten; no "other" is printed and none is imposed. The cost
//     has been paid before the ability resolves, so removing the Citadel
//     in response changes nothing, and each opponent loses 10 life (a
//     loss, not damage, CR 119.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2bd111bb-ce02-414c-b5b7-e0e037d8d96b",
		Name:         "Bolas's Citadel",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice ten nonland permanents: Each opponent loses 10 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(TapCost(), SacrificeN(10, "ten nonland permanents", Nonland())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return eachOpponentLosesLife(g, item, 10)
			},
		}},
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{func() game.CastPermission {
			p := PlayFromTopOfYourLibrary(game.PermissionFilter{},
				"Pay life equal to its mana value (Bolas's Citadel)")
			p.AltCostKey = "bolas_citadel"
			p.LifeEqualToManaValue = true
			return p
		}()},
	})
}
