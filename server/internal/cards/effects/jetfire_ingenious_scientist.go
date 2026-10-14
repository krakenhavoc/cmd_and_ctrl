package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jetfire, Ingenious Scientist // Jetfire, Air Guardian (#3276, ADR 0152) —
// a More Than Meets the Eye card.
//
// Front face, Legendary Artifact Creature — Robot {4}{U}, 3/4:
//
//	"More Than Meets the Eye {3}{U}
//	 Flying
//	 Remove one or more +1/+1 counters from among artifacts you control:
//	 Target player adds that much {C}. This mana can't be spent to cast
//	 nonartifact spells. Convert Jetfire."
//
// Back face, Legendary Artifact — Vehicle, 3/4:
//
//	"Living metal
//	 Flying
//	 {U}{U}{U}: Convert Jetfire, then adapt 3."
//
// The cost is the variable-AND-among counter removal (#3276): the
// activator announces the total (at least one) and the per-artifact
// split as the cost is paid, and the total is read off the paid-cost
// record. The mana goes to the chosen player and carries the "can't be
// spent to cast nonartifact spells" restriction, so it is spent by hand —
// weaker than printed, never stronger. Adapt 3 checks for +1/+1 counters
// as it resolves (CR 701.46a), after the convert.
//
// No simplification.
const jetfireOracleID = "38de7183-81ef-4a79-b231-7195ba89a281"

func init() {
	Register(Spec{
		OracleID:         jetfireOracleID,
		Name:             "Jetfire, Ingenious Scientist",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{MoreThanMeetsTheEye("{3}{U}")},
		Activated: []ActivatedAbility{{
			Label: "Remove one or more +1/+1 counters from among artifacts you control: Target player adds that much {C}. This mana can't be spent to cast nonartifact spells. Convert Jetfire.",
			Cost: game.AbilityCost{RemoveCounters: &game.CounterRemovalCost{
				Counter:  game.CounterPlusOne,
				N:        1,
				From:     TargetPermanent("artifacts you control", Artifact()),
				Among:    true,
				Variable: true,
			}},
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				var player uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						player = t.ID
						break
					}
				}
				// An illegal target means the whole ability does nothing
				// (CR 608.2b): no mana, and no convert.
				if player == uuid.Nil {
					return nil
				}
				if n := ctx.CountersRemoved(); n > 0 {
					if err := (AddMana{
						Player:       player,
						Amount:       ManaN(n, "C"),
						Restrictions: []string{ManaRestrictNotNonartifactSpell},
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return ConvertThis{}.Apply(ctx)
			},
		}},
	})
	Register(Spec{
		OracleID:        jetfireOracleID + "#1",
		Name:            "Jetfire, Air Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{LivingMetal()},
		Activated: []ActivatedAbility{{
			Label: "{U}{U}{U}: Convert Jetfire, then adapt 3.",
			Cost:  ManaCost("{U}{U}{U}"),
			// Adapt 3 puts +1/+1 counters on it.
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := (ConvertThis{}).Apply(NewContext(g, item)); err != nil {
					return err
				}
				return adaptN(g, item, 3)
			},
		}},
	})
}

// adaptN is "Adapt N" (CR 701.46a): if the source has no +1/+1 counters
// on it as the ability resolves, put N on it.
func adaptN(g *game.Game, item *game.StackItem, n int) error {
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || src.Counters[game.CounterPlusOne] > 0 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(NewContext(g, item))
}
