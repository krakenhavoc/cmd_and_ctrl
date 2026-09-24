package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yahenni, Undying Partisan — Legendary Creature — Aetherborn Vampire
// {2}{B}, 2/2 (Edea steal-and-sac deck, #1565):
//
//	"Haste
//	 Whenever a creature an opponent controls dies, put a +1/+1
//	 counter on Yahenni.
//	 Sacrifice another creature: Yahenni gains indestructible until
//	 end of turn."
//
// "A creature an OPPONENT controls" is judged on the dying creature's
// controller as it last existed on the battlefield (diedCreature reads
// it off the card in its new zone, where the layer-2 controller is
// still stamped). So a creature Yahenni's controller stole and then
// sacrificed does NOT grow Yahenni: at the moment it died, its
// controller was not an opponent. That is the printed answer and the
// reason a steal-and-sac deck plays the card for the sacrifice outlet,
// not the counters.
//
// "Sacrifice another creature" excludes Yahenni by name, the
// convention every legendary sac outlet in the catalog uses
// (b03NotNamed): in a singleton format the same name is the same
// creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fdba89eb-1cf5-46e6-9d09-1adb9bc40fcd",
		Name:            "Yahenni, Undying Partisan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && dead.Controller != source.Controller
			}, "Yahenni, Undying Partisan — put a +1/+1 counter on Yahenni",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "Sacrifice another creature: Yahenni gains indestructible until end of turn.",
			Cost:  SacrificeN(1, "another creature", Creature(), b03NotNamed("Yahenni, Undying Partisan")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Yahenni, Undying Partisan — indestructible until end of turn",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
