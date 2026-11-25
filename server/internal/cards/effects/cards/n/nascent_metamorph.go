package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nascent Metamorph — Creature — Shapeshifter {1}{U}, 1/1 (Commander 2020, S82
// OLD, #5513):
//
//	"Whenever this creature attacks or blocks, target opponent reveals cards
//	 from the top of their library until they reveal a creature card. This
//	 creature becomes a copy of that card until end of turn. Then that player
//	 puts all cards revealed this way on the bottom of their library in a
//	 random order."
//
// Morkrut Necropod's attacks-or-blocks trigger (a block counts once, however
// many attackers it blocks) over Vesuvan Drifter's copy: the run is revealed
// off the target opponent's library (a token is never the hit, CR 108.2), the
// Metamorph becomes a copy of the creature card for the turn (CR 707.2,
// CR 611.2; the copiable values are read as the effect begins, so the card
// leaving the library does not change it), and every revealed card, the hit
// included, goes to the bottom in a random order. A library with no creature
// card reveals itself whole, copies nothing and is put back in a random order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68f34b1c-e06b-4859-b7c0-62148341e38c",
		Name:         "Nascent Metamorph",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
				},
				"Nascent Metamorph — target opponent reveals cards from the top of their library until they reveal a creature card; this creature becomes a copy of it until end of turn; they put the revealed cards on the bottom in a random order",
				nascentMetamorphResolve),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

func nascentMetamorphResolve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var opp uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			opp = t.ID
		}
	}
	if opp == uuid.Nil {
		return nil
	}
	run, hit := effects.RevealUntil(ctx, opp, func(c game.Card) bool { return c.IsCreature() },
		"Nascent Metamorph — revealed until a creature card")
	if hit != uuid.Nil {
		if err := (effects.BecomeCopy{
			Targets:  []uuid.UUID{ctx.Source()},
			Of:       hit,
			Duration: effects.CopyUntilEndOfTurn,
			Label:    "Nascent Metamorph — becomes a copy of the revealed creature card until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	if len(run) == 0 {
		return nil
	}
	return g.PutOnBottomInRandomOrderForEffect(opp, game.ZoneLibrary, effects.CardsStillInALibrary(g, run))
}
