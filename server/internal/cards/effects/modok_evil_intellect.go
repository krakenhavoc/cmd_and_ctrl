package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// M.O.D.O.K., Evil Intellect — Legendary Artifact Creature — Villain
// {3}{B}{B}, 3/5 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever you draw your second card each turn, target opponent
//	 sacrifices a nontoken creature of their choice."
//
// The trigger reads the turn's draw tally (Roxxon Brutes's condition).
// The target is the opponent; the creature is their own choice, so it
// is a sacrifice prompt over their nontoken creatures and hexproof is
// irrelevant to it. An opponent with only tokens sacrifices nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "59f75a19-95cd-4ac8-94f1-27c1f915c34a",
		Name:            "M.O.D.O.K., Evil Intellect",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b41DrewYourSecondCardThisTurn(ev, source, g)
				}, "M.O.D.O.K. — target opponent sacrifices a nontoken creature of their choice", modokEdict),
				TargetPlayer("target opponent", Opponent()),
			),
		},
	})
}

func modokEdict(g *game.Game, item *game.StackItem) error {
	for _, t := range NewContext(g, item).LegalTargets() {
		if t.Kind == game.TargetPlayer {
			g.PlayerSacrificesForEffect(item.SourceCardID, t.ID,
				sacrificeSpec("a nontoken creature", Creature(), Not(IsTokenPredicate())),
				"M.O.D.O.K. — sacrifice a nontoken creature")
			return nil
		}
	}
	return nil
}
