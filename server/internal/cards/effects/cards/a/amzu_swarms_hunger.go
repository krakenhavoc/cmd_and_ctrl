package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amzu, Swarm's Hunger — Legendary Creature — Insect Shaman {3}{B}{G}, 3/3
// (Ravnica: Clue Edition, #5929):
//
//	"Flying, menace
//	 Other Insects you control have menace.
//	 Whenever one or more cards leave your graveyard, you may create a 1/1
//	 black and green Insect creature token, then put a number of +1/+1
//	 counters on it equal to the greatest mana value among those cards. Do
//	 this only once each turn."
//
// The lord half is Hordewing Skaab's TribeFilter. The trigger is Ark of
// Hunger's batch trigger over cards of yours that leave a graveyard
// (whenCardsLeaveYourGraveyard under OncePerBatch), so cards leaving together
// are one occurrence. "Those cards" are read at resolution off that batch
// (bnsB21GreatestManaValueLeft), whichever zone they went to. "You may" is
// the resolution question; "do this only once each turn" is Riveteers
// Ascendancy's record: the turn is marked only when a token is really made,
// so a "no" leaves a later batch free to do it, and a second trigger already
// on the stack does nothing once the first has made a token.
//
// No simplification.
func init() {
	const label = "Amzu, Swarm's Hunger — create a 1/1 black and green Insect token, then put +1/+1 counters on it equal to the greatest mana value among the cards that left"
	const done = label + " (done this turn)"
	effects.Register(effects.Spec{
		OracleID:        "6777863c-e52c-4a7c-b8d2-a16b1438b3d2",
		Name:            "Amzu, Swarm's Hunger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "menace"},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Insect"}, Others: true, YoursOnly: true}, "menace"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.OncePerBatch(effects.OnAny(effects.LeftGraveyardEvents,
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return source != nil && !g.ReplacementAppliedThisTurnForEffect(source, done) &&
						effects.WhenCardsLeaveYourGraveyard(ev, source, lki, g)
				}, label,
				func(g *game.Game, item *game.StackItem) error {
					src, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || g.ReplacementAppliedThisTurnForEffect(&src, done) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					greatest := effects.BnsB21GreatestManaValueLeft(g, item)
					g.MarkReplacementAppliedThisTurnForEffect(&src, done)
					cursor := effects.B25LastEventSeq(g)
					if err := (effects.CreateToken{Template: effects.TokenCard("1/1 black and green Insect"), N: 1}).Apply(ctx); err != nil {
						return err
					}
					for _, token := range effects.B27TokensCreatedByAfter(g, item.Controller, cursor) {
						if greatest > 0 {
							return effects.AddCounter{Target: token, Kind: game.CounterPlusOne, N: greatest}.Apply(ctx)
						}
					}
					return nil
				})),
				"Amzu, Swarm's Hunger — create a 1/1 black and green Insect token with +1/+1 counters?"),
		},
	})
}
