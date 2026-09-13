package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oran-Rief, the Vastwood — Land (EDHREC rank 732):
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 {T}: Put a +1/+1 counter on each green creature that entered this
//	 turn."
//
// The green token deck's free anthem: drop creatures, tap the land,
// everything new grows. "Entered this turn" is read off the event log
// (b06EnteredThisTurn) — an ETB more recent than the current turn's
// upkeep — because the engine keeps no such flag and summoning
// sickness is not the same question. "Green" reads the post-layer
// colours, and "each green creature" is every player's, as printed.
//
// The counters go through AddCounter, so Doubling Season and Hardened
// Scales apply.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e88027a6-24cc-4a8b-86db-734f26149ea8",
		Name:         "Oran-Rief, the Vastwood",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{T}: Put a +1/+1 counter on each green creature that entered this turn.",
			Cost:  TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, c := range g.BattlefieldCardsForEffect() {
					if !c.IsCreature() || !c.HasColor("G") || !b06EnteredThisTurn(g, c.InstanceID) {
						continue
					}
					if err := (AddCounter{Target: c.InstanceID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
