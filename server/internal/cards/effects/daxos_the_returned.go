package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daxos the Returned — Legendary Creature — Zombie Soldier {1}{W}{B}, 2/2
// (Commander 2015, #5546):
//
//	"Whenever you cast an enchantment spell, you get an experience counter.
//	 {1}{W}{B}: Create a white and black Spirit enchantment creature token.
//	 It has "This creature's power and toughness are each equal to the
//	 number of experience counters you have.""
//
// The cast trigger is WheneverYouCast over enchantment spells, with Ezuri's
// player counter (CR 122.1). The Spirit is a catalogued token template
// (printedC15DaxosSpiritToken) whose own layer 7a CDA reads its
// controller's experience counters, so it keeps growing with every counter
// and counts for whoever controls it; a player counter being placed
// invalidates the layers (EventPlayerCounterPlaced).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a9e8f60-eb13-4ae8-82df-d4c46da74451",
		Name:         "Daxos the Returned",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Enchantment(), "Daxos the Returned — you get an experience counter", youGetAnExperienceCounter),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}{B}: Create a white and black Spirit enchantment creature token. It has \"This creature's power and toughness are each equal to the number of experience counters you have.\"",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    ManaCost("{1}{W}{B}"),
			Effect:  Do(CreateToken{Template: C15DaxosSpiritToken(), N: 1}),
		}},
	})
}

// printedC15DaxosSpiritToken is Daxos the Returned's white and black Spirit
// enchantment creature token, a */* whose power and toughness are its
// controller's experience counters. Its slug is an on-disk identity: never
// renamed, never reused.
func printedC15DaxosSpiritToken() tokenTemplate {
	const text = "This creature's power and toughness are each equal to the number of experience counters you have."
	return tokenTemplate{
		Slug: "spirit-experience",
		Card: game.Card{
			Name:              "Spirit",
			TypeLine:          "Token Enchantment Creature — Spirit",
			Colors:            []string{"W", "B"},
			VariableToughness: true, // printed */*
		},
		Static: []game.StaticAbility{{
			Label:     text,
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := experienceCounters(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
		Text: text,
	}
}

// C15DaxosSpiritToken is that Spirit's card value for CreateToken.
func C15DaxosSpiritToken() game.Card { return tokenFromCatalog(printedC15DaxosSpiritToken) }
