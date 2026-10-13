package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Homura, Human Ascendant // Homura's Essence — {4}{R}{R} Legendary
// Creature — Human Monk 4/4 // Legendary Enchantment, a flip card
// (#1892, ADR 0151):
//
//	Top:     "Homura can't block.
//	          When Homura dies, return it to the battlefield flipped."
//	Flipped: "Creatures you control get +2/+2 and have flying and
//	          "{R}: This creature gets +1/+0 until end of turn.""
//
// The one card that says otherwise under CR 110.5b: the dies trigger
// returns the card from its graveyard flipped (ADR 0151 decision 6,
// game.ReturnFromGraveyardFlippedForEffect), under its owner's control
// (CR 400.3), as a new object (CR 400.7) that enters as Homura's
// Essence. If the card left the graveyard in response, nothing returns.
//
// The Essence is a red enchantment (CR 710.1c) that pumps and grants
// flying to every creature its controller controls, and grants each the
// firebreathing row as its own (ADR 0093): the creature's controller
// activates it and "this creature" is the creature.
//
// No simplification.
const (
	homuraOracleID = "fa023c32-2cec-416c-90b7-07b7a19c82d4"
	homuraGrant    = "homuras-essence/firebreathing"
)

func init() {
	Register(Spec{
		OracleID:     homuraOracleID,
		Name:         "Homura, Human Ascendant",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			}, "Homura, Human Ascendant — return it to the battlefield flipped",
				func(g *game.Game, item *game.StackItem) error {
					return ReturnFromGraveyard{
						Target:     item.SourceCardID,
						Dest:       game.ZoneBattlefield,
						Controller: uuid.Nil, // CR 400.3: its owner's control
						Flipped:    true,
					}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:     homuraOracleID + "#1",
		Name:         "Homura's Essence",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: homuraGrant,
			Activated: []ActivatedAbility{{
				Label:   "{R}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{R}"),
				Effect:  thisCreatureUntilEOT("Homura's Essence — +1/+0", 1, 0),
			}},
			Text: "{R}: This creature gets +1/+0 until end of turn.",
		}},
		Static: []game.StaticAbility{
			Labelled(b16Anthem(b16CreaturesYouControl, 2, 2), "Creatures you control get +2/+2."),
			Labelled(KeywordGrant(b16CreaturesYouControl, "flying"), "Creatures you control have flying."),
			Labelled(GrantAbilities(b16CreaturesYouControl, homuraGrant), "Creatures you control have \"{R}: This creature gets +1/+0 until end of turn.\""),
		},
	})
}
