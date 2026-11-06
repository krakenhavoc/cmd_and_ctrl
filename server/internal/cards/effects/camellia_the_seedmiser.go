package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Camellia, the Seedmiser — Legendary Creature — Squirrel Warlock
// {1}{B}{G}, 3/3 (Bloomburrow, #4467):
//
//	"Menace
//	 Other Squirrels you control have menace.
//	 Whenever you sacrifice one or more Foods, create a 1/1 green
//	 Squirrel creature token.
//	 {2}, Forage: Put a +1/+1 counter on each other Squirrel you control.
//	 (To forage, exile three cards from your graveyard or sacrifice a
//	 Food.)"
//
// The grant is the tribal keyword lord (TribalKeywordGrant), others only
// and yours only. "One or more Foods" is one trigger per batch of
// sacrifices (OncePerBatch over EventSacrifice, which fires while the
// Food is still on the battlefield to be read as one); a Food sacrificed
// to pay her own forage counts. "{2}, Forage:" is the forage cost (#5184,
// Forage): the activator names the way, and with fewer than three cards
// in the graveyard and no Food the ability can't be activated. The
// counters go on the other Squirrels as it resolves; "each other" is read
// off the ability's source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "312c7a7b-2084-42b0-9b13-0a6d3e43a73d",
		Name:            "Camellia, the Seedmiser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Squirrel"}, Others: true, YoursOnly: true}, "menace"),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b11SacrificedAFood(ev, source, g)
			}, "Camellia, the Seedmiser — create a 1/1 green Squirrel", Do(CreateToken{Template: TokenCard("1/1 green Squirrel"), N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Forage: Put a +1/+1 counter on each other Squirrel you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{2}"), Forage()),
			Effect:  mshB1CounterOnEachOtherYouControl(Subtype("Squirrel")),
		}},
	})
}
