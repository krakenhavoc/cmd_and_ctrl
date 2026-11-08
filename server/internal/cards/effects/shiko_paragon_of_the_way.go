package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shiko, Paragon of the Way — Legendary Creature — Spirit Dragon
// {2}{U}{R}{W}, 4/5 (Tarkir: Dragonstorm, #4463):
//
//	"Flying, vigilance
//	 When Shiko enters, exile target nonland card with mana value 3 or
//	 less from your graveyard. Copy it, then you may cast the copy
//	 without paying its mana cost. (A copy of a permanent spell becomes a
//	 token.)"
//
// Flying and vigilance are printed keywords. The enters trigger is
// Renegade Bull's exile-copy-and-cast (socB12ExileTargetsAndCopyThem, then
// CopyAndCast, ADR 0066's card-copy amendment) with one mandatory target:
// only a card that really reached exile is copied (CR 707.12), the copy is
// cast during the resolution (CR 608.2g) without paying its mana cost, with
// X at 0 (CR 107.3b), a copy not cast ceases to exist, and a permanent copy
// resolves as a token (CR 111.13). The original stays in exile. With the
// target gone by resolution nothing happens (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d3bdd3f2-9e38-436a-93a3-f87d9f4d4f86",
		Name:            "Shiko, Paragon of the Way",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Shiko, Paragon of the Way — exile target nonland card with mana value 3 or less from your graveyard; you may cast a copy of it free",
					func(g *game.Game, item *game.StackItem) error {
						return socB12ExileTargetsAndCopyThem(g, item, shikoCopyExiled)
					}),
				TargetCardInGraveyard("target nonland card with mana value 3 or less from your graveyard",
					YouOwn(), Nonland(), ManaValueLE(3))),
		},
	})
}

// shikoCopyExiled copies what reached exile and offers a free cast of the
// copy. An on-disk key: never renamed, never reused.
var shikoCopyExiled = CardTail("shiko-paragon-of-the-way/copy-exiled", func(ctx *Context, o game.TailOutcome) error {
	return CopyAndCast{Cards: o.IDs, Free: true}.Apply(NewContext(ctx.Game, ctx.Item))
})
