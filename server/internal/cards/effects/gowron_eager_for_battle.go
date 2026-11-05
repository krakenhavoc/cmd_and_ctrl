package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gowron, Eager for Battle — Legendary Creature — Klingon Warrior
// {2}{B}{R}, 3/3 (Star Trek, #4973):
//
//	"Whenever another creature dies, put a +1/+1 counter on Gowron.
//	 Sacrifice another creature: Gowron gains trample until end of turn.
//	 If you sacrificed a Klingon this way, you draw a card and lose 1
//	 life."
//
// The dies trigger sees any other creature, any controller's, going to a
// graveyard from the battlefield (CR 700.4), Yahenni's counter without
// the opponent clause. The activated ability is Yahenni's sacrifice
// outlet with trample: "another creature" is object identity (CR 109.1),
// so Gowron cannot pay it with itself. Whether the sacrificed creature
// was a Klingon is read off it as it last existed on the battlefield
// (CR 608.2h), so a changeling counts (CR 702.73a). A creature
// sacrificed this way also dies, so it triggers the first ability too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "40616029-e1a8-4e7c-bc7b-e381779c57cc",
		Name:         "Gowron, Eager for Battle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, AnotherCreatureDied, "Gowron, Eager for Battle — put a +1/+1 counter on Gowron",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice another creature: Gowron gains trample until end of turn. If you sacrificed a Klingon this way, you draw a card and lose 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    SacrificeAnotherN(1, "another creature", Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"trample"},
					Label:    "Gowron, Eager for Battle — trample until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				info, ok := ctx.SacrificedPermanent()
				if !ok || !(info.Characteristic.AllCreatureTypes || hasFold(info.Characteristic.Subtypes, "Klingon")) {
					return nil
				}
				if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -1)
			},
		}},
	})
}
