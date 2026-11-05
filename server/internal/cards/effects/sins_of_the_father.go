package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sins of the Father — Enchantment — Saga {3}{B}{R} (Star Trek
// Commander, #4974):
//
//	"I — Each opponent sacrifices a nontoken creature of their choice.
//	 II — For each opponent, you create a 2/1 red Klingon Warrior
//	      creature token with haste.
//	 III — Exile target player's graveyard. Each opponent loses life
//	       equal to the number of creature cards exiled this way."
//
// Chapter I is In the Pale Moonlight's edict: each opponent picks, and
// one with no nontoken creature sacrifices nothing. Chapter II counts the
// opponents still in the game as it resolves. Chapter III targets any
// player, you included, as it goes on the stack; the whole graveyard is
// exiled through the CR 614 window and only the creature cards that
// really reached exile are counted, then each opponent loses that much.
// The Saga is sacrificed after chapter III has resolved (CR 714.4).
//
// No simplification.
func init() {
	const name = "Sins of the Father"
	Register(Spec{
		OracleID:     "bce974a2-d9ad-4ef5-bc2e-7aa2833a39f5",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerEdict(ChapterTrigger(1, SagaChapterLabel(name, 1, "each opponent sacrifices a nontoken creature"),
				func(g *game.Game, item *game.StackItem) error {
					return EachPlayerSacrifices{ExceptController: true, Match: b04NontokenCreature, Label: "a nontoken creature"}.Apply(NewContext(g, item))
				})),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "a 2/1 red Klingon Warrior with haste for each opponent"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return CreateToken{Template: TokenCard("2/1 red Klingon Warrior with haste"), N: len(ctx.Opponents())}.Apply(ctx)
				}),
			ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "exile target player's graveyard; each opponent loses life equal to the creature cards exiled"),
				TargetPlayer("target player"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					p := ctx.PlayerByID(t.ID)
					if p == nil || p.Graveyard == nil {
						return nil
					}
					ids := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
					for _, c := range p.Graveyard.Cards {
						ids = append(ids, c.InstanceID)
					}
					return exileCardsThen(ctx, ids, sinsOfTheFatherDrain)
				}),
		},
	})
}

// sinsOfTheFatherDrain is chapter III's second sentence, told the cards
// that reached exile. An on-disk key.
var sinsOfTheFatherDrain = GraveyardExileTail("sins-of-the-father/creature-cards-drain",
	func(ctx *Context, exiled []uuid.UUID, wasCreature map[uuid.UUID]bool) error {
		n := 0
		for _, id := range exiled {
			if wasCreature[id] {
				n++
			}
		}
		if n == 0 {
			return nil
		}
		return EachOpponentLosesLife{N: n}.Apply(ctx)
	})
