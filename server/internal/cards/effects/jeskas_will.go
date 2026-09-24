package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jeska's Will — Sorcery {2}{R}:
//
//	"Choose one. If you control a commander as you cast this spell,
//	 you may choose both instead.
//	 • Add {R} for each card in target opponent's hand.
//	 • Exile the top three cards of your library. You may play them
//	   this turn."
//
// The batch-01 triage filed this under "mana pipeline", because a
// SPELL that adds mana had no path to a pool. `AddMana` (the same
// primitive Dark Ritual took in the batch's first pass) is that path,
// and the impulse exile the second bullet wants is the one Laelia and
// Bonehoard Dracosaur already use.
//
// Bullet one counts the hand at RESOLUTION, not at announce, so a
// discard in response shrinks the ritual. The target is an opponent
// and it is a real target: hexproof stops it, and an opponent who
// has left takes the whole bullet with them (CR 608.2b), which is why
// the body re-reads the clause through ctx.ModeTargets rather than
// item.Targets.
//
// Bullet two is "play", not "cast" (CR 601.1a), so a land among the
// three can be played as the turn's land drop. The permission is
// per-instance and expires with the turn; cards left in exile stay
// there.
//
// DECLARED SIMPLIFICATION, weaker than printed: the commander clause
// is not offered. "If you control a commander as you cast this spell,
// you may choose both instead" is a mode COUNT that depends on the
// board at announce, and `game.ModeSpec` bounds the count with plain
// Min/Max integers read at Register time. Offering both
// unconditionally would be stronger than printed (#259), so the card
// asks for one bullet, always — the same seam Akroma's Will records.
//
// Re-audited for #1565: still true on current develop.
// `game.CatalogModeSpec` takes only an oracle ID (no game or player
// state), so there is nowhere for "as you cast this" to be evaluated.
// The seam is the same conditional-mode-count gap both cards share;
// no new engine work has closed it.
func init() {
	Register(Spec{
		OracleID:     "0fd114c4-092b-4e28-b0dc-ef529f3bc73e",
		Name:         "Jeska's Will",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You pick one of the two bullets. Picking both while you control a commander isn't implemented.",
		},
		Modes: ChooseOne(
			ModeDoing("Add {R} for each card in target opponent's hand.",
				TargetPlayer("target opponent", Opponent()),
				jeskasWillRitual),
			ModeDoing("Exile the top three cards of your library. You may play them this turn.", nil,
				jeskasWillImpulse),
		),
	})
}

// jeskasWillRitual is the first bullet: one red mana per card in the
// chosen opponent's hand, counted as the spell resolves. A target
// that is no longer legal adds nothing rather than erroring.
func jeskasWillRitual(item *game.StackItem, ctx *Context, occurrence int) error {
	refs := ctx.ModeTargets(occurrence)
	if len(refs) == 0 {
		return nil
	}
	n := b14HandSize(ctx.Game, refs[0].ID)
	if n <= 0 {
		return nil
	}
	return AddMana{Player: item.Controller, Produced: strings.Repeat("{R}", n)}.Apply(ctx)
}

// jeskasWillImpulse is the second bullet: three cards off the top,
// exiled and playable for the rest of the turn.
func jeskasWillImpulse(item *game.StackItem, ctx *Context, _ int) error {
	_, err := b12ImpulseExileForTurn(ctx.Game, item, 3)
	return err
}
