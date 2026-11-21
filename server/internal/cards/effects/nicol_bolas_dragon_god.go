package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nicol Bolas, Dragon-God — Legendary Planeswalker — Bolas {U}{B}{B}{B}{R},
// loyalty 4 (Ravnica Remastered):
//
//	"Nicol Bolas has all loyalty abilities of all other planeswalkers on the
//	 battlefield.
//	 +1: You draw a card. Each opponent exiles a card from their hand or a
//	 permanent they control.
//	 −3: Destroy target creature or planeswalker.
//	 −8: Each opponent who doesn't control a legendary creature or
//	 planeswalker loses the game."
//
// The static is a layer-6 borrow of the loyalty rows of every other
// planeswalker (ADR 0155), read as they change, so Bolas's own loyalty is
// what pays for them and CR 606.3's one-loyalty-ability-a-turn still holds.
// The +1 draws, then asks each opponent for one object among the cards in
// their hand and the permanents they control. The −8 judges each opponent as
// it resolves (CR 104.3e).
//
// No simplification.
func init() {
	const clause = "Nicol Bolas has all loyalty abilities of all other planeswalkers on the battlefield."
	Register(Spec{
		OracleID:        "581eedc4-efe4-4ae4-baf2-88071e7e6f6d",
		Name:            "Nicol Bolas, Dragon-God",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Static: []game.StaticAbility{HasActivatedAbilitiesOf(game.AbilityBorrow{
			Label: clause,
			Lenders: BattlefieldLenders(func(_ *game.Game, _ *game.Card, c game.Card) bool {
				return c.IsPlaneswalker()
			}),
			Except: game.BorrowExceptNonLoyalty,
		})},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: You draw a card. Each opponent exiles a card from their hand or a permanent they control.",
				Cost:    LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 1},
				Effect: Do(
					DrawCards{N: 1},
					Instruction(nicolBolasDragonGodExile),
				),
			},
			{
				Label:   "−3: Destroy target creature or planeswalker.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
				Effect:  destroyFirstLegalCardTarget,
			},
			{
				Label: "−8: Each opponent who doesn't control a legendary creature or planeswalker loses the game.",
				Cost:  LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, opp := range ctx.Opponents() {
						if nicolBolasDragonGodControlsLegend(ctx, opp) {
							continue
						}
						if err := (LoseTheGame{Player: opp}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})
}

// nicolBolasDragonGodControlsLegend is "controls a legendary creature or
// planeswalker", read through the layers.
func nicolBolasDragonGodControlsLegend(ctx *Context, player uuid.UUID) bool {
	for _, c := range MatchingBattlefield(ctx, And(Or(Creature(), Planeswalker()), Legendary())) {
		if c.Controller == player {
			return true
		}
	}
	return false
}

// nicolBolasDragonGodExile asks each opponent for one card from their hand
// or one permanent they control and exiles it, keyed. An opponent with
// neither is not asked.
func nicolBolasDragonGodExile(ctx *Context) error {
	g := ctx.Game
	for _, opp := range ctx.Opponents() {
		p := g.PlayerByIDForEffect(opp)
		if p == nil || p.Eliminated {
			continue
		}
		var options []uuid.UUID
		if p.Hand != nil {
			for _, c := range p.Hand.Cards {
				options = append(options, c.InstanceID)
			}
		}
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == opp {
				options = append(options, c.InstanceID)
			}
		}
		if len(options) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  opp,
			Source:   ctx.Source(),
			Question: "Nicol Bolas, Dragon-God — exile a card from your hand or a permanent you control",
			Cards:    options,
			Min:      1,
			Max:      1,
			ThenKey:  nicolBolasDragonGodExiled,
			Asked:    game.NewAskedItem(ctx.Item),
		})
	}
	return nil
}

var nicolBolasDragonGodExiled = askedPick("choose-cards/nicol-bolas-dragon-god-exile", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	return ExileTarget{Target: r.Picked[0]}.Apply(ctx)
})
