package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b843_helpers.go — Ice Age creatures (S82 OLD batch 843, #5636).

// S82OldB843YouControlAPlains is "you control a Plains" (Dire Wolves), read through the layers.
func S82OldB843YouControlAPlains(target *game.Card, g *game.Game, source *game.Card) bool {
	if target.InstanceID != source.InstanceID {
		return false
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == source.Controller && c.HasSubtypeWord("Plains") {
			return true
		}
	}
	return false
}

// S82OldB843DefenderControlsASnowLand is Arctic Foxes's blocker test: a creature with power 2 or greater whose
// controller, the defending player, controls a snow land. The defending player is the player the attacker
// attacks, and only their creatures can block it.
func S82OldB843DefenderControlsASnowLand(g *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.Effective().Power >= 2 &&
		g.CountMatchingForEffect(c.Controller, []game.PermanentQuery{{Types: []string{"land"}, Supertypes: []string{"snow"}}}) > 0
}

// S82OldB843PestilenceRatsPower is "the number of other Rats on the battlefield".
func S82OldB843PestilenceRatsPower(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
	n := 0
	for _, o := range g.BattlefieldCardsForEffect() {
		if o.InstanceID != source.InstanceID && o.HasSubtypeWord("Rat") {
			n++
		}
	}
	c.Power = n
}

// S82OldB843ShyftUpkeep is "At the beginning of your upkeep, you may have this creature become the color or colors of
// your choice. (This effect lasts indefinitely.)" The question is asked as the trigger resolves (CR 603.5).
func S82OldB843ShyftUpkeep(g *game.Game, item *game.StackItem) error {
	if !OnBattlefield(g, item.SourceCardID) || SourceIsNewObject(g, item) {
		return nil
	}
	return MayChoice{
		Question: "Shyft — become the color or colors of your choice?",
		Then:     s82OldB843ShyftMay,
	}.Apply(NewContext(g, item))
}

var s82OldB843ShyftMay = MayYes("confirm/s82-old-b843-shyft", func(ctx *Context, _ game.ConfirmAnswer) error {
	if !OnBattlefield(ctx.Game, ctx.Source()) || SourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	ChooseColorsTail(game.ColorForBenefit, ctx.Game, ctx.Controller(), ctx.Source(),
		"Shyft — choose one or more colors", ctx.ItemTail(s82OldB843ShyftColors))
	return nil
})

// s82OldB843ShyftColors sets Shyft's colors indefinitely: for as long as it stays this object (CR 611.2a, 400.7).
var s82OldB843ShyftColors = CardTail("s82-old-b843/shyft-colors", func(ctx *Context, o game.TailOutcome) error {
	colors := game.ColorSetLetters(o.Tail.Word)
	self := ctx.Source()
	if len(colors) == 0 || !OnBattlefield(ctx.Game, self) || SourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	return ScopedEffectFor{
		Target:   self,
		Mods:     []game.Mod{game.SetColorsMod(colors...)},
		Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), self),
		Label:    "Shyft — becomes " + game.ColorSetName(o.Tail.Word),
	}.Apply(ctx)
})

// S82OldB843VampireVictims are the creature cards in any graveyard that went there from the battlefield this
// turn after being dealt damage by `vampire` (as the object it is or last was this turn).
func S82OldB843VampireVictims(g *game.Game, vampire uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.IsCreature() && PutIntoGraveyardFromBattlefieldThisTurn(g, c.InstanceID) && DealtDamageByThisTurn(g, vampire, c.InstanceID) {
				out = append(out, c.InstanceID)
			}
		}
	}
	return out
}

// S82OldB843KrovikanVampireHolds is the intervening "if a creature dealt damage by this creature this turn died".
func S82OldB843KrovikanVampireHolds(ctx *Context) bool {
	return len(S82OldB843VampireVictims(ctx.Game, ctx.Source())) > 0
}

// S82OldB843KrovikanVampireEndStep puts each creature card the Vampire damaged and that died this turn onto the
// battlefield under its controller's control, and for each sets up "Sacrifice it when you lose control of this
// creature" (a delayed trigger that lasts until it fires).
func S82OldB843KrovikanVampireEndStep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	vampire := item.SourceCardID
	for _, id := range S82OldB843VampireVictims(g, vampire) {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
			return err
		}
		if !OnBattlefield(g, id) {
			continue
		}
		if err := (DelayedOnEvent{
			Label:     "Krovikan Vampire — sacrifice the creature it returned",
			On:        []game.EventKind{game.EventControlChanged, game.EventLTB},
			Condition: s82OldB843LostTheVampire,
			Cards:     []uuid.UUID{vampire, id},
			Body:      s82OldB843SacrificeReturned,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// s82OldB843LostTheVampire is "when you lose control of this creature": the Vampire changed controller away from
// the delayed trigger's controller, or left the battlefield (so it is no longer under their control).
var s82OldB843LostTheVampire = game.DelayedCondition("s82-old-b843/krovikan-vampire-lost",
	func(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
		if len(dt.Cards) == 0 || ev.CardID != dt.Cards[0] {
			return false
		}
		switch ev.Kind {
		case game.EventControlChanged:
			return ev.Target == dt.Controller && ev.Actor != uuid.Nil
		case game.EventLTB:
			return true
		}
		return false
	})

var s82OldB843SacrificeReturned = game.SimpleDelayedBody("s82-old-b843/krovikan-vampire-sacrifice",
	func(g *game.Game, item *game.StackItem) error {
		if len(item.Targets) < 2 || item.Targets[1].Kind != game.TargetCard || !OnBattlefield(g, item.Targets[1].ID) {
			return nil
		}
		return SacrificePermanent{Target: item.Targets[1].ID}.Apply(NewContext(g, item))
	})
