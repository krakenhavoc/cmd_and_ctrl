package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sita Varma, Masked Racer — Legendary Creature — Human Rogue {G}{U},
// 2/3 (Aetherdrift, #4464):
//
//	"Exhaust — {X}{G}{G}{U}: Put X +1/+1 counters on Sita Varma. Then you
//	 may have the base power and toughness of each other creature you
//	 control become equal to Sita Varma's power until end of turn.
//	 (Activate each exhaust ability only once.)"
//
// The counters come first (CR 608.2c), so "Sita Varma's power" is read after
// them, as the player answers the "may": a Sita that has left by then is read
// as she last existed (CR 608.2h). The number is fixed when the effect begins
// to apply (CR 611.2c): the other creatures you control as it resolves get
// base power and toughness equal to it (layer 7b), under any counters and
// pumps they carry, until end of turn, and a creature that arrives later is
// not changed. Sita herself is not "each other creature".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b9bc41a-a43f-4be7-9452-f61f5ad5b482",
		Name:         "Sita Varma, Masked Racer",
		Completeness: CompletenessFull,
		XMatters:     true,
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {X}{G}{G}{U}: Put X +1/+1 counters on Sita Varma. Then you may have the base power and toughness of each other creature you control become equal to Sita Varma's power until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, SourceCounters: 1},
			Exhaust: true,
			Cost:    ManaCost("{X}{G}{G}{U}"),
			Effect: Do(
				Instruction(func(ctx *Context) error {
					if sourceIsNewObject(ctx.Game, ctx.Item) || ctx.X() <= 0 {
						return nil
					}
					return AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: ctx.X()}.Apply(ctx)
				}),
				Instruction(func(ctx *Context) error {
					return MayChoice{
						Question: "Sita Varma — have the base power and toughness of each other creature you control become equal to her power until end of turn?",
						Then:     sitaVarmaBaseAnswer,
					}.Apply(ctx)
				}),
			),
		}},
	})
}

// sitaVarmaBaseAnswer is the "you may" (keyed, #4291): each other creature the
// controller controls gets base power and toughness equal to Sita's power as
// last known.
var sitaVarmaBaseAnswer = MayYes("confirm/sita-varma-base-pt", func(ctx *Context, _ game.ConfirmAnswer) error {
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	power := info.Power
	source := ctx.Source()
	return ScopedEffectFor{
		Match: func(_ *game.Game, caster uuid.UUID, c game.Card) bool {
			return c.IsCreature() && c.Controller == caster && c.InstanceID != source
		},
		Mods:  []game.Mod{game.SetBasePowerMod(power), game.SetBaseToughnessMod(power)},
		Label: "Sita Varma — base power and toughness " + strconv.Itoa(power) + "/" + strconv.Itoa(power) + " until end of turn",
	}.Apply(ctx)
})
