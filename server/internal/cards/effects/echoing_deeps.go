package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Echoing Deeps — Land — Cave (The Lost Caverns of Ixalan, #4470):
//
//	"You may have this land enter tapped as a copy of any land card in a
//	 graveyard, except it's a Cave in addition to its other types.
//	 {T}: Add {C}."
//
// Body Double's enters-as-a-copy over graveyards (#4797, CR 707.2) with
// Vesuva's tapped rider: the copy_target prompt offers every land card in
// every graveyard, and the copy takes that card's printed values, the
// Cave subtype added (CR 707.9b), and enters tapped only when the copy is
// taken. Declining, or a graveyard with no land card, leaves Echoing Deeps
// as printed and untapped, tapping for {C}. The copy keeps no memory of
// being Echoing Deeps, so its own {T}: Add {C} is the copied land's, not
// this one (CR 707.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ef88214-f46d-473e-a55b-795a647e2f03",
		Name:         "Echoing Deeps",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf("Echoing Deeps", anyLandCardInAGraveyard,
				func(ev *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					ev.EntersTapped = true
					v.AddSubtype("Cave")
				}),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}",
		}},
	})
}

// anyLandCardInAGraveyard is the candidate list of "any land card in a
// graveyard": every seat's graveyard, tokens never being cards there.
func anyLandCardInAGraveyard(g *game.Game, _ uuid.UUID, _ uuid.UUID) []uuid.UUID {
	if g == nil {
		return nil
	}
	var out []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.IsLand() {
				out = append(out, c.InstanceID)
			}
		}
	}
	return out
}
