package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bolg, Erebor's Reckoning — Legendary Creature — Goblin Soldier
// {4}{B}{R}, 6/6 (The Hobbit Eternal, HOC):
//
//	"Trample
//	 At the beginning of each combat, other Goblins and Orcs you control get
//	 +2/+2 until end of turn. Creatures your opponents control get -1/-1
//	 until end of turn."
//
// One trigger on every player's beginning of combat (each combat, an added
// one included). Both halves are set when it resolves (CR 611.2c): the
// Goblins and Orcs you control then, other than Bolg (OtherThan), judged by
// their effective subtypes, so an Army that amassed "also a Goblin" counts;
// and the creatures your opponents control then. A creature that enters later
// in the turn gets neither.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "eab096fd-a7e2-4b96-99ec-1039d3a552f3",
		Name:     "Bolg, Erebor's Reckoning",
		// ADR 0126 §6: each combat's -1/-1 on your opponents' creatures removes the x/1s among them.
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfEachCombat("Bolg — other Goblins and Orcs you control get +2/+2 and creatures your opponents control get -1/-1 until end of turn",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					if err := (effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(ctx.Item.SourceCardID), effects.Or(effects.HasSubtype("Goblin"), effects.HasSubtype("Orc"))),
						Power: 2, Toughness: 2,
						Label: "Bolg, Erebor's Reckoning — other Goblins and Orcs get +2/+2",
					}).Apply(ctx); err != nil {
						return err
					}
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.OpponentControls()),
						Power: -1, Toughness: -1,
						Label: "Bolg, Erebor's Reckoning — opposing creatures get -1/-1",
					}.Apply(ctx)
				}))),
		},
	})
}
