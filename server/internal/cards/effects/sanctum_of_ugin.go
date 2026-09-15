package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sanctum of Ugin — Land (EDHREC rank 2940):
//
//	"{T}: Add {C}.
//	 Whenever you cast a colorless spell with mana value 7 or
//	 greater, you may sacrifice this land. If you do, search your
//	 library for a colorless creature card, reveal it, put it into
//	 your hand, then shuffle."
//
// The Eldrazi deck's chain tutor. A plain colourless land with a
// cast trigger: the controller's colorless spell at mana value
// seven or more (read off the stack) fires a "you may" — the
// trigger prompt — and on yes the Sanctum is sacrificed at
// resolution if it is still on the battlefield, and only then the
// search runs: a colorless creature card, revealed, to hand, then
// shuffle. A Sanctum that left in response cannot be sacrificed,
// so "if you do" fails and there is no search, as printed.
//
// Rides the catalog-wide deterministic search pick when only one
// card qualifies — the S22 chooser asks only when there is a choice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72cb5dcd-9b24-435c-921a-3766108374c4",
		Name:         "Sanctum of Ugin",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b27ColorlessSpellWithManaValueAtLeastCastByYou(ev, source, g, 7)
			}, "Sanctum of Ugin — sacrifice it, then search for a colorless creature card", b27SacrificeSelfThenTutorColorlessCreature), "Sanctum of Ugin — sacrifice it to search for a colorless creature card?"),
		},
	})
}
