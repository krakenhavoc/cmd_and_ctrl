package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82AfrB08EquipmentAttachedTo counts the Equipment attached to a permanent.
// It runs inside a layer pass that may hold only the read lock, so it walks
// the battlefield directly (see PumpAttachedPer).
func s82AfrB08EquipmentAttachedTo(g *game.Game, id uuid.UUID) int {
	if g == nil || g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if a.IsAttachedTo(id) && a.HasSubtype("Equipment") {
			n++
		}
	}
	return n
}

// s82AfrB08OfferAuraFromHand asks `player` to pick an Aura card from their
// hand (or none) and puts it onto the battlefield attached to `host`
// (Holy Avenger). The host is named by the effect, so the Aura is not asked
// for one (CR 303.4f); an Aura that can't legally enchant it stays in hand
// (CR 303.4i). With no Aura card in hand, or a host that has left the
// battlefield, nothing is asked.
func s82AfrB08OfferAuraFromHand(g *game.Game, player, source, host uuid.UUID) error {
	if c, ok := g.LookupCardForEffect(host); !ok || !c.IsCreature() {
		return nil
	}
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Eliminated || p.Hand == nil {
		return nil
	}
	var cands []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.IsAura() && !c.IsToken() {
			cands = append(cands, c.InstanceID)
		}
	}
	if len(cands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Destination: game.ChooseOntoBattlefield,
		Chooser:     player,
		Source:      source,
		Question:    "Holy Avenger — you may put an Aura card from your hand onto the battlefield attached to that creature",
		Cards:       cands,
		Min:         0,
		Max:         1,
		ThenKey:     s82AfrB08HolyAvengerPick,
		Carry:       []uuid.UUID{host},
	})
	return nil
}

// s82AfrB08HolyAvengerPick puts the picked Aura onto the battlefield
// attached to the host the prompt carries.
var s82AfrB08HolyAvengerPick = game.RegisterChooseCardsThen("choose-cards/s82-afr-b08-holy-avenger-aura",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) == 0 || len(r.Carry) == 0 {
			return nil
		}
		return g.PutOntoBattlefieldTogetherKeyedForEffect(
			[]game.BatchEntry{{CardID: r.Picked[0], From: game.ZoneHand}},
			game.ZoneEntryOptions{Controller: r.Chooser, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: r.Carry[0]}},
			game.KeyedTail{})
	})
