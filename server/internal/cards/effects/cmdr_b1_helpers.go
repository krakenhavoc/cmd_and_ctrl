package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b1_helpers.go — shared pieces of the Foundations Commander
// angel deck cards (batches 1 and 2, #5368).

// cmdrB1NontokenCreaturesYouControl selects Always Watching's
// subjects: nontoken creatures under the source's controller.
func cmdrB1NontokenCreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && !target.IsToken() && target.Controller == source.Controller
}

// cmdrB1Labeled sets a static's printed clause as its label.
func cmdrB1Labeled(s game.StaticAbility, label string) game.StaticAbility {
	s.Label = label
	return s
}

// cmdrB1CluesYouControl selects the Clues the source's controller
// controls: Clue is an artifact subtype, so a Clue that is not a token
// counts (CR 205.3g).
func cmdrB1CluesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.HasSubtype("Clue")
}

// cmdrB1Investigate is "investigate" as one Do step: create a Clue
// token (CR 701.16a).
func cmdrB1Investigate() Applier {
	return CreateToken{Template: ClueToken(), N: 1}
}

// cmdrB1GainThatMuchDamage is "you gain that much life" for a
// "whenever this creature is dealt damage" trigger settled with
// ThatMuchDamage: the item's event Amount is the total dealt.
func cmdrB1GainThatMuchDamage(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
		return nil
	}
	return GainLife{Amount: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
}

// cmdrB1CreatureAttacksYouOrYourPlaneswalker is Search the Premises'
// condition: one creature was declared as an attacker at the source's
// controller or at a planeswalker they control. An attack at a battle
// is not (CR 508.3e).
func cmdrB1CreatureAttacksYouOrYourPlaneswalker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.AttackingTarget == uuid.Nil {
		return false
	}
	switch g.ClassifyAttackTargetForEffect(c.AttackingTarget) {
	case game.AttackTargetPlayer, game.AttackTargetPlaneswalker:
		return g.DefendingPlayerForAttackForEffect(c.AttackingTarget) == source.Controller
	}
	return false
}

// cmdrB1IsYourFirstThreeTurns reports that it is `player`'s own turn
// and one of the first three they have begun.
func cmdrB1IsYourFirstThreeTurns(g *game.Game, player uuid.UUID) bool {
	mine := g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) &&
		g.Seats[g.Turn.ActiveSeat].ID == player
	return mine && g.TurnsBegunFor(player) <= 3
}

// cmdrB1GraspExiledOne is one exile settled, then the next pick, for
// Grasp of Fate. An on-disk key, registered in init (the body names
// itself again).
var cmdrB1GraspExiledOne game.TailRef

func init() {
	cmdrB1GraspExiledOne = CardTail("then/grasp-of-fate-exiled-one", func(ctx *Context, o game.TailOutcome) error {
		return cmdrB1GraspExileNext(ctx, o.Tail.Carry)
	})
}

// cmdrB1GraspExileNext exiles the first of `picks` until Grasp of Fate
// leaves the battlefield (CR 610.3), the rest riding the exile's
// continuation as its Carry.
func cmdrB1GraspExileNext(ctx *Context, picks []uuid.UUID) error {
	if len(picks) == 0 {
		return nil
	}
	id, rest := picks[0], picks[1:]
	t := ctx.ItemTail(cmdrB1GraspExiledOne)
	t.Carry = append([]uuid.UUID(nil), rest...)
	return ExileUntil{
		Target:     id,
		ThisLeaves: true,
		Label:      "Grasp of Fate — returns when Grasp of Fate leaves the battlefield",
		ThenTail:   t,
	}.Apply(ctx)
}

// cmdrB1GraspExileEach is Grasp of Fate's enters effect: every pick
// still legal as it resolves (CR 608.2b), one per opponent.
func cmdrB1GraspExileEach(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var picks []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			picks = append(picks, t.ID)
		}
	}
	return cmdrB1GraspExileNext(ctx, picks)
}
