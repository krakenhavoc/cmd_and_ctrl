package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Marwyn, the Nurturer — Legendary Creature — Elf Druid {2}{G}, 1/1
// (EDHREC rank 1197):
//
//	"Whenever another Elf you control enters, put a +1/+1 counter on
//	 Marwyn.
//	 {T}: Add an amount of {G} equal to Marwyn's power."
//
// The Elf deck's scaling mana dork. The trigger is the "another
// <type> you control enters" shape, reading the Elf's effective
// subtypes so a changeling counts; the counter goes on through
// AddCounter, so Hardened Scales applies. The mana ability is the
// Gaea's Cradle shape with the count read off Marwyn's own CURRENT
// power at activation — counters and anthems both count, a -1/-1
// counter subtracts — and a power of zero adds nothing and still
// taps. Summoning sickness applies to the tap, as for every creature
// mana ability (CR 302.6); the engine enforces it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee35de1c-aef1-4bd4-85fd-fe77bc927790",
		Name:         "Marwyn, the Nurturer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, "Elf")
			}, "Marwyn — put a +1/+1 counter on Marwyn", func(g *game.Game, item *game.StackItem) error {
				return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
			}),
		},
		ManaAbilities: []ManaAbility{{
			Cost:  ManaAbilityCost{Tap: true},
			Label: "Add an amount of {G} equal to Marwyn's power",
			ProducedFunc: func(g *game.Game, _ uuid.UUID, source uuid.UUID) string {
				c, ok := g.LookupCardForEffect(source)
				if !ok {
					return ""
				}
				return strings.Repeat("{G}", c.CurrentPower())
			},
		}},
	})
}
