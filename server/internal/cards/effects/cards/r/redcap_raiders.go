package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Redcap Raiders — Creature — Goblin Warrior {2}{R}, 3/2 (Throne of
// Eldraine, #5516):
//
//	"Whenever this creature attacks, you may tap an untapped non-Human
//	 creature you control. If you do, this creature gets +1/+1 and gains
//	 trample until end of turn."
//
// Civil Servant's offer (a pick of at most one untapped creature, re-checked
// as it is tapped, since the board moves under a prompt), then the "if you
// do" pump applies only when a creature was really tapped (CR 608.2c). The
// Raiders may tap itself when it has vigilance: "non-Human creature you
// control" is not "another". "Human" is a word of the card's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05c42b13-5125-477c-991f-692c4e22bc53",
		Name:         "Redcap Raiders",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Redcap Raiders — you may tap an untapped non-Human creature you control; if you do, this creature gets +1/+1 and gains trample until end of turn",
				redcapRaidersAttack),
		},
	})
}

func redcapRaidersOffer(g *game.Game, item *game.StackItem) []game.Card {
	word := g.TextReaderForEffect(item).TextSubtype("Human")
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && !c.Tapped && !c.HasSubtype(word) {
			out = append(out, c)
		}
	}
	return out
}

func redcapRaidersAttack(g *game.Game, item *game.StackItem) error {
	offer := redcapRaidersOffer(g, item)
	if len(offer) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(offer))
	for _, c := range offer {
		ids = append(ids, c.InstanceID)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Redcap Raiders — you may tap an untapped non-Human creature you control",
		Cards:    ids,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  redcapRaidersPick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// redcapRaidersPick taps the chosen creature if it is still an untapped
// non-Human creature of the controller's, and if it was tapped, pumps the
// Raiders. Keyed.
var redcapRaidersPick = effects.AskedPick("choose-cards/redcap-raiders-tap", func(ctx *effects.Context, r game.CardsChosen) error {
	word := ctx.TextSubtype("Human")
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != ctx.Controller() || c.HasSubtype(word) {
			continue
		}
		if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
		self := ctx.Item.SourceCardID
		if err := (effects.BoostUntilEOT{Target: self, Power: 1, Toughness: 1, Label: "Redcap Raiders — +1/+1"}).Apply(ctx); err != nil {
			return err
		}
		return effects.GrantKeywordUntilEOT{Target: self, Keywords: []string{"trample"}, Label: "Redcap Raiders — trample"}.Apply(ctx)
	}
	return nil
})
