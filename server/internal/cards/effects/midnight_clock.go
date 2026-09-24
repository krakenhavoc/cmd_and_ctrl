package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Midnight Clock — Artifact {2}{U}:
//
//	"{T}: Add {U}.
//	 {2}{U}: Put an hour counter on this artifact.
//	 At the beginning of each upkeep, put an hour counter on this
//	 artifact.
//	 When the twelfth hour counter is put on this artifact, shuffle
//	 your hand and graveyard into your library, then draw seven
//	 cards. Exile this artifact."
//
// The counter kind is a free-form string (game/counter_types.go —
// unknown names round-trip without validation); "hour" needs no
// constant of its own for one card. Both counter-placing abilities
// share one closure since AddCounter needs the source's own instance
// ID, which is only known at resolution time (item.SourceCardID).
//
// Caveat: the twelfth-counter payoff isn't implemented — there is no
// primitive yet that shuffles a hand AND a graveyard into a library
// together (teferi_akosa_of_zhalfir.go's shuffle helper is card-
// specific and single-zone). The mana ability and both hour-counter
// abilities work; the clock accumulates counters with no payoff.
func init() {
	Register(Spec{
		OracleID:     "c68faebc-b2cd-461b-b93e-e1fcd4816810",
		Name:         "Midnight Clock",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The twelfth-hour-counter payoff — shuffle your hand and graveyard into your library, draw seven, then exile this artifact — isn't implemented. The mana ability and the hour-counter abilities work."},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:  "{2}{U}: Put an hour counter on this artifact",
			Cost:   ManaCost("{2}{U}"),
			Effect: midnightClockPutHourCounter,
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(_ game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return true
			}, "Midnight Clock — put an hour counter", midnightClockPutHourCounter),
		},
	})
}

// midnightClockPutHourCounter is shared by the activated ability and
// the upkeep trigger — both are "put an hour counter on this
// artifact" and nothing else. Package-level so neither stack item
// captures a *Card or *Game.
func midnightClockPutHourCounter(g *game.Game, item *game.StackItem) error {
	return AddCounter{Target: item.SourceCardID, Kind: "hour", N: 1}.Apply(NewContext(g, item))
}
