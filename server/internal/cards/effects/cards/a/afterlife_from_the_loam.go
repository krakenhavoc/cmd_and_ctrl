package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Afterlife from the Loam — Sorcery {5}{B}{B}{B} (Tarkir: Dragonstorm
// Commander):
//
//	"Delve (Each card you exile from your graveyard while casting this spell
//	 pays for {1}.)
//	 For each player, choose up to one target creature card in that player's
//	 graveyard. Put those cards onto the battlefield under your control.
//	 They're Zombies in addition to their other types."
//
// Delve is `Delve: true`. "For each player" is one "up to one" clause for
// your own graveyard and one per possible opponent bound to the n-th seat
// after yours (Windgrace's Judgment's shape, with the owner of the graveyard
// as the binding), so a graveyard may be skipped and an eliminated
// opponent's offers nothing. Every target still legal returns together as one
// entry (Fire Lord Sozin's door) under your control, with the Zombie subtype
// added by the putting effect itself (EntersAs.AddSubtypes, ADR 0210), so
// "enters" triggers see a Zombie.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "151c867f-cc77-42de-9ae5-5fd2233ccdab",
		Name:         "Afterlife from the Loam",
		Completeness: effects.CompletenessFull,
		Delve:        true,
		Targets:      effects.CmdrB13PerPlayerGraveyardClauses("creature", effects.Creature()),
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			return effects.ReturnFromGraveyardTogether{
				Targets:    ids,
				Controller: ctx.Controller(),
				EntersAs: &game.EntersAs{
					AddSubtypes: ctx.TextSubtypes("Zombie"),
					Label:       "Afterlife from the Loam — Zombies in addition to their other types",
				},
			}.Apply(ctx)
		})),
	})
}
