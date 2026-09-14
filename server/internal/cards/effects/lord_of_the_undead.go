package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7714af0e-41d9-4609-967f-27233b46055f",
		Name:         "Lord of the Undead",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Zombie"}, Others: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, {T}: Return target Zombie card from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{1}{B}"), TapCost()),
			Targets: TargetCardInGraveyard("target Zombie card from your graveyard", YouOwn(), Subtype("Zombie")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b34ReturnChosenGraveyardCardToHand(NewContext(g, item))
			},
		}},
	})
}
