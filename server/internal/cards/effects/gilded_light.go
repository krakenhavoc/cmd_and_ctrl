package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gilded Light — Instant {1}{W} (Scourge):
//
//	"You gain shroud until end of turn. (You can't be the target of
//	 spells or abilities.)
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The first card to give a PLAYER shroud (#5842). It is the
// player-keyword grant (player_keywords.go) with KeywordShroud, until
// end of turn as printed. CR 702.18a makes shroud refuse every spell
// and ability, the caster's own included, which is where it parts from
// hexproof: canPlayerBeTargetedByLocked refuses a shrouded player for
// any source, and Kaya's hexproof waiver does not reach it. Cycling is
// the keyword's own hand ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f7e7b2aa-9fd2-41dd-b11e-3a233530dc83",
		Name:         "Gilded Light",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GainPlayerKeyword{
				Player:   ctx.Controller(),
				Keyword:  KeywordShroud,
				Label:    "Gilded Light — you gain shroud",
				Duration: DurationUntilEndOfTurn(ctx),
			}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
