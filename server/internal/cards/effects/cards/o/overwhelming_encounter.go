package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Overwhelming Encounter — Sorcery {3}{G}{G} (Commander Legends: Battle for
// Baldur's Gate, S82 GAP #5948):
//
//	"Creatures you control gain vigilance and trample until end of turn. Roll
//	 a d20.
//	 1—9 | Creatures you control get +2/+2 until end of turn.
//	 10—19 | Put two +1/+1 counters on each creature you control.
//	 20 | Put four +1/+1 counters on each creature you control."
//
// The keywords go on first, to the creatures the caster controls as the spell
// resolves (CR 611.2c). The d20 then goes through the one dice door and its tail
// reads the table over the creatures the caster controls at that moment (the
// roll may have waited on a replacement): +2/+2 until end of turn, or counters
// placed as one placement per creature (CR 614, so a doubler applies once).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7573ce9f-fc35-4d20-9b60-c739b8408444",
		Name:         "Overwhelming Encounter",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"vigilance", "trample"},
				Label:    "Overwhelming Encounter — vigilance and trample until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.RollDice{Sides: 20, N: 1, Then: overwhelmingEncounterRolled}.Apply(ctx)
		},
	})
}

// overwhelmingEncounterRolled reads the table. An on-disk key.
var overwhelmingEncounterRolled = effects.DiceTail("dice/overwhelming-encounter-rolled", func(ctx *effects.Context, rolls []int) error {
	switch n := rolls[0]; {
	case n >= 20:
		return effects.B11PutCountersOnEachCreatureYouControl(ctx.Game, ctx.Item, 4)
	case n >= 10:
		return effects.B11PutCountersOnEachCreatureYouControl(ctx.Game, ctx.Item, 2)
	}
	return effects.BoostUntilEOT{
		Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 2,
		Label: "Overwhelming Encounter — +2/+2 until end of turn",
	}.Apply(ctx)
})
