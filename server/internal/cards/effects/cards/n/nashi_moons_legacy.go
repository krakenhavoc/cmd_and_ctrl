package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
var nashiCopyExiled = effects.CardTail("nashi-moons-legacy/copy-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.CopyAndCast{Cards: o.IDs}.Apply(effects.NewContext(ctx.Game, ctx.Item))
})

func init() {
	effects.Register(effects.Spec{
		OracleID:        "ffa6c14e-3d8e-47a2-b102-d37538d60ef6",
		Name:            "Nashi, Moon's Legacy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace", "ward"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{1}"), "Nashi, Moon's Legacy — ward {1}"),
			effects.Targeting(
				effects.On(game.EventAttack, effects.ThisAttacked,
					"Nashi, Moon's Legacy — exile up to one target legendary or Rat card from your graveyard and copy it; you may cast the copy",
					effects.MatB04ExileTargetsAndCopy(nashiCopyExiled)),
				effects.TargetCardInGraveyard("up to one target legendary or Rat card from your graveyard",
					effects.YouOwn(), effects.Or(effects.Legendary(), effects.HasSubtype("Rat"))).WithCount(0, 1)),
		},
	})
}
