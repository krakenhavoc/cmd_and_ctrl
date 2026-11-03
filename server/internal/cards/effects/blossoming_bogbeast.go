package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blossoming Bogbeast — Creature — Beast 3/3 for {4}{G} (Secrets of
// Strixhaven Commander, #4834):
//
//	"Whenever this creature attacks, you gain 2 life. Then creatures you
//	 control gain trample and get +X/+X until end of turn, where X is the
//	 amount of life you gained this turn."
//
// One trigger, one resolution. The 2 life goes through GainLife first,
// so a life-gain replacement (a doubler, or "can't gain life") has had
// its say before X is read. X is the turn's tally of life its
// controller gained (b15LifeGainedThisTurn), read once as the effect
// applies (CR 608.2h), so it counts this trigger's own 2 and any life
// gained earlier in the turn. The creatures affected are fixed then too
// (CR 611.2c): a creature that enters later in the turn gets nothing.
func init() {
	Register(Spec{
		OracleID:     "30f3c3be-0fe9-463d-a245-e44701aec7f2",
		Name:         "Blossoming Bogbeast",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Blossoming Bogbeast — gain 2 life, then creatures you control gain trample and get +X/+X", blossomingBogbeastAttack),
		},
	})
}

func blossomingBogbeastAttack(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (GainLife{Player: item.Controller, Amount: 2}).Apply(ctx); err != nil {
		return err
	}
	x := b15LifeGainedThisTurn(g, item.Controller)
	yours := And(Creature(), YouControl())
	if err := (GrantKeywordUntilEOT{
		Match:    yours,
		Keywords: []string{"trample"},
		Label:    "Blossoming Bogbeast — trample",
	}).Apply(ctx); err != nil {
		return err
	}
	return BoostUntilEOT{
		Match:     yours,
		Power:     x,
		Toughness: x,
		Label:     "Blossoming Bogbeast — +X/+X",
	}.Apply(ctx)
}
