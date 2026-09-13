package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Haywire Mite — Artifact Creature — Insect {1}, 1/1 (EDHREC rank
// 683):
//
//	"When this creature dies, you gain 2 life.
//	 {G}, Sacrifice this creature: Exile target noncreature artifact
//	 or noncreature enchantment."
//
// A one-mana answer to a Rhystic Study or a Sol Ring that also gains
// two life on the way out — the sacrifice IS a death, so the dies
// trigger fires off the activation and resolves above the exile.
// The exile is a CR 602 activated ability with a mana plus
// sacrifice-self cost and a target clause: artifact or enchantment,
// and NOT a creature (so an artifact creature is out of reach, as
// printed).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "749d2994-44e7-40d3-8630-7bebed239e9e",
		Name:         "Haywire Mite",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Haywire Mite — you gain 2 life",
					func(g *game.Game, item *game.StackItem) error {
						return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
					})
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{G}, Sacrifice this creature: Exile target noncreature artifact or noncreature enchantment.",
			Cost:  Plus(ManaCost("{G}"), SacrificeThis()),
			Targets: TargetPermanent("target noncreature artifact or noncreature enchantment",
				Or(Artifact(), Enchantment()), Noncreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return ExileTarget{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
			},
		}},
	})
}
