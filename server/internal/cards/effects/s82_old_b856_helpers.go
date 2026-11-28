package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b856_helpers.go — Fallen Empires, counterspells, token makers and
// enchantments (S82 OLD b856, #5637).

// S82OldB856ThelonsCurseUpkeep is Thelon's Curse's upkeep trigger: "that player
// may choose any number of tapped blue creatures they control and pay {U} for
// each creature chosen this way. If the player does, untap those creatures." The
// player whose upkeep it is chooses, then is offered the total payment.
func S82OldB856ThelonsCurseUpkeep(g *game.Game, item *game.StackItem) error {
	player := TriggeringActor(item)
	if player == uuid.Nil {
		return nil
	}
	var tapped []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && c.Tapped && c.HasColorWord("U") {
			tapped = append(tapped, c.InstanceID)
		}
	}
	if len(tapped) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Thelon's Curse — choose any number of tapped blue creatures you control ({U} each)",
		Cards:    tapped,
		Min:      0,
		Max:      len(tapped),
		ThenKey:  s82OldB856ThelonsCursePick,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// s82OldB856ThelonsCursePick offers the {U}-per-creature payment for the creatures
// chosen. An on-disk key: never renamed, never reused.
var s82OldB856ThelonsCursePick = AskedPick("choose-cards/s82-old-b856-thelons-curse-choose",
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) == 0 {
			return nil
		}
		objs := make([]game.ObjectRef, 0, len(r.Picked))
		for _, id := range r.Picked {
			if c, ok := ctx.Game.LookupCardForEffect(id); ok {
				objs = append(objs, game.ObjectRef{ID: id, Epoch: c.ObjectEpoch})
			}
		}
		if len(objs) == 0 {
			return nil
		}
		cost := strings.Repeat("{U}", len(objs))
		return MayPay{
			Chooser:      r.Chooser,
			Cost:         cost,
			Question:     "Thelon's Curse — pay " + cost + " to untap the chosen creatures?",
			InThisStep:   true,
			Then:         s82OldB856ThelonsCursePaid,
			CarryObjects: objs,
		}.Apply(ctx)
	})

// s82OldB856ThelonsCursePaid untaps the chosen creatures that are still the objects
// that were chosen. An on-disk key: never renamed, never reused.
var s82OldB856ThelonsCursePaid = PayUnlessBranches("pay-unless/s82-old-b856-thelons-curse-untap",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		for _, o := range r.Objects {
			c, ok := ctx.Game.LookupCardForEffect(o.ID)
			if !ok || c.ObjectEpoch != o.Epoch || !OnBattlefield(ctx.Game, o.ID) {
				continue
			}
			if err := (UntapTarget{Target: o.ID}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}, nil)
