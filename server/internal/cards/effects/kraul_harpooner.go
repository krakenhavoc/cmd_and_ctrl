package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "332b6ed7-90de-46c0-85ad-1ad708b2c1e1",
		Name:            "Kraul Harpooner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Kraul Harpooner — it gets +X/+0 until end of turn, where X is the number of creature cards in your graveyard, then you may have it fight up to one target creature you don't control with flying",
					Do(
						Instruction(kraulHarpoonerPump),
						Instruction(kraulHarpoonerAskFight),
					)),
				TargetCreature("up to one target creature you don't control with flying", OpponentControls(), HasKeyword("flying")).WithCount(0, 1)),
		},
	})
}

func kraulHarpoonerPump(ctx *Context) error {
	if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	x := creatureCardsInGraveyard(ctx.Game, ctx.Controller())
	if x == 0 {
		return nil
	}
	return BoostUntilEOT{
		Target: ctx.Item.SourceCardID,
		Power:  x,
		Label:  "Kraul Harpooner — +X/+0 until end of turn",
	}.Apply(ctx)
}

func kraulHarpoonerAskFight(ctx *Context) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok || !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return MayChoice{
		Question: "Kraul Harpooner — have it fight the targeted creature?",
		Then:     kraulHarpoonerFight,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}

var kraulHarpoonerFight = MayYes("confirm/kraul-harpooner-fight", func(ctx *Context, r game.ConfirmAnswer) error {
	return b10Fight(ctx, ctx.Item.SourceCardID, carryAt(r, 0))
})
