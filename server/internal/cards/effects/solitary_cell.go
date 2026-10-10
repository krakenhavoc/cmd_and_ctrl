package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Solitary Cell — Artifact {R}{W} (Reality Fracture):
//
//	"When this artifact enters, exile target nonland permanent an
//	 opponent controls with mana value 3 or less until this artifact
//	 leaves the battlefield.
//	 {1}, {T}, Discard a legendary card: Draw a card."
//
// Oblivion Ring's shape (CR 610.3) with a mana-value cap on the target,
// plus a rummage-style activated ability whose discard is a legendary
// card from hand, chosen at announce.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "66a22b35-3dc0-4a74-8474-e73293403de6",
		Name:         "Solitary Cell",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Targets: TargetPermanent("target nonland permanent an opponent controls with mana value 3 or less",
					And(Nonland(), OpponentControls(), ManaValueLE(3))),
				Key:    solitaryCellExileLabel,
				Effect: exileChosenTargetUntilThisLeaves("Solitary Cell — the exiled card returns when Solitary Cell leaves the battlefield"),
			},
			UntilThisLeavesLegacyReturn("Solitary Cell — return the exiled card", solitaryCellExileLabel),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Discard a legendary card: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost: Plus(ManaCost("{1}"), TapCost(),
				DiscardCardsMatching(1, "a legendary card", func(c game.Card) bool { return c.IsLegendary() })),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.DrawNForEffect(item.Controller, 1)
			},
		}},
	})
}

const solitaryCellExileLabel = "Solitary Cell — exile target nonland permanent an opponent controls with mana value 3 or less"
