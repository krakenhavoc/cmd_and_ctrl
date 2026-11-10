package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Never // Return — split card with aftermath, oracle
// bf276236-18a6-4a48-b47d-5227832d26e6:
//
//	Never — Sorcery {1}{B}{B}: "Destroy target creature or planeswalker."
//	Return — Sorcery {3}{B}, aftermath: "Exile target card from a graveyard.
//	         Create a 2/2 black Zombie creature token."
//
// Insult // Injury's registration: Never under the bare oracle ID, Return
// under "#1" (ADR 0103). Aftermath is the engine's (CR 702.127a). Return is
// one spell with one target: if the card has left its graveyard by
// resolution the spell is countered on resolution and makes no Zombie
// (CR 608.2b); otherwise the exile and the token are both done.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bf276236-18a6-4a48-b47d-5227832d26e6",
		Name:         "Never",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve:    cmdrB25DestroyChosenTarget,
	})
	Register(Spec{
		OracleID:     "bf276236-18a6-4a48-b47d-5227832d26e6#1",
		Name:         "Return",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target card from a graveyard"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				return CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
