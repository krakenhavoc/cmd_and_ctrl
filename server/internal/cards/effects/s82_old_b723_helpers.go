package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b723_helpers.go — Urza's Saga damage, S82 OLD batch 723 (#5624).

// S82OldB723WildfireDamage is "Wildfire deals 4 damage to each creature", run
// once the lands have left (a keyed run continuation, so a land that is also a
// creature is gone before the damage, CR 608.2c). An on-disk key: never
// renamed, never reused.
var S82OldB723WildfireDamage = RunThen("prompt-run/s82-old-b723-wildfire-damage", func(ctx *Context, _ game.RunDone) error {
	return DamageEachMatching(ctx, Creature(), 4)
})

// S82OldB723TappedAnIsland is "a player taps an Island for mana": a mana
// ability of an Island that tapped it, whoever controls it. The tapper is the
// event's actor.
func S82OldB723TappedAnIsland(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !B20LandTappedForMana(ev, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && IsLandWithSubtype("island")(c)
}

// S82OldB723DamageThatPlayer is "this deals 1 damage to that player", the
// player being the event's actor (Scald's tapper).
func S82OldB723DamageThatPlayer(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		return DealDamage{Source: item.SourceCardID, Target: item.Trigger.Event.Actor, Amount: n}.Apply(NewContext(g, item))
	}
}

// S82OldB723RedSpellDealsDamage is the condition of Sulfuric Vapors: the
// damage's source is a spell (on the stack or resolving, any controller's)
// that is red as it would deal the damage.
func S82OldB723RedSpellDealsDamage(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
	if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageTarget == uuid.Nil || ev.DamageSource == uuid.Nil {
		return false
	}
	if _, ok := g.DamageSourceSpellControllerForEffect(ev.DamageSource); !ok {
		return false
	}
	ch, ok := DamageSourceCharacteristics(ev, g)
	if !ok {
		return false
	}
	for _, c := range ch.Colors {
		if c == "R" {
			return true
		}
	}
	return false
}

// S82OldB723CreaturesWithLeastToughness lists the creatures on the battlefield
// tied for the least current toughness, in battlefield order.
func S82OldB723CreaturesWithLeastToughness(g *game.Game) []uuid.UUID {
	if g.Battlefield == nil {
		return nil
	}
	var out []uuid.UUID
	least := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.IsCreature() {
			continue
		}
		t := c.CurrentToughness()
		switch {
		case len(out) == 0 || t < least:
			least = t
			out = append(out[:0], c.InstanceID)
		case t == least:
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// S82OldB723ScytheDamage is Purging Scythe's upkeep effect: 2 damage to the
// creature with the least toughness, read as the ability resolves (CR 608.2h),
// anyone's. With one creature at the least toughness it is dealt the damage
// with no question asked; with a tie the ability's controller picks one of the
// tied creatures. No creature on the battlefield: nothing happens.
func S82OldB723ScytheDamage(g *game.Game, item *game.StackItem) error {
	tied := S82OldB723CreaturesWithLeastToughness(g)
	switch len(tied) {
	case 0:
		return nil
	case 1:
		return DealDamage{Source: item.SourceCardID, Target: tied[0], Amount: 2}.Apply(NewContext(g, item))
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Purging Scythe — choose the creature with the least toughness to take 2 damage",
		Cards:    tied,
		Min:      1,
		Max:      1,
		ThenKey:  s82OldB723ScythePick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// s82OldB723ScythePick deals the 2 damage to the chosen tied creature, keyed
// (#4253): the asking item rides the prompt as data.
var s82OldB723ScythePick = AskedPick("choose-cards/s82-old-b723-scythe-pick", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 || !OnBattlefield(ctx.Game, r.Picked[0]) {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: r.Picked[0], Amount: 2}.Apply(ctx)
})

// S82OldB723Antagonism is Antagonism's effect at each player's end step: 2
// damage to that (active) player unless one of their opponents was dealt damage
// this turn. Every other player is an opponent.
func S82OldB723Antagonism(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	active := item.Trigger.Event.Actor
	if active == uuid.Nil {
		return nil
	}
	for _, p := range LivePlayers(g) {
		if p.ID != active && g.DamageTakenThisTurn(p.ID) > 0 {
			return nil
		}
	}
	return DealDamage{Source: item.SourceCardID, Target: active, Amount: 2}.Apply(NewContext(g, item))
}
