package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Death Mutation — Sorcery {6}{B}{G} (Apocalypse, #5614):
//
//	"Destroy target nonblack creature. It can't be regenerated. Create X
//	 1/1 green Saproling creature tokens, where X is that creature's mana
//	 value."
//
// Befoul's destruction that ignores regeneration (CR 701.19c), then the
// caster's Saprolings. X is the creature's mana value as it last existed
// on the battlefield (CR 608.2h), read before it is destroyed, so the
// tokens come whether or not it died (an indestructible creature still
// counts). A creature that turned black in response is an illegal target
// and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f705338-b00b-4a6d-924f-443e7c5670b7",
		Name:         "Death Mutation",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target nonblack creature", NonBlack()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, mv, ok := firstLegalTargetCardManaValue(ctx)
			if !ok {
				return nil
			}
			if err := (DestroyTarget{Target: id, CantBeRegenerated: true}).Apply(ctx); err != nil {
				return err
			}
			if mv <= 0 {
				return nil
			}
			return CreateToken{Template: TokenCard("1/1 green Saproling"), N: mv}.Apply(ctx)
		},
	})
}
