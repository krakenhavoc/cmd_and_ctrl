package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Behold the Sinister Six! — Sorcery {6}{B}:
//
//	"Return up to six target creature cards with different names from
//	 your graveyard to the battlefield."
//
// Eerie Ultimatum's shape with a cap and a creature filter, and — the
// reason it is here — a card that TARGETS, so the set rule is the
// whole of what it needed (#1559). "With different names" is
// EachDifferentName: the picker greys a second card of a name already
// picked, the announce gate refuses the pair (CR 601.2c) and the bot is
// never offered it.
//
// Resolution (CR 608.2b): a card exiled in response is skipped and the
// rest still return, under their owner's control — the caster's, since
// "from your graveyard" — in announce order, each through the ordinary
// reanimation path.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "6fa27f8a-bade-460f-853a-abc1c05c946a",
		Name:         "Behold the Sinister Six!",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("up to six target creature cards with different names from your graveyard",
			Creature(), YouOwn()).WithCount(0, 6).EachDifferent(EachDifferentName()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToBattlefield(ctx)
		},
	})
}
