package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmc_b8_helpers.go — shared vocabulary for TMNT Commander batch 8
// (#5011).

// tmcB8APlayerCast is "whenever a player casts a <spell> spell": any
// player's cast, the source's controller included, matching `spell`
// (nil for every spell). YouCast and AnOpponentCast are its halves.
func tmcB8APlayerCast(spell CardPredicate) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventCast || ev.Actor == uuid.Nil {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && (spell == nil || spell(g, ev.Actor, game.ThroughTextOf(source, c)))
	}
}

// tmcB8ALandEnteredDuringYourTurn is "whenever a land enters during your
// turn" (Foe-liage): a land entering under ANY player's control, while the
// source's controller is the active player.
func tmcB8ALandEnteredDuringYourTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	_, ok := b39LandEnteredController(ev, g)
	return ok && isActivePlayer(g, source.Controller)
}

// tmcB8TargetPlayerPutsCountersOnACreature is "target player chooses a
// creature they control and puts N +1/+1 counters on it" (Mikey & Mona,
// Mutant Sitters). The choice is the target player's, made at resolution,
// and it does not target. The target player is the one who puts the
// counters (#3942's placer). No creature, no counters; one
// creature needs no prompt, since the choice is forced.
func tmcB8TargetPlayerPutsCountersOnACreature(ctx *Context, player uuid.UUID, n int, question string) error {
	var candidates []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			candidates = append(candidates, c.InstanceID)
		}
	}
	switch len(candidates) {
	case 0:
		return nil
	case 1:
		return nothingIfGone(ctx.Game.AddCounterByForEffect(player, candidates[0], game.CounterPlusOne, n))
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     ctx.Source(),
		Question:   question,
		Cards:      candidates,
		Min:        1,
		Max:        1,
		ThenKey:    tmcB8CountersOnThePickedCreature,
		LiveZone:   game.ZoneBattlefield,
		CarryNums:  []int{n},
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

// tmcB8CountersOnThePickedCreature puts the carried number of +1/+1
// counters on the creature the chooser picked, placed by the chooser.
var tmcB8CountersOnThePickedCreature = game.RegisterChooseCardsThen("choose-cards/tmc-b8-counters-on-the-picked-creature", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) != 1 || len(r.Nums) == 0 {
		return nil
	}
	return nothingIfGone(g.AddCounterByForEffect(r.Chooser, r.Picked[0], game.CounterPlusOne, r.Nums[0]))
})

// tmcB8TargetPlayerReturnsFromGraveyard is "target player returns a
// <match> card from their graveyard to their hand": the target player
// picks one, if they have any.
func tmcB8TargetPlayerReturnsFromGraveyard(ctx *Context, player uuid.UUID, match func(game.Card) bool, question string) error {
	p := ctx.Game.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var cards []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if match(c) {
			cards = append(cards, c.InstanceID)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	queueReturnPickedToHand(ctx, player, question, cards)
	return nil
}
