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
// LAYER, stated. CR 613 puts "is every creature type" in layer 4 and
// this grants a keyword in layer 6 — see the note on
// GrantAllCreatureTypesUntilEOT in tribal.go. Unobservable today, one
// line to move if a "creatures lose all abilities" effect ever ships.
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
func init() {
	Register(Spec{
		OracleID: "9b2cdbed-c733-409b-b0e4-2c8960c25111",
		Name:     "Maskwood Nexus",
		Static: []game.StaticAbility{
			TribalKeywordGrant(
				TribeFilter{YoursOnly: true},
				game.KeywordChangeling,
			),
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
