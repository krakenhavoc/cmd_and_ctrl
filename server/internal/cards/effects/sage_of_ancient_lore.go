package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sage of Ancient Lore // Werewolf of Ancient Hunger — {4}{G} Creature —
// Human Shaman Werewolf */* // Creature — Werewolf */* (#4071, ADR 0132
// amendment 2026-10-16):
//
//	Front: "Sage of Ancient Lore's power and toughness are each equal to
//	        the number of cards in your hand.
//	        When this creature enters, draw a card.
//	        At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Vigilance, trample
//	        Werewolf of Ancient Hunger's power and toughness are each
//	        equal to the total number of cards in all players' hands.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// Each face's P/T is a layer 7a characteristic-defining ability,
// Psychosis Crawler's shape, with DependsOnHandSize so a draw or a
// discard resizes it between recomputes. Turning over swaps which count
// it reads in the same pass. "All players' hands" counts every player
// still in the game.
//
// No simplification.
func init() {
	const oracle = "a46454ad-d912-42f7-b228-492796161d82"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Sage of Ancient Lore",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{handSizeCDA(
			"Power and toughness are each equal to the number of cards in your hand.",
			func(g *game.Game, source *game.Card) int {
				return cardsInHand(g, source.Controller)
			})},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sage of Ancient Lore — draw a card", Do(DrawCards{N: 1})),
			TransformAtUpkeepIfNoSpellsLastTurn("Sage of Ancient Lore"),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Werewolf of Ancient Hunger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Static: []game.StaticAbility{handSizeCDA(
			"Power and toughness are each equal to the total number of cards in all players' hands.",
			func(g *game.Game, _ *game.Card) int {
				n := 0
				for _, p := range g.Seats {
					if p != nil && !p.Eliminated {
						n += cardsInHand(g, p.ID)
					}
				}
				return n
			})},
		Triggered: []game.TriggeredAbility{
			TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Werewolf of Ancient Hunger"),
		},
	})
}

// handSizeCDA is a "power and toughness are each equal to <a count of
// cards in hands>" characteristic-defining ability (CR 604.3, layer 7a).
func handSizeCDA(label string, count func(g *game.Game, source *game.Card) int) game.StaticAbility {
	return game.StaticAbility{
		Label:             label,
		Layer:             game.Layer7PT,
		SubLayer:          game.SubLayer7A_CDA,
		DependsOnHandSize: true,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := count(g, source)
			c.Power = n
			c.Toughness = n
		},
	}
}
