package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vial Smasher the Fierce — Legendary Creature — Goblin Berserker {1}{B}{R},
// 2/3:
//
//	"Whenever you cast your first spell each turn, choose an opponent at
//	 random. Vial Smasher deals damage equal to that spell's mana value to
//	 that player or a planeswalker that player controls.
//	 Partner"
//
// The trigger asks for the turn's first spell by the cast tally, which is
// bumped before the cast event fires, and needs Vial Smasher on the
// battlefield as the spell is cast (it is the trigger's source). The
// opponent is drawn from the keyed random stream among the opponents still
// in the game, as the ability resolves; then, only if that opponent
// controls a planeswalker, the controller picks the recipient among that
// player and their planeswalkers (the printed ruling: random first, choice
// after, no actions between). The amount is the spell's mana value as it
// last existed on the stack, X included (CR 608.2h).
func init() {
	Register(Spec{
		OracleID:     "1b355139-4891-436c-9079-0ce17ecc2d65",
		Name:         "Vial Smasher the Fierce",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, youCastYourFirstSpellThisTurn,
				"Vial Smasher the Fierce — choose an opponent at random; it deals damage equal to that spell's mana value to that player or a planeswalker they control",
				vialSmasherDamage),
		},
	})
}

// youCastYourFirstSpellThisTurn is "whenever you cast your first spell each
// turn": the tally is bumped before the cast event fires, so the first
// spell makes it one.
func youCastYourFirstSpellThisTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor == source.Controller && g.CastTallyFor(ev.Actor).Total == 1
}

func vialSmasherDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if item.Trigger == nil {
		return nil
	}
	var pool []uuid.UUID
	for _, opp := range ctx.Opponents() {
		if p := g.PlayerByIDForEffect(opp); p != nil && !p.Eliminated {
			pool = append(pool, opp)
		}
	}
	pick := g.ChooseAtRandomForEffect(randomDraw(ctx), pool, 1)
	if len(pick) != 1 {
		return nil
	}
	opp := pick[0]
	name := "an opponent"
	if p := g.PlayerByIDForEffect(opp); p != nil && p.Name != "" {
		name = p.Name
	}
	g.EmitEvent(game.Event{
		Kind:   game.EventOptionChosen,
		Actor:  item.Controller,
		CardID: item.SourceCardID,
		Label:  "Vial Smasher the Fierce: " + name + " (at random)",
	})
	spell := item.Trigger.Event.CardID
	var walkers []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == opp && c.IsPlaneswalker() {
			walkers = append(walkers, c.InstanceID)
		}
	}
	if len(walkers) == 0 {
		return vialSmasherHit(ctx, opp, spell)
	}
	options := []game.ChoiceOption{{Label: name, Player: opp}}
	for _, id := range walkers {
		if c, ok := g.LookupCardForEffect(id); ok {
			options = append(options, game.ChoiceOption{Label: c.Name, Cards: []uuid.UUID{id}})
		}
	}
	return PickOption{
		Question: "Vial Smasher the Fierce — deal damage to " + name + " or a planeswalker they control",
		Options:  options,
		ThenKey:  vialSmasherRecipient,
		Carry:    []uuid.UUID{spell},
	}.Apply(ctx)
}

// vialSmasherRecipient is the controller's choice among the chosen
// opponent and their planeswalkers. An on-disk key.
var vialSmasherRecipient = OptionPickThen("option-pick/vial-smasher-the-fierce-recipient", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 {
		return nil
	}
	target := r.Option.Player
	if target == uuid.Nil && len(r.Option.Cards) == 1 {
		target = r.Option.Cards[0]
	}
	if target == uuid.Nil {
		return nil
	}
	return vialSmasherHit(ctx, target, r.Carry[0])
})

// vialSmasherHit deals damage equal to the cast spell's mana value to a
// player or planeswalker.
func vialSmasherHit(ctx *Context, target, spell uuid.UUID) error {
	mv, ok := ctx.Game.LastKnownSpellManaValueForEffect(spell)
	if !ok || mv <= 0 {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: target, Amount: mv}.Apply(ctx)
}
