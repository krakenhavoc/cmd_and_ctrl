package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Esika, God of the Tree // The Prismatic Bridge — a modal double-faced
// card (oracle 92023a5d…), one file for both faces because neither is
// much without the other.
//
// Front, Legendary Creature — God {1}{G}{G}, 1/4:
//
//	"Vigilance
//	 {T}: Add one mana of any color.
//	 Other legendary creatures you control have vigilance and "{T}: Add
//	 one mana of any color.""
//
// Back, Legendary Enchantment {W}{U}{B}{R}{G} (key "<oracle>#1", the
// MDFC face keyspace mdfc_lands.go and Deluge of the Dead use):
//
//	"At the beginning of your upkeep, reveal cards from the top of your
//	 library until you reveal a creature or planeswalker card. Put that
//	 card onto the battlefield and the rest on the bottom of your library
//	 in a random order."
//
// The Bridge is the shared reveal-until sentence (#745): the run is
// revealed, the creature or planeswalker enters through the CR 614
// pipeline without being cast, and the rest go to the bottom in an
// order drawn from the game's seeded RNG.
//
// DECLARED SIMPLIFICATION on the front face, weaker than printed: the
// grant to other legendary creatures is not implemented. Vigilance
// could be granted, but the "{T}: Add one mana of any color" half is a
// granted MANA ABILITY, and the engine can grant keywords but not
// abilities (the granted-ability seam). Granting half of the static
// would make the text read as if it worked, so the whole grant is
// omitted and declared; Esika herself is a whole vigilant mana dork.
func init() {
	Register(Spec{
		OracleID:        esikaGodOfTheTreeOracleID,
		Name:            "Esika, God of the Tree",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Other legendary creatures you control don't gain vigilance or the mana ability."},
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
	Register(Spec{
		OracleID:     esikaGodOfTheTreeOracleID + "#1",
		Name:         "The Prismatic Bridge",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("The Prismatic Bridge — reveal until a creature or planeswalker card and put it onto the battlefield",
				Do(RevealUntilThenPutOntoBattlefield{
					Match:  prismaticBridgeHit,
					Reason: "The Prismatic Bridge — revealed until a creature or planeswalker card",
				})),
		},
	})
}

const esikaGodOfTheTreeOracleID = "92023a5d-a143-4950-a71b-d736e6b8e959"

// prismaticBridgeHit is "a creature or planeswalker card".
func prismaticBridgeHit(c game.Card) bool { return c.IsCreature() || c.IsPlaneswalker() }
