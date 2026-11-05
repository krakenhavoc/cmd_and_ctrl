package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Big Mother Mouser — Artifact Creature — Robot {4}, 0/0 (Teenage Mutant
// Ninja Turtles Commander, #5011):
//
//	"This creature enters with two +1/+1 counters on it.
//	 Whenever this creature attacks, double the number of +1/+1 counters
//	 on it.
//	 When this creature dies, create a number of 1/1 colorless Robot
//	 artifact creature tokens equal to the number of +1/+1 counters on
//	 this creature."
//
// The entry counters are a replacement on its own entry (CR 614.1c,
// Kalonian Hydra's b10EntersWithCounters), so the 0/0 never meets a state
// check without them. The attack trigger doubles by putting as many
// +1/+1 counters on it again as it has (CR 701.10e), placed through
// AddCounter, so a counter doubler applies. The dies trigger looks back
// (CR 603.10a): the count is the +1/+1 counters it had as it left
// (Chasm Skulker's SourcePermanent read), so -1/-1 counters that killed
// it in the same event do not shrink the number (the card's ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94751968-e006-45d0-970b-0ed45d28e6cc",
		Name:         "Big Mother Mouser",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters("+1/+1", 2, "Big Mother Mouser: enters with two +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Big Mother Mouser — double the +1/+1 counters on it", bigMotherMouserDouble),
			WhenThisDies("Big Mother Mouser — create a 1/1 Robot for each +1/+1 counter it had", bigMotherMouserRobots),
		},
	})
}

// bigMotherMouserDouble doubles the +1/+1 counters on the Mouser itself.
func bigMotherMouserDouble(g *game.Game, item *game.StackItem) error {
	return doublePlusOneCountersOnEach(g, item, func(c game.Card) bool {
		return c.InstanceID == item.SourceCardID
	})
}

// bigMotherMouserRobots makes one Robot per +1/+1 counter the Mouser had
// as it died.
func bigMotherMouserRobots(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok || info.Counters[game.CounterPlusOne] <= 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 colorless Robot artifact"), N: info.Counters[game.CounterPlusOne]}.Apply(ctx)
}
