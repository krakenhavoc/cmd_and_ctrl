package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "147eabab-9d93-4cf4-811b-0efa3b84c5b5",
		Name:         "Mardu Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Mardu Charm deals 4 damage to target creature.",
				TargetCreature("target creature"),
				DealFixedDamageToModesTarget(4)),
			ModeDoing("Create two 1/1 white Warrior creature tokens. They gain first strike until end of turn.", nil,
				marduCharmWarriors),
			ModeDoing("Target opponent reveals their hand. You choose a noncreature, nonland card from it. That player discards that card.",
				TargetPlayer("target opponent", Opponent()),
				ModeTargetRevealsYouChooseDiscard(And(Noncreature(), Nonland()), "noncreature, nonland card")),
		),
	})
}

// marduCharmWarriors is the second bullet: two Warriors, then first
// strike until end of turn for the ones that were made.
func marduCharmWarriors(item *game.StackItem, ctx *Context, _ int) error {
	return ctx.Game.CreateTokensThenForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: TokenCard("1/1 white Warrior"), Count: 2}},
	}, func(g *game.Game, created []uuid.UUID) error {
		next := NewContext(g, item)
		for _, id := range created {
			if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: "Mardu Charm"}).Apply(next); err != nil {
				return err
			}
		}
		return nil
	})
}
