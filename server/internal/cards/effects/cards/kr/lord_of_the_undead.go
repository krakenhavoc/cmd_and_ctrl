package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lord of the Undead — Creature — Zombie {1}{B}{B}, 2/2 (EDHREC rank
// 3594):
//
//	"Other Zombie creatures get +1/+1.
//	 {1}{B}, {T}: Return target Zombie card from your graveyard to
//	 your hand."
//
// The classic Zombie lord. Read the printed card: the anthem has NO
// controller clause — every other Zombie creature at the table gets
// +1/+1, an opponent's included, which is a real and printed
// drawback (Lord of Atlantis's shape). The activated ability is a
// targeted regrowth over the controller's own graveyard: a "Zombie
// card", so a Kindred Zombie sorcery is a legal pick, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7714af0e-41d9-4609-967f-27233b46055f",
		Name:         "Lord of the Undead",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Zombie"}, Others: true}, 1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}: Return target Zombie card from your graveyard to your hand.",
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target Zombie card from your graveyard", effects.YouOwn(), effects.Subtype("Zombie")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B34ReturnChosenGraveyardCardToHand(effects.NewContext(g, item))
			},
		}},
	})
}
