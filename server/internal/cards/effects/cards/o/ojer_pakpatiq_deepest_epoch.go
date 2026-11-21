package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ojer Pakpatiq, Deepest Epoch — Legendary Creature — God {2}{U}{U},
// 4/3, the front face of a transforming card whose back is Temple of
// Cyclical Time:
//
//	"Flying
//	 Whenever you cast an instant spell from your hand, it gains
//	 rebound. (Exile it as it resolves. At the beginning of your next
//	 upkeep, you may cast it from exile without paying its mana cost.)
//	 When Ojer Pakpatiq dies, return it to the battlefield tapped and
//	 transformed under its owner's control with three time counters on
//	 it."
//
// The cast trigger is ThatSpellGains over the stack step of the layer
// pass (ADR 0107 PR 4), the same shape as Taigam, Ojutai Master. The
// dies trigger is ReturnFromGraveyard{Transformed: true} (#1900, ADR
// 0079 amendment of 2026-10-08), with the counters riding the entry
// event so a counter-doubler sees them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        OjerPakpatiqOracleID,
		Name:            "Ojer Pakpatiq, Deepest Epoch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastFromYourHand(effects.Instant()),
				"Ojer Pakpatiq — that spell gains rebound",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ThatSpellGains{Keywords: []string{game.KeywordRebound}}.Apply(effects.NewContext(g, item))
				}),
			effects.OjerDiesReturnTransformed("Ojer Pakpatiq", map[string]int{"time": 3}),
		},
	})
}

// Temple of Cyclical Time — Land, the back face of Ojer Pakpatiq:
//
//	"(Transforms from Ojer Pakpatiq, Deepest Epoch.)
//	 {T}: Add {U}. Remove a time counter from this land.
//	 {2}{U}, {T}: Transform this land. Activate only if it has no time
//	 counters on it and only as a sorcery."
//
// The counter comes off as part of the mana ability, after the mana is
// produced (a ManaAbility.Rider); a land with no counters still taps
// for {U} and removes nothing. The transform is an ordinary activated
// ability behind a Condition, and sorcery speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     OjerPakpatiqOracleID + "#1",
		Name:         "Temple of Cyclical Time",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "{T}: Add {U}. Remove a time counter from this land.",
			Rider:    removeTimeCounterRider,
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{U}, {T}: Transform this land. Activate only if it has no time counters on it and only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			SorcerySpeed: true,
			Condition:    sourceHasNoTimeCounters,
			Effect:       effects.Do(effects.TransformThis{}),
		}},
	})
}

const OjerPakpatiqOracleID = "34ef174e-1b3d-43d5-9f72-3d35befbdd7f"

// removeTimeCounterRider is "Remove a time counter from this land",
// the second sentence of the mana ability. Nothing to remove is not an
// error.
func removeTimeCounterRider(g *game.Game, _, source uuid.UUID) error {
	if c, ok := g.LookupCardForEffect(source); !ok || c.Counters["time"] <= 0 {
		return nil
	}
	return g.AddCounterForEffect(source, "time", -1)
}

// sourceHasNoTimeCounters gates Temple of Cyclical Time's transform:
// "Activate only if it has no time counters on it".
func sourceHasNoTimeCounters(g *game.Game, _, source uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(source)
	return ok && c.Counters["time"] <= 0
}
