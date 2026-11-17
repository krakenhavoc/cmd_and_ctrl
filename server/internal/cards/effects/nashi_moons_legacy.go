package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nashi, Moon's Legacy — Legendary Creature — Rat Shaman {B}{G}{U}, 3/4
// (March of the Machine: The Aftermath, #5476):
//
//	"Menace, ward {1}
//	 Whenever Nashi attacks, exile up to one target legendary or Rat card
//	 from your graveyard and copy it. You may cast the copy. (You still pay
//	 its costs. A copy of a permanent spell becomes a token.)"
//
// Kaervek, the Punisher's exile-and-copy at the printed cost (CopyAndCast
// with no Free, CR 707.12), with no follow-up. "Up to one target" is
// declared on the clause: with no legal card the trigger is not put on the
// stack (CR 603.3d). A permanent copy resolves as a token (CR 707.10b).
//
// No simplification.
var nashiCopyExiled = CardTail("nashi-moons-legacy/copy-exiled", func(ctx *Context, o game.TailOutcome) error {
	return CopyAndCast{Cards: o.IDs}.Apply(NewContext(ctx.Game, ctx.Item))
})

func init() {
	Register(Spec{
		OracleID:        "ffa6c14e-3d8e-47a2-b102-d37538d60ef6",
		Name:            "Nashi, Moon's Legacy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "ward"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{1}"), "Nashi, Moon's Legacy — ward {1}"),
			Targeting(
				On(game.EventAttack, ThisAttacked,
					"Nashi, Moon's Legacy — exile up to one target legendary or Rat card from your graveyard and copy it; you may cast the copy",
					matB04ExileTargetsAndCopy(nashiCopyExiled)),
				TargetCardInGraveyard("up to one target legendary or Rat card from your graveyard",
					YouOwn(), Or(Legendary(), HasSubtype("Rat"))).WithCount(0, 1)),
		},
	})
}
