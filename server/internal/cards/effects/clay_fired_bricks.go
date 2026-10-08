package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clay-Fired Bricks // Cosmium Kiln — a transforming artifact (#2124,
// ADR 0137):
//
//	Clay-Fired Bricks — Artifact {1}{W}
//	  "When this artifact enters, search your library for a basic Plains
//	   card, reveal it, put it into your hand, then shuffle. You gain 2
//	   life.
//	   Craft with artifact {5}{W}{W}"
//	Cosmium Kiln — Artifact
//	  "When this artifact enters, create two 1/1 colorless Gnome artifact
//	   creature tokens.
//	   Creatures you control get +1/+1."
//
// The search is The Birth of Meletis's (a BASIC Plains, revealed), then
// the life. Craft with artifact takes another artifact you control or an
// artifact card from your graveyard (the Bricks pay "Exile this
// artifact" and cannot also be the material, CR 118.3). Crafted, the
// Kiln is a new object, so its own enters trigger makes the Gnomes, and
// the anthem pumps them. Requested for the Sami Whammy deck (#2190).
//
// No simplification.
const clayFiredBricksOracleID = "90e1bcef-53af-4c08-b713-bf10dfcbd577"

func init() {
	Register(Spec{
		OracleID:     clayFiredBricksOracleID,
		Name:         "Clay-Fired Bricks",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Clay-Fired Bricks — search for a basic Plains, gain 2 life", clayFiredBricksEnters),
		},
		Activated: []ActivatedAbility{
			Craft("Craft with artifact {5}{W}{W}", "{5}{W}{W}", CraftWith("artifact")),
		},
	})

	Register(Spec{
		OracleID:     clayFiredBricksOracleID + "#1",
		Name:         "Cosmium Kiln",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cosmium Kiln — create two 1/1 Gnome artifact creature tokens", cosmiumKilnGnomes),
		},
		Static: []game.StaticAbility{b16Anthem(b41CreatureYouControlStatic, 1, 1)},
	})
}

// clayFiredBricksEnters searches, and gains the life in the search's
// continuation: the search may stop on a prompt, and "You gain 2 life"
// is printed after it.
func clayFiredBricksEnters(g *game.Game, item *game.StackItem) error {
	controller, source := item.Controller, item.SourceCardID
	return SearchLibrary{
		Player:    controller,
		Predicate: isBasicPlains,
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Clay-Fired Bricks — search for a basic Plains",
		Then: func(g *game.Game, _ []uuid.UUID) error {
			return g.ChangePlayerLifeForEffect(source, controller, 2)
		},
	}.Apply(NewContext(g, item))
}

func cosmiumKilnGnomes(g *game.Game, item *game.StackItem) error {
	return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 colorless Gnome artifact"), N: 2}.Apply(NewContext(g, item))
}
