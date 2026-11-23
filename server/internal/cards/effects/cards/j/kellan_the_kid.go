package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kellan, the Kid — Legendary Creature — Human Faerie Rogue {G}{W}{U}, 3/3
// (Outlaws of Thunder Junction, #4468):
//
//	"Flying, lifelink
//	 Whenever you cast a spell from anywhere other than your hand, you may
//	 cast a permanent spell with equal or lesser mana value from your hand
//	 without paying its mana cost. If you don't, you may put a land card
//	 from your hand onto the battlefield."
//
// The cast is made during the trigger's resolution (CR 608.2g) through the
// free hand-cast offer (free_hand_cast.go): a pass-closed {0} permission over
// the permanent cards in the hand whose spell has mana value at most that of
// the spell that was cast (an X spell counts its announced X, CR 202.3e).
// Declining the cast, or having nothing to cast, is "if you don't": the land
// pick then runs (a "may", so declining it is a real answer). Having cast
// something ends the ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "afacc200-16ff-40f3-9d43-268438131d24",
		Name:            "Kellan, the Kid",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B25CastFromNotHand(ev, source.Controller)
			}, "Kellan, the Kid — you may cast a permanent spell with equal or lesser mana value from your hand without paying its mana cost; if you don't, you may put a land card from your hand onto the battlefield",
				kellanTheKidTrigger),
		},
	})
}

// kellanTheKidTrigger offers the free cast over the hand's permanent cards
// no more expensive than the spell that triggered it.
func kellanTheKidTrigger(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	you, source := item.Controller, item.SourceCardID
	mv := 0
	if spell, ok := g.LookupCardForEffect(ctx.Trigger().Event.CardID); ok {
		x := 0
		if it := g.StackItemForEffect(ctx.Trigger().Event.CardID); it != nil {
			x = it.XValue
		}
		mv = spell.ManaValueWithX(x)
	}
	perm := effects.FreeCastSpell{Filter: game.PermissionFilter{NonLandPermanentOnly: true}}.WithManaValueAtMost(mv).
		On(g.FreeCastFromHandPermission(you, source, "Kellan, the Kid"))
	candidates := effects.FreeCastCandidates(g, &perm, effects.AllHandCardIDs(g, you))
	if len(candidates) == 0 {
		return kellanPutALand(ctx)
	}
	if !g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      you,
		Source:       source,
		Reason:       "Kellan, the Kid",
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordHandCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Don't cast a permanent spell",
		Then:         kellanAfterTheCast,
		Carry:        candidates,
	}) {
		g.RunResolutionCastThenForEffect(kellanAfterTheCast, you, source, nil)
	}
	return nil
}

// kellanAfterTheCast is "if you don't, you may put a land card from your
// hand onto the battlefield": it runs once the cast prompt is answered, told
// the cards that were cast as Picked. An on-disk key: never renamed, never
// reused.
var kellanAfterTheCast = effects.AskedPick("choose-cards/kellan-the-kid-if-you-dont", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) > 0 {
		return nil
	}
	return kellanPutALand(ctx)
})

// kellanPutALand is "you may put a land card from your hand onto the
// battlefield".
func kellanPutALand(ctx *effects.Context) error {
	return effects.PutFromHandOntoBattlefield{
		Player:   ctx.Controller(),
		Match:    effects.Land(),
		Optional: true,
	}.Apply(ctx)
}
