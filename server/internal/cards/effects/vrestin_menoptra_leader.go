package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:                   "1db6a042-303e-4b76-bb6c-63df846fd6c9",
		Name:                       "Vrestin, Menoptra Leader",
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"flying"},
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersWithCastX("Vrestin — create X 1/1 green and white Alien Insect creature tokens with flying",
				func(g *game.Game, item *game.StackItem) error {
					x := item.Params.Amount
					if x <= 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 green and white Alien Insect with flying"), N: x}.Apply(NewContext(g, item))
				}),
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) {
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
	ctx := NewContext(g, item)
	insect := ctx.TextSubtype("Insect")
	for _, id := range AttackingCreatures(g) {
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.Controller != item.Controller || !c.HasSubtype(insect) || !g.AttackedThisTurn(id) {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
