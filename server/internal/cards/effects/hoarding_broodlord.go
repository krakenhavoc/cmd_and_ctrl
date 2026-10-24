package effects

import (
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
// "Spells you cast from exile have convoke" is a granted tap cost
// (convoke_grant.go): any spell its controller casts from exile, the
// card this found or any other, may be convoked.
//
// No simplification. Mass Casualty deck request (#3245).
func init() {
	Register(Spec{
		OracleID:        "18e856b5-aab7-4802-8cec-1d44dd95d9f7",
		Name:            "Hoarding Broodlord",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		TapCost:         Convoke(),
		Purpose:         game.Purpose{Tutors: 1},
		GrantedTapCosts: []game.GrantedTapCost{
			SpellsYouCastHaveConvoke("Spells you cast from exile have convoke.", FromZone(game.ZoneExile)),
		},
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
		ThenTail: game.KeyedTail{Ref: hoardingBroodlordPermission, Player: controller, Source: item.SourceCardID},
	}.Apply(NewContext(g, item))
}

// hoardingBroodlordPermission is "you may cast it for as long as it
// remains exiled": a permission over each card the search exiled, held by
// Tail.Player. An on-disk key.
var hoardingBroodlordPermission = game.RegisterTail("hoarding-broodlord/cast-permission", func(g *game.Game, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player:   o.Tail.Player,
			Zone:     game.ZoneExile,
			Duration: game.WhileInZoneDuration(),
			Source:   o.Tail.Source,
		})
	}
	return nil
})
