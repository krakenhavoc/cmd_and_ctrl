package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Renegade Bull — Creature — Ox {4}{R}, 0/5 (Secrets of Strixhaven
// Commander, #4834):
//
//	"Trample
//	 Whenever you cast an instant or sorcery spell, this creature gets
//	 +X/+0 until end of turn, where X is that spell's mana value.
//	 Whenever this creature attacks, exile up to one target instant or
//	 sorcery card from your graveyard and copy it. You may cast the copy
//	 without paying its mana cost."
//
// Trample is the printed keyword. The cast trigger reads the spell's mana
// value from the spell as it was cast (X counted, an alternative cost not
// counted, CR 202.3), falling back to last-known information when the
// spell has already left the stack.
//
// The attack trigger is Kaervek, the Punisher's exile-copy-and-cast
// (CopyAndCast, ADR 0066's card-copy amendment): the card must really
// reach exile to be copied, the copy is cast during the resolution
// (CR 608.2g) without paying its mana cost, with X at 0 (CR 107.3b), and a
// copy not cast ceases to exist. The target is "up to one", so the
// trigger can go on the stack with no target and then does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "61747805-0c34-4beb-8cbf-97bac7a998d0",
		Name:            "Renegade Bull",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Renegade Bull — gets +X/+0 until end of turn, X that spell's mana value",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return BoostUntilEOT{
						Target: ctx.Source(),
						Power:  socB4TriggeringSpellManaValue(g, item),
						Label:  "Renegade Bull — +X/+0",
					}.Apply(ctx)
				}),
			Targeting(
				WheneverThisAttacks("Renegade Bull — exile up to one target instant or sorcery card from your graveyard and copy it; you may cast the copy",
					renegadeBullExileAndCopy),
				TargetCardInGraveyard("up to one target instant or sorcery card from your graveyard",
					YouOwn(), Or(Instant(), Sorcery())).WithCount(0, 1)),
		},
	})
}

func renegadeBullExileAndCopy(g *game.Game, item *game.StackItem) error {
	return socB12ExileTargetsAndCopyThem(g, item, renegadeBullCopyExiled)
}

var renegadeBullCopyExiled = CardTail("renegade-bull/copy-exiled", func(ctx *Context, o game.TailOutcome) error {
	return CopyAndCast{Cards: o.IDs, Free: true}.Apply(NewContext(ctx.Game, ctx.Item))
})
