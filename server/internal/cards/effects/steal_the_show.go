package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steal the Show — Sorcery {2}{R} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Choose one or both —
//	 • Target player discards any number of cards, then draws that many
//	   cards.
//	 • Steal the Show deals damage equal to the number of instant and
//	   sorcery cards in your graveyard to target creature or
//	   planeswalker."
//
// The bullets run in printed order (CR 608.2c) and the second waits for
// any prompt the first asked, so with both chosen the discard is answered
// and the discarded instants and sorceries are already in the graveyard
// when the damage counts them. The count is taken once, as the damage is
// dealt (CR 608.2h), from the caster's own graveyard (Steal the Show
// itself is still on the stack); a bullet with no legal target cannot be
// chosen (CR 601.2b), and a target gone by resolution takes nothing
// (CR 608.2b).
//
// "Any number" is a real choice by the targeted player, from nothing to
// the whole hand, and the draw is the discard's continuation, for exactly
// the cards that were discarded (CR 701.9a). That continuation is Mishra's
// Command's, keyed (#4264), so a table waiting on the discard is a
// restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de403543-e61a-44fc-85e1-1e621d90ca8b",
		Name:         "Steal the Show",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeDoing("Target player discards any number of cards, then draws that many cards.",
				TargetPlayer("target player"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					n := b14HandSize(ctx.Game, t.ID)
					if n == 0 {
						return nil
					}
					return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   t.ID,
						Source:   item.SourceCardID,
						N:        n,
						UpTo:     true,
						Question: "Steal the Show — discard any number of cards, then draw that many",
					}, mishrasCommandDraw, nil)
				}),
			ModeDoing("Steal the Show deals damage equal to the number of instant and sorcery cards in your graveyard to target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					n := g2CardsInYourGraveyard(MatchInstantOrSorcery)(item, ctx)
					if n <= 0 {
						return nil
					}
					return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
				}),
		),
	})
}
