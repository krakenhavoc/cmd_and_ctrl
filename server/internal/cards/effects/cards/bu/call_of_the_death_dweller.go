package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Call of the Death-Dweller — Sorcery {2}{B} (Ikoria: Lair of Behemoths, #5512):
//
//	"Return up to two target creature cards with total mana value 3 or less
//	 from your graveyard to the battlefield. Put a deathtouch counter on
//	 either of them. Then put a menace counter on either of them."
//
// The total is judged at announce and again as the spell resolves, when a set
// that no longer fits has no legal target (CR 608.2b). The cards enter
// together; the counters are put on after they have entered, each chosen by the
// controller from the creatures that actually returned, deathtouch first. The
// two counters may go on the same creature. With one creature back both go on
// it without a question, and with none nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97773064-6534-4761-8c73-3ffd713f4569",
		Name:         "Call of the Death-Dweller",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		Targets: effects.TargetCardInGraveyard("up to two target creature cards with total mana value 3 or less from your graveyard",
			effects.YouOwn(), effects.Creature()).WithCount(0, 2).WithTotal(effects.TotalManaValueAtMost(3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnFromGraveyardTogether{
				Targets:  effects.LegalTargetCardIDs(ctx),
				ThenTail: callOfTheDeathDwellerCounters,
			}.Apply(ctx)
		},
	})
}

// callOfTheDeathDwellerCounters is the rest of the spell after the entry: the
// two counters, put on creatures among those that entered. An on-disk key.
var callOfTheDeathDwellerCounters = effects.TogetherTail("call-of-the-death-dweller/counters", func(ctx *effects.Context, entered []uuid.UUID, _ int) error {
	return effects.S82OldB114CounterPick(ctx.Game, ctx.Controller(), ctx.Source(), "Call of the Death-Dweller", entered, game.CounterDeathtouch, game.CounterMenace)
})
