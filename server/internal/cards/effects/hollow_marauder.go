package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hollow Marauder — Creature — Specter Rogue {6}{B}, 4/2 (Outlaws of Thunder Junction):
//
//	"This spell costs {1} less to cast for each creature card in your
//	 graveyard.
//	 Flying
//	 When this creature enters, any number of target opponents each discard a
//	 card. For each of those opponents who didn't discard a card with mana
//	 value 4 or greater, draw a card."
//
// The cost reduction is a self modifier counting the creature cards in your
// graveyard as the spell is cast (CR 601.2f; generic mana only). The enters
// trigger targets any number of opponents (none is legal), who each discard a
// card of their choice in one prompt run; the draws wait for the last discard
// (a registered run continuation, so a table waiting on a prompt is a restore
// point). One card is drawn for each targeted opponent whose discard did not
// include a card with mana value 4 or greater: an opponent with an empty hand
// discarded nothing, so they count. A card with {X} in its cost has X as 0
// in the graveyard (CR 202.3e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1a4075f3-6c18-4cbe-99e0-f403dd78a7b0",
		Name:            "Hollow Marauder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(func(q game.CostQuery) int {
				if q.Game == nil {
					return 0
				}
				p := q.Game.PlayerByIDForEffect(q.Controller)
				if p == nil || p.Graveyard == nil {
					return 0
				}
				n := 0
				for _, c := range p.Graveyard.Cards {
					if c.IsCreature() {
						n++
					}
				}
				return n
			}, "This spell costs {1} less to cast for each creature card in your graveyard."),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Hollow Marauder — any number of target opponents each discard a card; draw for each who didn't discard a card with mana value 4 or greater",
				hollowMarauderEnters),
				TargetPlayer("any number of target opponents", Opponent()).WithCount(0, 0)),
		},
	})
}

func hollowMarauderEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	carry := []uuid.UUID{item.Controller}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			carry = append(carry, t.ID)
		}
	}
	if len(carry) == 1 {
		return nil
	}
	return g.PlayersDiscardThenKeyedForEffect(carry[1:], game.DiscardPrompt{
		Source:   item.SourceCardID,
		N:        1,
		Question: "Hollow Marauder — discard a card",
	}, hollowMarauderDraws, carry)
}

// hollowMarauderDraws runs once the last discard has landed: the controller
// (the first ID on the carry) draws a card for each targeted opponent who did
// not discard a card with mana value 4 or greater. An on-disk key.
var hollowMarauderDraws = RunThen("prompt-run/hollow-marauder-draws", func(ctx *Context, r game.RunDone) error {
	if len(r.Carry) < 2 {
		return nil
	}
	draws := 0
	for _, opp := range r.Carry[1:] {
		bigDiscard := false
		for _, s := range r.Landed {
			if s.Seat != opp {
				continue
			}
			for _, id := range s.Cards {
				if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.ManaValue() >= 4 {
					bigDiscard = true
				}
			}
		}
		if !bigDiscard {
			draws++
		}
	}
	if draws == 0 {
		return nil
	}
	return DrawCards{Player: r.Carry[0], N: draws}.Apply(ctx)
})
