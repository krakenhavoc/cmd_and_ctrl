package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moseo, Vein's New Dean — Legendary Creature — Bird Skeleton Warlock {2}{B},
// 2/1 (Secrets of Strixhaven, #4456):
//
//	"Flying
//	 When Moseo enters, create a 1/1 black and green Pest creature token
//	 with "Whenever this token attacks, you gain 1 life."
//	 Infusion — At the beginning of your end step, if you gained life this
//	 turn, return up to one target creature card with mana value X or less
//	 from your graveyard to the battlefield, where X is the amount of life
//	 you gained this turn."
//
// The Pest is the Secrets of Strixhaven one (PestAttackToken, life on
// attack). Infusion is an intervening "if" (CR 603.4): checked as the end
// step begins and again as the trigger resolves. The target is chosen as the
// trigger goes on the stack, among the creature cards in your graveyard whose
// mana value is at most the life you have gained so far (ManaValueAtMostLife
// GainedThisTurn); it is judged again as the trigger resolves, with the life
// gained by then (CR 608.2b), and "up to one" lets you choose none. The
// creature returns untapped under your control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c2e2f21a-0338-44d7-96a5-5a277f0ec419",
		Name:            "Moseo, Vein's New Dean",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Moseo, Vein's New Dean — create a 1/1 Pest that gains you life when it attacks",
				Do(CreateToken{Template: PestAttackToken(), N: 1})),
			Targeting(
				InterveningIf(socB29YouGainedLifeThisTurn,
					On(game.EventBeginEndStep, AllOf(ByYou, socB29YouGainedLifeThisTurnWhen),
						"Moseo, Vein's New Dean — Infusion: return up to one target creature card with mana value X or less from your graveyard to the battlefield",
						moseoInfusion)),
				TargetCardInGraveyard("up to one target creature card with mana value X or less from your graveyard, where X is the life you gained this turn",
					YouOwn(), Creature(), ManaValueAtMostLifeGainedThisTurn()).WithCount(0, 1),
			),
		},
	})
}

// socB29YouGainedLifeThisTurn is the CR 603.4 recheck of Infusion's "if you
// gained life this turn" as the trigger resolves. The life-gained tally only
// rises within a turn, so a trigger that held as the step began still holds.
var socB29YouGainedLifeThisTurn = CondBoard{
	Name: "you gained life this turn",
	Test: func(ctx *Context) bool { return b15LifeGainedThisTurn(ctx.Game, ctx.Controller()) > 0 },
}

// socB29YouGainedLifeThisTurnWhen is the same "if" as the trigger's When,
// read as the end step begins.
func socB29YouGainedLifeThisTurnWhen(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return b15LifeGainedThisTurn(g, source.Controller) > 0
}

// moseoInfusion returns the chosen creature card to the battlefield.
func moseoInfusion(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
}
