package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Covetous Castaway // Ghostly Castigator (#1855, ADR 0107 §4) — a
// disturb card.
//
// Front face, Creature — Human {1}{U}, 1/3:
//
//	"When this creature dies, mill three cards.
//	 Disturb {3}{U}{U} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Creature — Spirit, 3/4:
//
//	"Flying
//	 When this creature enters, you may shuffle up to three target cards
//	 from your graveyard into your library.
//	 If Ghostly Castigator would be put into a graveyard from anywhere,
//	 exile it instead."
//
// The mill lands the Castaway beside the cards it milled, and the
// disturb cast takes it out again. The Castigator's "you may" is the
// trigger's optional prompt; the up-to-three clause may name no card.
// The targets still in the graveyard as it resolves go into the
// library as one move, then the library is shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         CovetousCastawayOracleID,
		Name:             "Covetous Castaway",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Disturb("{3}{U}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Covetous Castaway — mill three cards", effects.Do(effects.MillCards{N: 3})),
		},
	})
	castigator := effects.WhenThisEnters(ghostlyCastigatorLabel, effects.ShuffleTargetsIntoYourLibrary)
	castigator.Targets = effects.TargetCardInGraveyard("up to three target cards from your graveyard", effects.YouOwn()).WithCount(0, 3)
	effects.Register(effects.Spec{
		OracleID:        CovetousCastawayOracleID + "#1",
		Name:            "Ghostly Castigator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.Optional(castigator, ghostlyCastigatorLabel+"?")},
		Replacements:    []game.ReplacementEffect{effects.DisturbedExile("Ghostly Castigator")},
	})
}

const (
	CovetousCastawayOracleID = "8f3e6554-eb8a-4096-81a3-2411186d9cb4"
	ghostlyCastigatorLabel   = "Ghostly Castigator — shuffle up to three target cards from your graveyard into your library"
)
