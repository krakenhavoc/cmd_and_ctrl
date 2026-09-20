package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greater Good — Enchantment {2}{G}{G} (issue #1117):
//
//	"Sacrifice a creature: Draw cards equal to the sacrificed
//	 creature's power, then discard three cards."
//
// A card-advantage engine for a deck that wants its creatures dead
// anyway: convert whatever is about to die into a fresh grip, minus
// the three worst cards in it. Greater Good is an enchantment, not a
// creature, so the cost never needs an "another" exclusion — nothing
// about the ability can eat itself.
//
// The sacrificed creature is read back off the event log at
// resolution, since the stack item does not carry the cost it paid
// (b17PermanentSacrificedToPay, the same read Jarad, Golgari Lich
// Lord uses). "Discard three cards" is the player's own choice
// (CR 701.8a), not a random discard, so it goes through the prompted
// discard path and self-clamps to whatever the hand holds after the
// draw.
//
// Sandbox simplification, declared — the same one Jarad already
// carries: the sacrificed creature's power is its printed power plus
// its +1/+1 and -1/-1 counters as they were when it left. A bonus
// from another permanent's static ability (an anthem) is not in it,
// because the trigger has no last-known characteristic to read for a
// creature that left as a COST rather than dying to an effect.
func init() {
	Register(Spec{
		OracleID:     "dc0593c2-ccb4-4648-a592-c5bcd121dc72",
		Name:         "Greater Good",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The sacrificed creature's power counts its +1/+1 and -1/-1 counters but not a bonus from another permanent, such as an anthem.",
		},
		Activated: []ActivatedAbility{{
			Label: "Sacrifice a creature: Draw cards equal to the sacrificed creature's power, then discard three cards.",
			Cost:  SacrificeACreature(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				fed, ok := b17PermanentSacrificedToPay(g, item)
				if !ok {
					return nil
				}
				power := b17LastKnownPowerOffBattlefield(g, fed)
				ctx := NewContext(g, item)
				if power > 0 {
					if err := (DrawCards{Player: item.Controller, N: power}).Apply(ctx); err != nil {
						return err
					}
				}
				g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player:   item.Controller,
					Source:   item.SourceCardID,
					N:        3,
					Question: "Greater Good — discard three cards",
				})
				return nil
			},
		}},
	})
}
