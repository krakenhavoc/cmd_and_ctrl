package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Danse Macabre — Sorcery {3}{B}{B} (Forgotten Realms Commander, #5681):
//
//	"Each player sacrifices a nontoken creature of their choice. Roll a
//	 d20 and add the toughness of the creature you sacrificed this way.
//	 1—14 | Return a creature card put into a graveyard this way to the
//	 battlefield under your control.
//	 15+ | Return up to two creature cards put into graveyards this way
//	 to the battlefield under your control."
//
// Rise of the Witch-king's edict: every player, the caster included,
// picks their own nontoken creature in APNAP order, a player with none is
// skipped (CR 701.21a), and everything chosen is sacrificed together. The
// roll waits for the sacrifices to land.
//
// The modifier (CR 706.2, RollDice.Modifier) is the toughness of the
// creature the caster sacrificed as it last existed on the battlefield
// (CR 608.2h), counters and anthems included; a caster who sacrificed
// nothing adds nothing.
//
// "Put into a graveyard this way" is each sacrificed card that is a
// creature card in a graveyard as the table is read: a commander its
// owner moved to the command zone, or a card a replacement exiled, is not
// one. The cards are chosen then, not targeted, from any graveyard, and
// the 15+ row's two enter together (CR 603.6a). The 1—14 row returns one
// if there is one; the 15+ row may return none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a9c58cfe-9630-48fa-84ad-50268eb725ed",
		Name:         "Danse Macabre",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.EachPlayerSacrificesThenKeyedForEffect(
				ctx.Source(), uuid.Nil,
				sacrificeSpec("a nontoken creature", Creature(), Not(IsTokenPredicate())),
				"Danse Macabre — sacrifice a nontoken creature",
				danseMacabreRoll, ctx.RunData([]uuid.UUID{ctx.Controller()}))
		},
	})
}

// danseMacabreRoll is the d20 once the sacrifices have landed: the caster
// is the run's carry. An on-disk key: never renamed, never reused.
var danseMacabreRoll = RunThen("prompt-run/danse-macabre-roll", func(ctx *Context, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	toughness := 0
	if mine := r.Sacrifices().By(r.Carry[0]); len(mine) > 0 {
		if ref, ok := ctx.Game.PermanentRefForEffect(mine[0]); ok {
			if info, ok := ctx.Game.PermanentForEffect(ref); ok {
				toughness = info.Toughness
			}
		}
	}
	return RollDice{Sides: 20, N: 1, Modifier: toughness, Then: danseMacabreResult, TailCarry: r.Sacrifices().Cards()}.Apply(ctx)
})

// danseMacabreResult is the results table: the tail's Carry is every
// card sacrificed. An on-disk key: never renamed, never reused.
var danseMacabreResult = CardTail("dice/danse-macabre-result", func(ctx *Context, o game.TailOutcome) error {
	result := DiceResult(o)
	if len(o.Rolls) == 0 || result < 1 {
		return nil
	}
	var candidates []uuid.UUID
	for _, id := range o.Tail.Carry {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if z := ctx.Game.FindCardZoneForEffect(id); !ok || z == nil || z.Kind != game.ZoneGraveyard || !c.IsCreature() {
			continue
		}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return nil
	}
	prompt := game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Danse Macabre — return a creature card put into a graveyard this way to the battlefield under your control",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		ThenKey:  danseMacabrePicked,
		Asked:    game.NewAskedItem(ctx.Item),
	}
	if result >= 15 {
		prompt.Question = "Danse Macabre — return up to two creature cards put into graveyards this way to the battlefield under your control"
		prompt.Min, prompt.Max = 0, 2
	}
	ctx.Game.QueueChooseCardsForEffect(prompt)
	return nil
})

// danseMacabrePicked is the chosen cards entering under the caster's
// control. An on-disk key: never renamed, never reused.
var danseMacabrePicked = askedPick("choose-cards/danse-macabre-return", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{Targets: r.Picked, Controller: ctx.Controller()}.Apply(ctx)
})
