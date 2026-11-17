package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exile — Instant {2}{W} (Alliances, #5634):
//
//	"Exile target nonwhite attacking creature. You gain life equal to its
//	 toughness."
//
// The target must be attacking and not white both when Exile is cast and
// as it resolves (CR 608.2b). The toughness is read as the creature last
// existed on the battlefield (CR 608.2h), so it is taken just before the
// exile, pumps and all; the life goes to Exile's controller, whoever
// controlled the creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c91d50b-bf58-4d06-b3b7-4b6db867bfce",
		Name:         "Exile",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target nonwhite attacking creature", Not(OfColor("W")), AttackingCreature()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    exileTheAttackerAndGainItsToughness,
	})
}

func exileTheAttackerAndGainItsToughness(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	id := item.Targets[0].ID
	card, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	// The gain does not depend on the exile: the text does not say "if
	// you do", so a replacement that sends the creature elsewhere still
	// leaves the life (Swords to Plowshares' reading). A toughness below
	// zero gains nothing.
	life := max(card.CurrentToughness(), 0)
	if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Player: ctx.Controller(), Amount: life}.Apply(ctx)
}
