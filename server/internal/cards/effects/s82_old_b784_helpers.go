package effects

import "github.com/google/uuid"

// S82OldB784GainLifePerTargetOpponentsPermanent is "You gain <per> life for each
// <match> target opponent controls" (Renewing Dawn, Starlight): the target
// opponent's matching permanents are counted once, as the spell resolves, and
// the caster gains that much. A target gone by resolution does nothing (CR
// 608.2b).
func S82OldB784GainLifePerTargetOpponentsPermanent(ctx *Context, per int, match CardPredicate) error {
	victim := TargetedPlayer(ctx)
	if victim == uuid.Nil {
		return nil
	}
	n := len(PermanentsControlledByMatching(ctx.Game, victim, match))
	return GainLife{Amount: per * n}.Apply(ctx)
}

// S82OldB784AnswerSacrificeAnIslandOrThis is Thing from the Deep's tax answered:
// the controller sacrifices an Island of their choice, or the Leviathan.
var S82OldB784AnswerSacrificeAnIslandOrThis = sacrificeOneOrThisAnswer("option-pick/s82-old-b784-sacrifice-this-unless-you-sacrifice-an-island", "an Island", Subtype("Island"))
