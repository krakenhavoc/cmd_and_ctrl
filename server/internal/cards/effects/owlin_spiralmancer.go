package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Owlin Spiralmancer — Creature — Bird Wizard {3}{U}, 3/4 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Flying, vigilance
//	 Whenever you cast your first spell with {X} in its mana cost each
//	 turn, you may copy it. You may choose new targets for the copy. (A
//	 copy of a permanent spell becomes a token.)"
//
// Nev, the Practical Dean's trigger condition (socB9FirstXSpellThisTurn):
// the whole turn's casts are scanned, so an X spell cast before the
// Spiralmancer arrived uses the turn up (the ruling). The "you may" is
// asked as the trigger resolves; a yes is a CR 707.10 copy with the
// new-targets offer (CR 707.10c). The copy keeps the spell's X, modes
// and paid additional costs, and is created, not cast (the rulings). The
// trigger names the spell without targeting it, so a spell countered in
// response is still copied from last-known information (CR 608.2h). A
// copy of a permanent spell resolves into a token (CR 707.10f, the
// engine's copy path).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "36d6d6c0-721c-40c3-b81f-4f027fcf847e",
		Name:            "Owlin Spiralmancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, socB9FirstXSpellThisTurn, "Owlin Spiralmancer — you may copy your first X spell this turn", owlinSpiralmancerCopy),
				"Owlin Spiralmancer — copy that spell? You may choose new targets for the copy."),
		},
	})
}

// owlinSpiralmancerCopy copies the spell that triggered it.
func owlinSpiralmancerCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return CopySpell{StackID: ctx.Trigger().Event.CardID, ChooseNewTargets: true, FromLastKnown: true}.Apply(ctx)
}
