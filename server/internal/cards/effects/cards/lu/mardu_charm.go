package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mardu Charm — Instant {R}{W}{B}:
//
//	"Choose one —
//	 • Mardu Charm deals 4 damage to target creature.
//	 • Create two 1/1 white Warrior creature tokens. They gain first
//	   strike until end of turn.
//	 • Target opponent reveals their hand. You choose a noncreature,
//	   nonland card from it. That player discards that card."
//
// The second bullet gives first strike to the two tokens it made and to
// nothing else: the grant is pinned to each token once it exists, so a
// token-creation replacement that changes the count still grants it to
// every token actually created. The third bullet is the revealed-hand
// pick (ADR 0116) filtered to noncreature, nonland cards; a hand with
// none is revealed and nothing is discarded (CR 609.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "147eabab-9d93-4cf4-811b-0efa3b84c5b5",
		Name:         "Mardu Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Mardu Charm deals 4 damage to target creature.",
				effects.TargetCreature("target creature"),
				effects.DealFixedDamageToModesTarget(4)), effects.ForTargets(effects.DamageToTarget(0, 4))),
			effects.ModeDoing("Create two 1/1 white Warrior creature tokens. They gain first strike until end of turn.", nil,
				marduCharmWarriors),
			effects.ModeDoing("Target opponent reveals their hand. You choose a noncreature, nonland card from it. That player discards that card.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				effects.ModeTargetRevealsYouChooseDiscard(effects.And(effects.Noncreature(), effects.Nonland()), "noncreature, nonland card")),
		),
	})
}

// marduCharmWarriors is the second bullet: two Warriors, then first
// strike until end of turn for the ones that were made.
func marduCharmWarriors(item *game.StackItem, ctx *effects.Context, _ int) error {
	return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("1/1 white Warrior")), Count: 2}},
	}, effects.WithWord(effects.NewContext(ctx.Game, item).ItemTail(effects.ThenCreatedGainKeywordUntilEOT), "first strike"))
}
