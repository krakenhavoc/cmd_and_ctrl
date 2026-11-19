package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ordruun Mentor — Creature — Minotaur Soldier {2}{R/W}, 3/2 (Ravnica:
// Clue Edition, Ravnica Allegiance):
//
//	"Mentor (Whenever this creature attacks, put a +1/+1 counter on target
//	 attacking creature with lesser power.)
//	 Whenever you attack a player, target creature that's attacking that
//	 player gains first strike until end of turn."
//
// Mentor is the shared Mentor() trigger. The second ability is once per
// player attacked in a declaration (Echoing Assault's shape, CR 508.3e),
// targeting a creature attacking THAT player, re-checked as it resolves
// (CR 608.2b); the first strike lasts until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "56c8b610-0753-45ce-b1cc-6a5e58538286",
		Name:         "Ordruun Mentor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Mentor("Ordruun Mentor — mentor"),
			OncePerBatchPerPlayer(Targeting(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b16YouAttackedAPlayer(ev, source, g)
			}, "Ordruun Mentor — target creature that's attacking that player gains first strike until end of turn", ordruunMentorFirstStrike),
				AttackingThatPlayer(TargetCreature("target creature that's attacking that player")))),
		},
	})
}

func ordruunMentorFirstStrike(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   t.ID,
		Keywords: []string{"first strike"},
		Label:    "Ordruun Mentor — first strike until end of turn",
	}.Apply(ctx)
}
