package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maskwood Nexus — Artifact for {4}:
//
//	"Creatures you control are every creature type. The same is true
//	 for creature spells you control and creature cards you own that
//	 aren't on the battlefield.
//	 {3}, {T}: Create a 2/2 blue Shapeshifter creature token with
//	 changeling."
//
// The card that turns any pile into a tribal deck, and the one that
// proves S26's "every creature type" mechanism is shared rather than
// changeling-specific: the grant is the same layer-4 fact a printed
// changeling carries, so a lord, Coat of Arms, Cavern of Souls' spend
// restriction and the wire badge all see it without learning a second
// question.
//
// LAYER 4, even though it reads like a keyword. See
// AllCreatureTypesGrant in tribal.go: declaring it in layer 6 would
// order it against every lord's keyword half by timestamp, and a
// Goblin Chieftain that entered first would grant haste before the
// Bear became a Goblin. There is a test that enters the Nexus LAST for
// exactly that reason.
//
// The second sentence is two more layer-4 statics over objects off the
// battlefield (#3862, game/zone_layers.go): one over creature SPELLS
// whose controller is the Nexus' controller, one over creature CARDS
// its controller owns in a hand, library, graveyard, exile or the
// command zone. So a Cavern of Souls named for Elf pays for a Bear cast
// from your hand, a "search for an Elf card" finds one, and a
// commander in the command zone is every creature type too. A printed
// changeling needs none of this, because CR 702.73a is a
// characteristic-defining ability on the card itself.
//
// "Creatures you control" reads the Creature type, which other
// layer-4 effects add and remove, so the Nexus DEPENDS on them under
// CR 613.8a and applies after them whichever entered first. A Vehicle
// crewed while the Nexus was already out is every creature type; a
// permanent that switches its own creature type off (The Warring
// Triad under eight graveyard cards, a slumbering Arixmethes) is not,
// whenever it arrived. Both orders of both pairs are pinned in
// layer_dependency_pairs_test.go.
//
// The property itself is a layer-4 FACT on the Characteristic, not
// the changeling keyword in the ability list (#670), which is what
// makes a creature that lost all its abilities in layer 6 still every
// creature type — the Nexus' own 2021-02-05 ruling. A layer-4 subtype
// SET newer than the Nexus still overwrites it, because setting the
// subtypes is what "is an Elk" means: an Aura attached AFTER the
// Nexus entered leaves an Elk (or an Insect) and nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b2cdbed-c733-409b-b0e4-2c8960c25111",
		Name:         "Maskwood Nexus",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			AllCreatureTypesGrant(TribeFilter{YoursOnly: true}),
			{
				Label:        "Creature spells you control are every creature type.",
				Layer:        game.Layer4Type,
				AffectsZones: SpellsOnTheStack(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && target.IsCreature()
				},
				Apply: applyAllCreatureTypes,
			},
			{
				Label:        "Creature cards you own that aren't on the battlefield are every creature type.",
				Layer:        game.Layer4Type,
				AffectsZones: CardsOffTheBattlefield(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Owner == source.Controller && target.IsCreature()
				},
				Apply: applyAllCreatureTypes,
			},
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Create a 2/2 blue Shapeshifter with changeling",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{
					Controller: item.Controller,
					Template:   BlueShapeshifterToken(),
					N:          1,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
