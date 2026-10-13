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
// The front face's static (#3172) is two layer 6 statics over "other
// legendary creatures you control": a keyword grant of vigilance, and the
// quoted mana ability as an ADR 0093 ability grant (a granted mana ability
// is the same Cryptolith Rite shape). "Other" is by instance, so a second
// Esika is a recipient; "legendary" is read off the effective supertypes.
func init() {
	Register(Spec{
		OracleID:        esikaGodOfTheTreeOracleID,
		Name:            "Esika, God of the Tree",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Grants:          []AbilityGrant{AnyColorManaGrant(esikaGrant)},
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(otherLegendaryCreaturesYouControl, "vigilance"),
				"Other legendary creatures you control have vigilance."),
			Labelled(GrantAbilities(otherLegendaryCreaturesYouControl, esikaGrant),
				"Other legendary creatures you control have \"{T}: Add one mana of any color.\""),
		},
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

// esikaGrant is the bundle key of Esika's granted mana ability.
const esikaGrant = "esika-god-of-the-tree/any-color"

// otherLegendaryCreaturesYouControl is Esika's recipient set.
func otherLegendaryCreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
		target.IsCreature() && target.IsLegendary()
}
