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
// changeling-specific: the grant is the same keyword a printed
// changeling carries, so a lord, Coat of Arms, Cavern of Souls' spend
// restriction and the wire badge all see it without learning a second
// question.
//
// LAYER 4, even though the marker it writes is a keyword string.
// See AllCreatureTypesGrant in tribal.go: declaring it in layer 6
// would order it against every lord's keyword half by timestamp, and
// a Goblin Chieftain that entered first would grant haste before the
// Bear became a Goblin. There is a test that enters the Nexus LAST
// for exactly that reason.
//
// SANDBOX SIMPLIFICATION — the second sentence is NOT implemented.
// "Creature spells you control and creature cards you own that aren't
// on the battlefield" is a continuous effect reaching into every
// zone, and the layer engine maintains effective characteristics for
// battlefield permanents only; Card.HasSubtype off the battlefield is
// a method on a card with no game to consult, so there is nowhere to
// ask whether a Nexus is out. A printed changeling is unaffected by
// this — CR 702.73a is a characteristic-defining ability on the card
// itself, which is exactly why it works in every zone and this does
// not. The direction is weaker than printed: the Nexus still makes
// your battlefield every type, it just doesn't reach your hand.
//
// What that costs in practice is the tribal-tutor and
// cast-restriction cases: a Cavern of Souls named for Elf will not
// pay for a Bear in your hand even with a Nexus out. Closing it needs
// the layer engine to compute a characteristic for non-battlefield
// cards, which is a real engine change and not a card change.
//
// DECLARED GAPS, CR 613.8. "Creatures you control" reads the Creature
// type, which other layer-4 effects add and remove, so the Nexus
// depends on them and should apply after them whatever the
// timestamps. The layer engine orders layer 4 by timestamp only, so:
//
//   - a permanent that becomes a creature AFTER the Nexus entered is
//     not every creature type. Crew is the common case: the crew
//     effect is stamped when the ability resolves, so a Vehicle
//     crewed with a Nexus already out misses every lord. Every
//     turn-scoped "becomes a creature" effect has the same shape.
//   - a permanent that switches its own creature type off and entered
//     AFTER the Nexus (The Warring Triad under eight graveyard cards,
//     a slumbering Arixmethes) keeps the every-type marker, because
//     the marker rides the ability list the self type-strip does not
//     clear.
//
// Those pairs are pinned, skipped, in layer_dependency_pairs_test.go.
//
// DECLARED GAP, not ordering: the every-type marker is stored as the
// changeling keyword (AllCreatureTypesGrant), so layer-6 ability
// removal wipes it. A creature you control that has lost all its
// abilities (Kenrith's Transformation attached before the Nexus
// entered) is not every creature type, although the Nexus's effect
// comes from the Nexus and the rules keep it. Only that order is
// wrong. Both ability-loss Auras on a creature here (Kenrith's
// Transformation, Darksteel Mutation) also set the creature types in
// layer 4, independently of the Nexus, so timestamps decide: an Aura
// attached AFTER the Nexus entered leaves an Elk (or Insect) and
// nothing else, which is the rules answer and what the engine gives.
func init() {
	Register(Spec{
		OracleID:     "9b2cdbed-c733-409b-b0e4-2c8960c25111",
		Name:         "Maskwood Nexus",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Creature spells you cast and creature cards you own outside the battlefield aren't every creature type — only creatures you control on the battlefield are.",
			"A permanent that becomes a creature after Maskwood Nexus is on the battlefield, such as a Vehicle you crew, isn't every creature type.",
			"The Warring Triad or a sleeping Arixmethes, Slumbering Isle that entered after Maskwood Nexus still counts as every creature type while it isn't a creature.",
			"A creature you control that lost all its abilities before Maskwood Nexus entered, such as one already enchanted by Kenrith's Transformation, isn't every creature type.",
		},
		Static: []game.StaticAbility{
			AllCreatureTypesGrant(TribeFilter{YoursOnly: true}),
		},
		Activated: []ActivatedAbility{{
			Label: "{3}, {T}: Create a 2/2 blue Shapeshifter with changeling",
			Cost:  Plus(ManaCost("{3}"), TapCost()),
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
