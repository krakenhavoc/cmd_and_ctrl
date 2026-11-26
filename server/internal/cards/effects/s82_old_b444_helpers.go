package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Innistrad batch 444 helpers: Curses, the Auras that enchant a player.

// OldB444AtEnchantedPlayersUpkeep is the trigger condition "at the beginning
// of enchanted player's upkeep": the upkeep is that of the player the Curse
// is attached to (CR 303.4, 702.5).
func OldB444AtEnchantedPlayersUpkeep(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventBeginUpkeep && source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Actor
}

// OldB444EnchantedPlayersUpkeepTrigger is a Curse's upkeep payoff: the player
// is the one whose upkeep it is, read off the trigger's event.
func OldB444EnchantedPlayersUpkeepTrigger(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventBeginUpkeep, OldB444AtEnchantedPlayersUpkeep, label, effect)
}

// OldB444CreatureCardsInYourGraveyard counts the creature cards in the
// graveyard of the Aura's controller ("you" is the Aura's controller, not the
// host's, CR 109.5), for Wreath of Geists.
func OldB444CreatureCardsInYourGraveyard(g *game.Game, source *game.Card) int {
	return B11CreatureCardsInGraveyard(g, source.Controller)
}

// OldB444CreaturesEnchantedPlayerControls is the static scope "creatures
// enchanted player controls" for a Curse.
func OldB444CreaturesEnchantedPlayerControls(target *game.Card, _ *game.Game, source *game.Card) bool {
	return source.AttachedTo.Kind == game.TargetPlayer && target.Controller == source.AttachedTo.ID && target.IsCreature()
}

// OldB444CombatDamageToEnchantedPlayerBy is Curse of Stalked Prey's condition:
// a creature dealt combat damage to the enchanted player. The creature is the
// event's source.
func OldB444CombatDamageToEnchantedPlayerBy(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	if source.AttachedTo.Kind != game.TargetPlayer || source.AttachedTo.ID != ev.Target {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && c.IsCreature()
}

// OldB444ExileTwoFromYourGraveyard is Curse of Oblivion's payoff: the player
// whose upkeep it is exiles two cards from their graveyard, their choice; with
// fewer than two, all of them.
func OldB444ExileTwoFromYourGraveyard(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	player := item.Trigger.Event.Actor
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil || p.Graveyard.Size() == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, p.Graveyard.Size())
	for _, c := range p.Graveyard.Cards {
		ids = append(ids, c.InstanceID)
	}
	n := min(2, len(ids))
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     item.SourceCardID,
		Question:   "Curse of Oblivion — exile two cards from your graveyard",
		Cards:      ids,
		Min:        n,
		Max:        n,
		Zone:       game.ZoneGraveyard,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    OldB444CurseOfOblivionExile,
	})
	return nil
}

// OldB444CurseOfOblivionExile is the answer to the pick: the chosen cards are
// exiled together. An on-disk key: never renamed, never reused.
var OldB444CurseOfOblivionExile = game.RegisterChooseCardsThen("choose-cards/curse-of-oblivion-exile", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(r.Picked, game.KeyedTail{})
})

// OldB444DamageEnchantedPlayerOrAPlaneswalker is Curse of the Pierced Heart's
// payoff: it deals 1 damage to the player whose upkeep it is or a planeswalker
// that player controls, the Aura's controller choosing as it resolves (not a
// target). With no planeswalker to choose, the player takes it.
func OldB444DamageEnchantedPlayerOrAPlaneswalker(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	victim := item.Trigger.Event.Actor
	p := g.PlayerByIDForEffect(victim)
	if p == nil {
		return nil
	}
	var walkers []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == victim && c.IsPlaneswalker() {
			walkers = append(walkers, c.InstanceID)
		}
	}
	if len(walkers) == 0 {
		return DealDamage{Source: item.SourceCardID, Target: victim, Amount: 1}.Apply(ctx)
	}
	options := []game.ChoiceOption{{Label: p.Name, Player: victim}}
	for _, id := range walkers {
		if c, ok := g.LookupCardForEffect(id); ok {
			options = append(options, game.ChoiceOption{Label: c.Name, Cards: []uuid.UUID{id}})
		}
	}
	return PickOption{
		Question: "Curse of the Pierced Heart — deal 1 damage to " + p.Name + " or a planeswalker they control",
		Options:  options,
		ThenKey:  OldB444PiercedHeartRecipient,
	}.Apply(ctx)
}

// OldB444PiercedHeartRecipient is the Aura's controller's choice among the
// enchanted player and their planeswalkers. An on-disk key.
var OldB444PiercedHeartRecipient = OptionPickThen("option-pick/curse-of-the-pierced-heart-recipient", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil {
		return nil
	}
	target := r.Option.Player
	if target == uuid.Nil && len(r.Option.Cards) == 1 {
		target = r.Option.Cards[0]
	}
	if target == uuid.Nil {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: target, Amount: 1}.Apply(ctx)
})
