package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b699_helpers.go — Mercadian Masques, "Mana and ramp, Token makers,
// Graveyard recursion" (S82 OLD batch 699, #5619). Append-only.

// S82OldB699RamosAbilities is the two mana abilities of the five Ramos
// artifacts: "{T}: Add <symbol>" and "Sacrifice this artifact: Add <symbol>"
// (CR 605.1a). The sacrifice needs no {T}, so a freshly cast Ramos can be spent
// the turn it arrives.
func S82OldB699RamosAbilities(symbol string) []ManaAbility {
	return []ManaAbility{
		{Cost: ManaAbilityCost{Tap: true}, Produced: symbol, Label: "{T}: Add " + symbol},
		{Cost: ManaAbilityCost{Sacrifice: true}, Produced: symbol, Label: "Sacrifice this artifact: Add " + symbol},
	}
}

// S82OldB699DealtDamageMayCreateTokens is "Whenever this creature is dealt
// damage, you may create that many <template> creature tokens" (Saber Ants):
// the trigger waits for the damage to finish and counts the total this
// creature was dealt at the same time (CR 120.3, 510.2); the "may" is a real
// decline, asked as the trigger resolves.
func S82OldB699DealtDamageMayCreateTokens(name, template string) game.TriggeredAbility {
	label := name + " — create that many " + template + " creature tokens"
	return Optional(game.TriggeredAbility{
		Watches:    []game.EventKind{game.EventDealDamage},
		Key:        label,
		AtBatchEnd: ThatMuchDamage(),
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return B35SelfWasDealtDamage(ev, source)
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Amount = ev.Amount
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			if item.Params.Amount <= 0 {
				return nil
			}
			return CreateToken{Controller: item.Controller, Template: TokenCard(template), N: item.Params.Amount}.Apply(NewContext(g, item))
		},
	}, name+" — create that many tokens?")
}

// s82OldB699MidnightRitualZombies is "for each creature card exiled this way,
// create a 2/2 black Zombie creature token": the creature cards that really
// reached exile, counted before the move.
var s82OldB699MidnightRitualZombies = GraveyardExileTail("s82-old-b699/midnight-ritual-zombies",
	func(ctx *Context, exiled []uuid.UUID, wasCreature map[uuid.UUID]bool) error {
		n := CreatureCardsAmong(exiled, wasCreature)
		if n == 0 {
			return nil
		}
		return CreateToken{Controller: ctx.Controller(), Template: TokenCard("2/2 black Zombie"), N: n}.Apply(ctx)
	})

// S82OldB699MidnightRitual is "Exile X target creature cards from your
// graveyard. For each creature card exiled this way, create a 2/2 black Zombie
// creature token."
func S82OldB699MidnightRitual(_ *game.StackItem, ctx *Context) error {
	return ExileTargetCardsThen(ctx, s82OldB699MidnightRitualZombies)
}

// S82OldB699EnslavedHorrorReturn is "each other player may return a creature
// card from their graveyard to the battlefield": every other player in APNAP
// order who has a creature card in their graveyard is asked, may take none,
// and the chosen cards enter together under their owners' control, each put by
// the player it enters under (CR 110.2a). One run, so nothing moves until the
// last player has answered.
func S82OldB699EnslavedHorrorReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var legs []game.ChooseCardsPrompt
	var players []uuid.UUID
	for _, id := range ApnapLivingSeats(g) {
		if id == ctx.Controller() {
			continue
		}
		cards := GraveyardCardIDs(ctx, id, game.Card.IsCreature)
		if len(cards) == 0 {
			continue
		}
		players = append(players, id)
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    id,
			FromPlayer: id,
			Source:     ctx.Source(),
			Question:   "Enslaved Horror — you may return a creature card from your graveyard to the battlefield",
			Cards:      cards,
			Min:        0,
			Max:        1,
			Zone:       game.ZoneGraveyard,
		})
	}
	if len(legs) == 0 {
		return nil
	}
	_, err := g.ChooseCardsRunThenKeyedForEffect(legs, s82OldB699EnslavedHorrorReturned, ctx.Source(), ctx.RunData(players))
	return err
}

// s82OldB699EnslavedHorrorReturned puts every chosen creature card onto the
// battlefield together once each player has answered. An on-disk key: never
// renamed, never reused.
var s82OldB699EnslavedHorrorReturned = RunThen("prompt-run/s82-old-b699-enslaved-horror-return", func(ctx *Context, r game.RunDone) error {
	picks := r.Picks()
	var chosen []uuid.UUID
	for _, id := range r.Carry {
		chosen = append(chosen, picks.By(id)...)
	}
	if len(chosen) == 0 {
		return nil
	}
	return ctx.Game.PutByEnteringPlayerForEffect(func() error {
		return ReturnFromGraveyardTogether{Targets: chosen}.Apply(ctx)
	})
})
