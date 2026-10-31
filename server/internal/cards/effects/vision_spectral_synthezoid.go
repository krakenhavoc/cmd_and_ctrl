package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vision, Spectral Synthezoid — {6}{U}{U} Legendary Artifact Creature —
// Robot Hero, 2/5 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Once during each of your turns, you may cast a noncreature or Robot
//	 spell from your hand without paying its mana cost."
//
// Flying is the printed keyword. The cast is Omniscience's granted free
// offer from the hand, limited to once each turn and to its controller's
// own turns (Access Maze's OnceEachTurn and YourTurnOnly), the limit
// belonging to this Vision object. "Noncreature or Robot" is a union the
// offer's one filter cannot say, so it is the offer's SpellsAnyOf: a spell
// that is either a noncreature spell or has the Robot subtype is reached,
// by ONE offer and ONE once-each-turn claim, so one cast of either kind
// spends the turn's use. Both are judged on the face being cast; a land is
// never offered one (it is played, not cast).
//
// No simplification.
const grantedAltCostVisionSpectral = "granted-free-vision-spectral-synthezoid"

func init() {
	offer := func() game.AlternativeCost {
		return game.AlternativeCost{
			Key:   grantedAltCostVisionSpectral,
			Label: "Cast it without paying its mana cost",
		}
	}
	Register(Spec{
		OracleID:        "1a44261a-2b6a-45dc-91a4-b2eb219ba6fb",
		Name:            "Vision, Spectral Synthezoid",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{{
			Offer: offer(),
			Zones: []game.ZoneKind{game.ZoneHand},
			SpellsAnyOf: []game.PermissionFilter{
				{NoncreatureOnly: true},
				{CreatureType: "Robot"},
			},
			SpellsText:   "noncreature or Robot spells",
			OnceEachTurn: true,
			YourTurnOnly: true,
		}},
	})
}
