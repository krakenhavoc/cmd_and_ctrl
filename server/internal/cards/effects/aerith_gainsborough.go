package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aerith Gainsborough — Legendary Creature — Human Cleric {2}{W}, 2/2
// (Final Fantasy, #4462):
//
//	"Lifelink
//	 Whenever you gain life, put a +1/+1 counter on Aerith Gainsborough.
//	 When Aerith Gainsborough dies, put X +1/+1 counters on each
//	 legendary creature you control, where X is the number of +1/+1
//	 counters on Aerith Gainsborough."
//
// The life-gain trigger fires once per life-gain event, and puts the
// counter on Aerith only while she is the same object on the
// battlefield. The death trigger reads X from Aerith as she last existed
// on the battlefield (CR 608.2h, the engine's last-known record of the
// source object), and counts only +1/+1 counters. "Each legendary
// creature you control" is read as it resolves; Aerith herself is gone
// by then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dd9a6f6b-f9b7-4292-bc80-b7da326de032",
		Name:            "Aerith Gainsborough",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Aerith Gainsborough — put a +1/+1 counter on Aerith Gainsborough",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			WhenThisDies("Aerith Gainsborough — put X +1/+1 counters on each legendary creature you control",
				aerithGainsboroughDies),
		},
	})
}

// aerithGainsboroughDies puts Aerith's last-known +1/+1 counter count on
// each legendary creature her controller controls.
func aerithGainsboroughDies(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	n := info.Counters[game.CounterPlusOne]
	if n <= 0 {
		return nil
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || !c.IsCreature() || !c.IsLegendary() {
			continue
		}
		if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: n}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
