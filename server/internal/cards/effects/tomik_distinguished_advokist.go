package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tomik, Distinguished Advokist — Legendary Creature — Human Advisor
// {W}{W}, 2/3:
//
//	"Flying
//	 Lands on the battlefield and land cards in graveyards can't be the
//	 targets of spells or abilities your opponents control.
//	 Your opponents can't play land cards from graveyards."
//
// Two rule gates, one per sentence, and the card waited until both
// existed (ADR 0109's PR 5 amendment):
//
//   - THE TARGETING SENTENCE is ADR 0109 §6's TargetingRestrictions,
//     read at the engine's two targeting choke points: a land on the
//     battlefield, or a land card in any graveyard, is never offered to
//     a spell or ability an opponent of Tomik's controller controls,
//     can't be announced as its target (CR 601.2c), and an opponent's
//     spell already aimed at one when Tomik arrives loses that target
//     at resolution (CR 608.2b). Tomik's controller's own spells and
//     abilities are untouched. A permanent is judged by its effective
//     types, so a land animated into a creature is still a land, and a
//     creature that is not a land is not protected.
//   - THE LAND-PLAY SENTENCE is ADR 0109 §4's land-play gate: an
//     opponent may not play a land card out of a graveyard (Crucible of
//     Worlds, Ramunap Excavator), whichever graveyard it is in. Lands
//     played from hand are untouched.
//
// "Your opponents" is read off Tomik's CONTROLLER at each question, so
// a Tomik that changes hands protects its new controller's lands.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9895a33f-9bbd-4440-8c1a-0d401431b77f",
		Name:            "Tomik, Distinguished Advokist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		TargetingRestrictions: []game.TargetingRestriction{
			OpponentsCantTarget("Lands on the battlefield and land cards in graveyards can't be the targets of spells or abilities your opponents control.",
				Land(), game.ZoneBattlefield, game.ZoneGraveyard),
		},
		LandPlayRestrictions: []game.LandPlayRestriction{
			OpponentsCantPlayLandsFrom("Your opponents can't play land cards from graveyards.", game.ZoneGraveyard),
		},
	})
}
