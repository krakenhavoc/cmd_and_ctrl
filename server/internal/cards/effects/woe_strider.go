package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Woe Strider — Creature — Horror {2}{B}, 3/2:
//
//	"When this creature enters, create a 0/1 white Goat creature
//	 token.
//	 Sacrifice another creature: Scry 1.
//	 Escape—{3}{B}{B}, Exile four other cards from your graveyard.
//	 (You may cast this card from your graveyard for its escape
//	 cost.)
//	 This creature escapes with two +1/+1 counters on it."
//
// A free sac outlet (Viscera Seer's shape) stapled to a body that
// recurs itself: hard-cast it for the Goat and the scry engine, and
// when it dies it is back in the graveyard, four cards away from
// returning as a 5/4. EscapeWithCounters (Voracious Typhon's
// constructor) carries the cost, the exile clause and the counters
// together so the card can't ship as a vanilla 4/3 that forgets its
// own bonus.
//
// Sandbox simplification, declared: "ANOTHER creature" is enforced by
// NAME (b03NotNamed), the same gap Warren Soultrader and Bartolomé del
// Presidio carry — a sacrifice clause built once at init has no
// instance to exclude and can only exclude the source by what it's
// called. Woe Strider is not legendary, so unlike those two a SECOND
// copy on the battlefield (not just a token copy) is also excluded,
// which is weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:      "3adbd963-e85d-4569-963a-4472594f06f9",
		Name:          "Woe Strider",
		Completeness:  CompletenessCaveats,
		Caveats:       []string{"A second copy or token copy of Woe Strider can't be sacrificed to its own ability."},
		CastableZones: []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{
			EscapeWithCounters("{3}{B}{B}", 4, 2),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Woe Strider — create a 0/1 white Goat",
				Do(CreateToken{Template: TokenCard("0/1 white Goat"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "Sacrifice another creature: Scry 1.",
			Cost: game.AbilityCost{
				SacrificeOther: sacrificeSpec("another creature", Creature(), b03NotNamed("Woe Strider")),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Scry{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
