package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hoarding Broodlord — Creature — Dragon {5}{B}{B}{B}, 7/6:
//
//	"Convoke
//	 Flying
//	 When this creature enters, search your library for a card, exile it
//	 face down, then shuffle. For as long as that card remains exiled,
//	 you may play it.
//	 Spells you cast from exile have convoke."
//
// The search exiles face down and grants the play permission over the
// card as it lands, which also makes its controller the card's one
// viewer (the table sees a card back). The permission lasts while the
// card stays in exile, so it can be played on a later turn at its own
// timing and cost.
//
// Caveat: "spells you cast from exile have convoke" is not given. Convoke
// is read off the spell's own card, and no permanent can grant it to
// another spell yet (Dazzling Theater's gap).
//
// Mass Casualty deck request (#3245).
func init() {
	Register(Spec{
		OracleID:        "18e856b5-aab7-4802-8cec-1d44dd95d9f7",
		Name:            "Hoarding Broodlord",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Spells you cast from exile don't get convoke."},
		PrintedKeywords: []string{"flying"},
		TapCost:         Convoke(),
		Purpose:         game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hoarding Broodlord — search for a card, exile it face down; you may play it", hoardingBroodlordEnters),
		},
	})
}

func hoardingBroodlordEnters(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	return SearchLibrary{
		Player:   controller,
		Limit:    1,
		Dest:     game.ZoneExile,
		FaceDown: game.FaceDownPermitted,
		Shuffle:  true,
		Reason:   "Hoarding Broodlord — search your library for a card",
		Then: func(g *game.Game, found []uuid.UUID) error {
			for _, id := range found {
				if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
					continue
				}
				g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
					Player:   controller,
					Zone:     game.ZoneExile,
					Duration: game.WhileInZoneDuration(),
					Source:   item.SourceCardID,
				})
			}
			return nil
		},
	}.Apply(NewContext(g, item))
}
