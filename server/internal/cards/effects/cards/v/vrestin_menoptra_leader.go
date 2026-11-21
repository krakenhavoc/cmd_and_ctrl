package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vrestin, Menoptra Leader — Legendary Creature — Alien Insect Scout
// {X}{G}{G}{W}, 0/0 (Doctor Who):
//
//	"Flying
//	 Vrestin enters with X +1/+1 counters on it.
//	 When Vrestin enters, create X 1/1 green and white Alien Insect creature tokens
//	 with flying.
//	 Whenever you attack with one or more Insects, put a +1/+1 counter on each of
//	 them."
//
// X is the X the spell was cast with: the counters are the entry clause
// (XCounters, CR 614.1c) and the tokens read it off the permanent as the trigger is
// built (CR 107.3m). The attack trigger fires once per declaration of attackers
// that includes an Insect you control (the word read through the source's text);
// "each of them" is the Insects you control that were declared as attackers and
// are attacking as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "1db6a042-303e-4b76-bb6c-63df846fd6c9",
		Name:                       "Vrestin, Menoptra Leader",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"flying"},
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersWithCastX("Vrestin — create X 1/1 green and white Alien Insect creature tokens with flying",
				func(g *game.Game, item *game.StackItem) error {
					x := item.Params.Amount
					if x <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 green and white Alien Insect with flying"), N: x}.Apply(effects.NewContext(g, item))
				}),
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.HasSubtype(source.TextSubtype("Insect"))
			}, "Vrestin — put a +1/+1 counter on each Insect that attacked", vrestinCounters)),
		},
	})
}

// vrestinCounters puts a +1/+1 counter on each Insect the controller has that was
// declared as an attacker and is attacking now.
func vrestinCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	insect := ctx.TextSubtype("Insect")
	for _, id := range effects.AttackingCreatures(g) {
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.Controller != item.Controller || !c.HasSubtype(insect) || !g.AttackedThisTurn(id) {
			continue
		}
		if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
