package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Squee, Dubious Monarch — Legendary Creature — Goblin Noble {2}{R}, 2/2
// (Dominaria United):
//
//	"Haste
//	 Whenever Squee attacks, create a 1/1 red Goblin creature token that's
//	 tapped and attacking.
//	 You may cast this card from your graveyard by paying {3}{R} and exiling
//	 four other cards from your graveyard rather than paying its mana cost."
//
// The attack trigger is Brimaz's: the Goblin is made tapped and attacking
// whatever its controller picks (CR 508.4), was never declared as an
// attacker, and so triggers nothing that watches declarations. The
// graveyard cast is an ALTERNATIVE cost (CR 118.9), escape's shape with its
// own key: it is not escape (CR 702.138), so nothing that reads "escaped"
// sees it and no counters or exile clause ride on it. The four other cards
// are exiled as a cost, so a graveyard with fewer than four others can't
// pay it; the card itself is never one of them. Cast from the graveyard
// for its printed mana cost is impossible: the offer is the only way in.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "124a5dd7-c4f8-479c-9eb0-669aca01d045",
		Name:            "Squee, Dubious Monarch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		CastableZones:   []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{{
			Key:                "squee-dubious-monarch-graveyard",
			Label:              "Cast from your graveyard—{3}{R}, Exile four other cards from your graveyard",
			ManaCost:           "{3}{R}",
			FromZone:           game.ZoneGraveyard,
			ExileFromGraveyard: CardsInYourGraveyard(4, "Exile four other cards from your graveyard"),
			PayLabel:           "four other cards from your graveyard",
		}},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Squee, Dubious Monarch — create a 1/1 red Goblin creature token that's tapped and attacking",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokensAttackingYourChoice{
						Template: RedGoblinToken(),
						N:        1,
						Tapped:   true,
						Label:    "Squee, Dubious Monarch",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
