package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Academy Researchers — Creature — Human Wizard {1}{U}{U}, 2/2 (Urza's Saga,
// S82 OLD batch 730, #5624):
//
//	"When this creature enters, you may put an Aura card from your hand onto
//	 the battlefield attached to this creature."
//
// The controller picks one Aura card from their hand, or none (the "may"). The
// host is named by the effect, so the Aura is not asked for one (CR 303.4f);
// only an Aura that could legally enchant the Researchers is offered, and one
// that can't stays in hand (CR 303.4i). With the Researchers gone as the trigger
// resolves, or no such Aura in hand, nothing is asked. The pick is rechecked as
// it is answered: a card that left the hand, or a host that has left, does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df2d2c52-475b-4057-9cf8-7886700f0fae",
		Name:         "Academy Researchers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Academy Researchers — you may put an Aura card from your hand onto the battlefield attached to this creature",
				func(g *game.Game, item *game.StackItem) error {
					host := item.SourceCardID
					if !effects.OnBattlefield(g, host) || effects.SourceIsNewObject(g, item) {
						return nil
					}
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil || p.Eliminated || p.Hand == nil {
						return nil
					}
					var cands []uuid.UUID
					for _, c := range p.Hand.Cards {
						if c.IsAura() && !c.IsToken() && effects.CanBeAttachedTo(g, item.Controller, c, host) {
							cands = append(cands, c.InstanceID)
						}
					}
					if len(cands) == 0 {
						return nil
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Destination: game.ChooseOntoBattlefield,
						Chooser:     item.Controller,
						Source:      host,
						Question:    "Academy Researchers — you may put an Aura card from your hand onto the battlefield attached to this creature",
						Cards:       cands,
						Min:         0,
						Max:         1,
						ThenKey:     academyResearchersPick,
						Carry:       []uuid.UUID{host},
					})
					return nil
				}),
		},
	})
}

// academyResearchersPick puts the picked Aura onto the battlefield attached to
// the Researchers (the one carried ID), if both are still where they were and
// the Aura can still enchant it. An on-disk key.
var academyResearchersPick = game.RegisterChooseCardsThen("choose-cards/academy-researchers-aura", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 || !effects.OnBattlefield(g, r.Carry[0]) {
		return nil
	}
	card, ok := g.LookupCardForEffect(r.Picked[0])
	z := g.FindCardZoneForEffect(r.Picked[0])
	if !ok || z == nil || z.Kind != game.ZoneHand || !effects.CanBeAttachedTo(g, r.Chooser, card, r.Carry[0]) {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(
		[]game.BatchEntry{{CardID: r.Picked[0], From: game.ZoneHand}},
		game.ZoneEntryOptions{Controller: r.Chooser, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: r.Carry[0]}},
		game.KeyedTail{})
})
