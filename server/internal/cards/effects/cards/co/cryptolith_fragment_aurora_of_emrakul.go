package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cryptolith Fragment // Aurora of Emrakul — Artifact {3} // Creature — Eldrazi
// Reflection, 1/4 (Innistrad Remastered, #5948):
//
//	Cryptolith Fragment: "This artifact enters tapped.
//	 {T}: Add one mana of any color. Each player loses 1 life.
//	 At the beginning of your upkeep, if each player has 10 or less life,
//	 transform this artifact."
//	Aurora of Emrakul: "Flying, deathtouch
//	 Whenever this creature attacks, each opponent loses 3 life."
//
// A double-faced card, the back face its own catalog row (oracle ID plus "#1").
// The mana ability is one atomic resolution (CR 605.3b): the mana, then every
// player still in the game loses 1 life as its rider, a loss that is not a cost
// and happens at any life total. The upkeep trigger has an intervening "if"
// (CR 603.4) on every player's life, read as it triggers and again as it
// resolves, and turns the artifact over (TransformThis). The back face is a
// creature with two keywords and Debris Beetle's drain on attack.
//
// No simplification.
func init() {
	const oracle = "60c8a1d4-8d0a-4906-bcaf-d3a09d57c452"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Cryptolith Fragment",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color. Each player loses 1 life.",
			Rider:    gapB08EachPlayerLosesALife,
		}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.GapB08EachPlayerHasAtMostLife(10),
				effects.AtYourUpkeep("Cryptolith Fragment — transform this artifact", effects.Do(effects.TransformThis{}))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Aurora of Emrakul",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Aurora of Emrakul — each opponent loses 3 life", effects.Do(effects.EachOpponentLosesLife{N: 3})),
		},
	})
}

// gapB08EachPlayerLosesALife is the mana ability's rider: every player still in
// the game loses 1 life, in seat order, the source being the permanent that made
// the mana.
func gapB08EachPlayerLosesALife(g *game.Game, _, source uuid.UUID) error {
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if err := g.ChangePlayerLifeForEffect(source, p.ID, -1); err != nil {
			return err
		}
	}
	return nil
}
