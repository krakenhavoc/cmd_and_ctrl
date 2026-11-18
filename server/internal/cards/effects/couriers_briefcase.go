package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Courier's Briefcase — Artifact — Treasure {1}{G} (Streets of New Capenna):
//
//	"When this artifact enters, create a 1/1 green and white Citizen
//	 creature token.
//	 {T}, Sacrifice this artifact: Add one mana of any color.
//	 {W}{U}{B}{R}{G}, {T}, Sacrifice this artifact: Draw three cards."
//
// A Treasure card (the subtype is printed, so Treasure payoffs see it) with
// the Treasure mana ability, Glittering Stockpile's shape. The draw ability
// is an ordinary activated ability whose cost sacrifices the source, so it
// resolves even though the Briefcase is gone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "00cd3bf9-0dd6-4ec0-a841-1839a71604ab",
		Name:         "Courier's Briefcase",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}, Sacrifice this artifact: Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Courier's Briefcase — create a 1/1 green and white Citizen creature token",
				Do(CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}{U}{B}{R}{G}, {T}, Sacrifice this artifact: Draw three cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 3},
			Cost:    Plus(ManaCost("{W}{U}{B}{R}{G}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{N: 3}.Apply(NewContext(g, item))
			},
		}},
	})
}
