package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nyx Weaver — Enchantment Creature — Spider {1}{B}{G}, 2/3 (EDHREC
// rank 1724):
//
//	"Reach
//	 At the beginning of your upkeep, mill two cards.
//	 {1}{B}{G}, Exile this creature: Return target card from your
//	 graveyard to your hand."
//
// The Golgari self-mill engine that cashes itself in for the best card
// it milled. The exile ability is #1404's cost, ExileThis() paid from
// the BATTLEFIELD: the Weaver leaves at announce and never reaches the
// graveyard, so it cannot return itself (the target was chosen while
// it was still on the battlefield, and it is not a card in a graveyard
// then), and a dies payoff never sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6f8bf968-0571-4582-8c04-9791e44d5df0",
		Name:            "Nyx Weaver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Nyx Weaver — mill two cards", Do(MillCards{N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}{G}, Exile this creature: Return target card from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{1}{B}{G}"), ExileThis()),
			Targets: TargetCardInGraveyard("target card from your graveyard", YouOwn()),
			Effect:  returnFirstLegalGraveyardTargetToHand,
		}},
	})
}
