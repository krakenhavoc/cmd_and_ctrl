package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plargg, Dean of Chaos // Augusta, Dean of Order — a modal double-faced card
// (Strixhaven: School of Mages, #5502).
//
//	Plargg, Dean of Chaos   {1}{R}   2/2
//	Legendary Creature — Orc Shaman
//	  {T}, Discard a card: Draw a card.
//	  {4}{R}, {T}: Reveal cards from the top of your library until you reveal
//	  a nonlegendary, nonland card with mana value 3 or less. You may cast
//	  that card without paying its mana cost. Put all revealed cards not cast
//	  this way on the bottom of your library in a random order.
//
//	Augusta, Dean of Order  {2}{W}   1/3
//	Legendary Creature — Human Cleric
//	  Other tapped creatures you control get +1/+0.
//	  Other untapped creatures you control get +0/+1.
//	  Whenever you attack, untap each creature you control, then tap any
//	  number of creatures you control.
//
// Plargg's discard is a cost chosen as the ability is activated (CR 602.2b),
// so the draw happens only for a discard that was really paid. The second
// ability is OldB43RevealUntilAndOfferFreeCast: a reveal that leaves the card
// in the library while the free cast is offered, and bottoms what was not cast.
// A token in the library is revealed but never the hit (CR 108.2).
//
// Augusta's two statics read the tapped bit through the layered view (Saryth's
// shape), so a creature gets +1/+0 only while tapped and +0/+1 only while
// untapped, and the change follows it as it taps and untaps; she herself is
// never affected. "Whenever you attack" is one trigger per declaration of
// attackers (OncePerBatch, Banon's shape): it untaps every creature you control,
// attackers included, and then offers a pick of any number of your creatures to
// tap (including none), re-checked as it resolves; the picks may be creatures
// that were not attacking.
//
// No simplification.
func init() {
	const oracle = "a0e85762-e433-4a23-bbee-69daeebb5898"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Plargg, Dean of Chaos",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Discard a card: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, Discards: 1},
				Cost:    effects.Plus(effects.TapCost(), effects.DiscardACard()),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "{4}{R}, {T}: Reveal cards from the top of your library until you reveal a nonlegendary, nonland card with mana value 3 or less. You may cast that card without paying its mana cost. Put all revealed cards not cast this way on the bottom of your library in a random order.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{4}{R}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.OldB43RevealUntilAndOfferFreeCast(effects.NewContext(g, item), "Plargg, Dean of Chaos", func(c game.Card) bool {
						return !c.IsLand() && !c.IsLegendary() && c.ManaValue() <= 3
					})
				},
			},
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Augusta, Dean of Order",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.Labelled(effects.B16Anthem(effects.SarythOtherCreatureYouControl(true), 1, 0), "Other tapped creatures you control get +1/+0."),
			effects.Labelled(effects.B16Anthem(effects.SarythOtherCreatureYouControl(false), 0, 1), "Other untapped creatures you control get +0/+1."),
		},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Augusta, Dean of Order — untap each creature you control, then tap any number of creatures you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var mine []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() {
							mine = append(mine, c.InstanceID)
						}
					}
					if err := effects.UntapEach(ctx, mine); err != nil {
						return err
					}
					if len(mine) == 0 {
						return nil
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  item.Controller,
						Source:   item.SourceCardID,
						Question: "Augusta, Dean of Order — tap any number of creatures you control",
						Cards:    mine,
						Min:      0,
						Max:      0,
						LiveZone: game.ZoneBattlefield,
						ThenKey:  augustaTapPick,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				})),
		},
	})
}

// augustaTapPick taps the picked creatures that are still untapped creatures
// its controller controls. An on-disk key.
var augustaTapPick = effects.AskedPick("choose-cards/old-b043-augusta-tap", func(ctx *effects.Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != ctx.Item.Controller {
			continue
		}
		if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
})
