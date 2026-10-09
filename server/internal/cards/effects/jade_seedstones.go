package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jade Seedstones // Jadeheart Attendant — a transforming artifact
// (#2124, ADR 0137):
//
//	Jade Seedstones — Artifact {3}{G}
//	  "When this artifact enters, distribute three +1/+1 counters among
//	   one, two, or three target creatures you control.
//	   Craft with creature {5}{G}{G}"
//	Jadeheart Attendant — Artifact Creature — Golem, 7/7
//	  "When this creature enters, you gain life equal to the mana value
//	   of the exiled card used to craft it."
//
// The front trigger is Armament Dragon's divided counters. The back
// face reads CR 702.167c's link (Card.CraftedWith) through
// CraftMaterials when the trigger resolves: a material that has left
// exile by then (a commander its owner moved to the command zone) or a
// token, which ceased to exist there, is no exiled card, and the
// Attendant gains nothing for it. A double-faced material's mana value
// is its front face's, as every card's is outside the stack and the
// battlefield.
//
// No simplification.
const jadeSeedstonesOracleID = "fa2c984a-6bc1-4ec8-93d7-87f9f06b841e"

func init() {
	Register(Spec{
		OracleID:     jadeSeedstonesOracleID,
		Name:         "Jade Seedstones",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Jade Seedstones — distribute three +1/+1 counters", jadeSeedstonesCounters),
				TargetCreature("one, two, or three target creatures you control", YouControl()).
					WithCount(1, 3).Dividing(Divide(3))),
		},
		Activated: []ActivatedAbility{
			Craft("Craft with creature {5}{G}{G}", "{5}{G}{G}", CraftWith("creature")),
		},
	})

	Register(Spec{
		OracleID:     jadeSeedstonesOracleID + "#1",
		Name:         "Jadeheart Attendant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Jadeheart Attendant — gain life equal to the crafted card's mana value", jadeheartAttendantGainLife),
		},
	})
}

func jadeSeedstonesCounters(g *game.Game, item *game.StackItem) error {
	return PutDividedCounters(NewContext(g, item), game.CounterPlusOne)
}

// jadeheartAttendantGainLife is "you gain life equal to the mana value
// of the exiled card used to craft it", read as the trigger resolves.
func jadeheartAttendantGainLife(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	total := 0
	for _, c := range CraftMaterials(ctx) {
		total += c.ManaValue()
	}
	if total <= 0 {
		return nil
	}
	return GainLife{Player: item.Controller, Amount: total}.Apply(ctx)
}
