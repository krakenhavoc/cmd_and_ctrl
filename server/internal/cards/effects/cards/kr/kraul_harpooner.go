package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kraul Harpooner — Creature — Insect Warrior {1}{G}, 3/2 (Ravnica: Clue
// Edition, Ravnica Allegiance):
//
//	"Reach
//	 Undergrowth — When this creature enters, choose up to one target
//	 creature you don't control with flying. This creature gets +X/+0 until
//	 end of turn, where X is the number of creature cards in your
//	 graveyard, then you may have this creature fight that creature."
//
// Undergrowth is an ability word. The target is up to one (zero is legal),
// and the pump happens whether or not a target was chosen, with X counted
// as the trigger resolves. Only then is the controller asked whether to fight
// (a "you may" that belongs to the fight alone). The fight happens only if the
// Harpooner is still the same permanent and the target is still a legal
// creature you don't control with flying (CR 608.2b, CR 701.14b); a target
// that lost flying in response is not fought.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "332b6ed7-90de-46c0-85ad-1ad708b2c1e1",
		Name:            "Kraul Harpooner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Kraul Harpooner — it gets +X/+0 until end of turn, where X is the number of creature cards in your graveyard, then you may have it fight up to one target creature you don't control with flying",
					effects.Do(
						effects.Instruction(kraulHarpoonerPump),
						effects.Instruction(kraulHarpoonerAskFight),
					)),
				effects.TargetCreature("up to one target creature you don't control with flying", effects.OpponentControls(), effects.HasKeyword("flying")).WithCount(0, 1)),
		},
	})
}

func kraulHarpoonerPump(ctx *effects.Context) error {
	if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	x := effects.CreatureCardsInGraveyard(ctx.Game, ctx.Controller())
	if x == 0 {
		return nil
	}
	return effects.BoostUntilEOT{
		Target: ctx.Item.SourceCardID,
		Power:  x,
		Label:  "Kraul Harpooner — +X/+0 until end of turn",
	}.Apply(ctx)
}

func kraulHarpoonerAskFight(ctx *effects.Context) error {
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok || !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return effects.MayChoice{
		Question: "Kraul Harpooner — have it fight the targeted creature?",
		Then:     kraulHarpoonerFight,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}

var kraulHarpoonerFight = effects.MayYes("confirm/kraul-harpooner-fight", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	return effects.B10Fight(ctx, ctx.Item.SourceCardID, effects.CarryAt(r, 0))
})
