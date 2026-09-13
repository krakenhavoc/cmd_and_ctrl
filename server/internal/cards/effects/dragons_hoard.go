package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragon's Hoard — Artifact {3} (EDHREC rank 1194):
//
//	"Whenever a Dragon you control enters, put a gold counter on this
//	 artifact.
//	 {T}, Remove a gold counter from this artifact: Draw a card.
//	 {T}: Add one mana of any color."
//
// The Dragon deck's three-mana rock. The trigger is the landfall
// shape with "Dragon" for "land", reading effective subtypes so a
// changeling counts; the mana ability is Commander's Sphere's
// five-way pick with no identity narrowing — "any color" is what the
// card says.
//
// DECLARED SIMPLIFICATION: the draw ability is not implemented.
// "Remove a gold counter from this artifact" is a cost component the
// engine cannot express — AbilityCost carries tap, sacrifice, mana,
// life, loyalty and crew, and nothing that removes a counter (the
// Walking Ballista gap) — and shipping the draw with the cost
// deferred to resolution would let a proliferate in response bank a
// counter the printed card had already spent, which is the #259
// direction. The counters still accumulate on the card, visibly, so
// the ability is whole the day a counter-removal cost lands; until
// then the Hoard is a rock that counts Dragons, which is the Mines
// of Moria posture and runs the weaker way.
func init() {
	Register(Spec{
		OracleID:     "8cf77dc4-763b-41b7-a5da-0ef2734f08e6",
		Name:         "Dragon's Hoard",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Removing a gold counter to draw a card isn't implemented — the counters build up but can't be spent."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.HasSubtype("Dragon")
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Dragon's Hoard — put a gold counter on it",
					func(g *game.Game, item *game.StackItem) error {
						return AddCounter{Target: item.SourceCardID, Kind: "gold", N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			Produced:                "{W|U|B|R|G}",
			Label:                   "Add one mana of any color",
			IgnoreCommanderIdentity: true,
		}},
	})
}
