package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystic Sanctuary — Land — Island:
//
//	"({T}: Add {U}.)
//	 This land enters tapped unless you control three or more other
//	 Islands.
//	 When this land enters untapped, you may put target instant or
//	 sorcery card from your graveyard on top of your library."
//
// The blue Zendikar Rising "sanctuary" land, and half of the
// Sanctuary–Fetchland loop every blue deck that plays it is really
// playing. Three pieces, all of them existing vocabulary:
//
//   - The mana ability is not declared. The type line says Island, so
//     the engine derives the intrinsic {U} from the land type
//     (CR 305.6) — and that derivation reads EFFECTIVE subtypes, so a
//     Sanctuary under a Blood Moon is a Mountain and taps for {R},
//     which is correct and is why declaring the ability by hand would
//     be wrong.
//   - "Unless you control three or more other ISLANDS" is the
//     slowland condition with a land-type filter; the count reads
//     effective subtypes too, so an Urborg-ed board does not turn
//     every land into an Island but a Prismatic Omen board does.
//   - "When this land enters untapped" is the Gingerbread Cabin
//     condition: the harvester runs with the entry replacement
//     already applied, so the source's tapped flag is the answer.
//
// The trigger is optional and it TARGETS, so an empty graveyard means
// no trigger at all (CR 603.3d) rather than a prompt with nothing to
// point at, and a card that leaves the graveyard in response makes it
// fizzle (CR 608.2b).
//
// The put goes through the shared exit primitive, so a commander card
// in the graveyard gets the CR 903.9 offer on the way to the library.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17b60106-a4c7-410a-8ac3-ec8e74e29a7c",
		Name:         "Mystic Sanctuary",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersTappedUnless(otherLandsWithSubtypeAtLeast("island", 3)),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					On(game.EventETB, selfEnteredUntapped,
						"Mystic Sanctuary — put an instant or sorcery on top of your library",
						mysticSanctuaryPutOnTop),
					"Mystic Sanctuary — put target instant or sorcery card from your graveyard on top of your library?"),
				TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					Or(Instant(), Sorcery()), YouOwn()),
			),
		},
	})
}

// mysticSanctuaryPutOnTop tucks the chosen card to the top of its
// owner's library. A target that has gone is not an error.
func mysticSanctuaryPutOnTop(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return g.TuckToLibraryForEffect(t.ID, false)
	}
	return nil
}
