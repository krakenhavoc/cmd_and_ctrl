package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Princess Yue — Legendary Creature — Human Noble Ally {2}{U}, 3/2
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"When Princess Yue dies, if she was a nonland creature, return this
//	 card to the battlefield tapped under your control. She's a land
//	 named Moon. She gains "{T}: Add {C}." (She's still legendary.)
//	 {T}: Scry 2."
//
// The Enduring cycle's return (glimmer_return.go) with three more words:
//
//   - "If she was a nonland creature" is CR 603.4's intervening-if, read
//     once off the CR 603.10 last-known information: a Moon that is
//     somehow made a creature and dies was a land, and stays in the
//     graveyard.
//   - "She's a land named Moon. She gains '{T}: Add {C}.'" is ADR 0210's
//     entry overwrite (game.EntersAs): her card types are SET to Land
//     (CR 205.1a), so Human, Noble and Ally go with Creature and she is
//     not a creature (the ruling), her supertypes stay (CR 205.4b: still
//     legendary), she is named Moon (layer 3, CR 612.8) and gains the
//     mana ability as a granted bundle. All of it is there from the
//     moment she lands, so a legendary Princess Yue and a legendary Moon
//     coexist (the ruling). She keeps her own abilities, so the Moon
//     still taps to scry 2.
//   - "This card" is the card in the graveyard: one that has left it by
//     the time the trigger resolves stays where it is (CR 608.2b), and
//     "under your control" is the trigger's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     princessYueOracle,
		Name:         "Princess Yue",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:  princessYueMoonMana,
			Text: "{T}: Add {C}.",
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			}},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source) && keywordSliceContains(lki.Types, "Creature") &&
					!keywordSliceContains(lki.Types, "Land")
			}, "Princess Yue — return her to the battlefield tapped as a land named Moon", princessYueReturnsAsMoon),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Scry 2.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 2},
			Cost:    TapCost(),
			Effect:  Do(Scry{N: 2}),
		}},
	})
}

const princessYueOracle = "08c59e5e-d01c-49b7-a812-8ea75e9e945d"

// princessYueMoonMana is the granted "{T}: Add {C}." bundle's key. An
// on-disk identity: never renamed.
const princessYueMoonMana = "princess-yue/moon-mana"

// princessYueReturnsAsMoon is the dies trigger's resolution.
func princessYueReturnsAsMoon(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard || !sameGraveyardObject(z, item) {
		return nil
	}
	return ReturnFromGraveyardTogether{
		Targets:    []uuid.UUID{item.SourceCardID},
		Controller: item.Controller,
		Tapped:     true,
		EntersAs: &game.EntersAs{
			Types:  []string{"Land"},
			Name:   "Moon",
			Grants: []string{princessYueMoonMana},
			Label:  "Princess Yue — a land named Moon with \"{T}: Add {C}.\"",
		},
	}.Apply(NewContext(g, item))
}
