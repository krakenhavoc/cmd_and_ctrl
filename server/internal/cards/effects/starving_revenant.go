package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starving Revenant — Creature — Spirit Horror {2}{B}{B}, 4/4:
//
//	"When this creature enters, surveil 2. Then for each card you put
//	 on top of your library, you draw a card and you lose 3 life.
//	 Descend 8 — Whenever you draw a card, if there are eight or more
//	 permanent cards in your graveyard, target opponent loses 1 life
//	 and you gain 1 life."
//
// # The surveil
//
// "For each card you put on top" is a count of the cards the surveil
// kept, and Surveil.Then is told no such count. The count is derived
// from two facts the engine does hand over: how many cards the player
// looked at (SurveilThenForEffect returns it, after any CR 614
// keyword-action replacement has rewritten it) and how many left the
// library. Every card that does not go back on top goes to the
// graveyard (CR 701.25a), and nothing else moves a card out of this
// library while the prompt is open, so kept = looked - the fall in
// library size. The draw and the life loss ride Then, as Scry's
// warning says they must.
//
// Each kept card is its own pair: draw one, then lose 3. A draw from an
// empty library is still a draw attempt (CR 704.5b).
//
// # Descend 8
//
// A trigger on each of your draws, with an intervening if (CR 603.4)
// that counts permanent cards (CR 110.4) in your graveyard: asked as
// the draw happens, and again as the trigger resolves. Not a keyword
// to the engine, only a condition read off the board.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ca969eb-3d79-4d1f-8d9d-7b8204ad166a",
		Name:         "Starving Revenant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Starving Revenant — surveil 2, then draw and lose 3 life for each card kept on top",
				starvingRevenantSurveil),
			{
				Watches: []game.EventKind{game.EventDrawCard},
				Key:     "Starving Revenant — target opponent loses 1 life and you gain 1 life",
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor == source.Controller && permanentCardsInGraveyard(g, source.Controller) >= 8
				},
				Targets: TargetPlayer("target opponent", Opponent()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if permanentCardsInGraveyard(g, item.Controller) < 8 {
						return nil
					}
					return drainTargetOpponentOne(g, item)
				},
			},
		},
	})
}

// starvingRevenantSurveil is the enters trigger: surveil 2, then a
// draw and 3 life lost per card put back on top.
func starvingRevenantSurveil(g *game.Game, item *game.StackItem) error {
	me := item.Controller
	p := g.PlayerByIDForEffect(me)
	if p == nil {
		return nil
	}
	before := len(p.Library.Cards)
	var looked int
	looked = g.SurveilThenForEffect(me, item.SourceCardID, 2, func(g *game.Game) error {
		q := g.PlayerByIDForEffect(me)
		if q == nil {
			return nil
		}
		kept := looked - (before - len(q.Library.Cards))
		for i := 0; i < kept; i++ {
			if err := g.DrawNForEffect(me, 1); err != nil {
				return err
			}
			if err := g.ChangePlayerLifeForEffect(item.SourceCardID, me, -3); err != nil {
				return err
			}
		}
		return nil
	})
	return nil
}
