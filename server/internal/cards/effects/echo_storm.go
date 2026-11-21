package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Echo Storm — Sorcery {3}{U}{U}:
//
//	"When you cast this spell, copy it for each time you've cast your
//	 commander from the command zone this game. You may choose new targets
//	 for the copies.
//	 Create a token that's a copy of target artifact."
//
// Storm's trigger with a different count (CR 702.40a's copy, CR 707.10): the
// number of times the caster has cast a commander from the command zone,
// summed over every commander they have (Player.CommanderCasts, which only a
// cast from the command zone bumps; Commander's Insight's count). It is read
// as the trigger resolves. The trigger is the caster's, names the spell
// without targeting it and copies from last-known information, so the copies
// are made even if the Echo Storm is countered in response (CR 608.2h).
// Each copy is a separate resolution that makes its own token copy of its own
// target artifact, and the caster may choose new targets for each.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92442ca0-e49d-4fab-8d3a-83405c1b1437",
		Name:         "Echo Storm",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Targets:      TargetPermanent("target artifact", Artifact()),
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Echo Storm — copy it for each time you've cast your commander from the command zone", EchoStormCopies),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return CreateTokenCopy{Controller: item.Controller, Copy: t.ID, N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}

// echoStormCopies copies the spell once per commander cast from the command
// zone, counted as the trigger resolves.
func EchoStormCopies(g *game.Game, item *game.StackItem) error {
	n := SocB3CommanderCastsFromCommandZone(g, item.Controller)
	if n <= 0 {
		return nil
	}
	return CopySpell{
		StackID:          item.SourceCardID,
		Controller:       item.Controller,
		Count:            n,
		ChooseNewTargets: true,
		FromLastKnown:    true,
	}.Apply(NewContext(g, item))
}
