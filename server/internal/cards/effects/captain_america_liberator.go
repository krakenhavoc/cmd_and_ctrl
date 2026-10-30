package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain America, Liberator — Legendary Creature — Human Soldier Hero
// {3}{W}{W}, 5/5 (Marvel Super Heroes Commander, #4455):
//
//	"When Captain America enters, you may search your library for an
//	 Equipment card with mana value 3 or less, put it onto the battlefield,
//	 then shuffle.
//	 Whenever Captain America attacks, for each Equipment attached to him,
//	 create a 1/1 white Soldier creature token."
//
// The search is Garruk, Unleashed's optional find to the battlefield with
// Cloud, Midgar Mercenary's Equipment test and a mana-value cap (off the
// library, so X is 0). The Equipment arrives unattached. The attack trigger
// counts the Equipment attached to him as it resolves: with Captain America
// gone by then nothing is attached to him on the battlefield, so no tokens
// are made.
//
// Flagged reading: CR 608.2h would count the Equipment as it was attached to
// him when he last existed; the engine keeps no attachment record for a
// permanent that left, so that corner (Captain America removed in response to
// his own attack trigger) makes none, weaker than printed.
func init() {
	Register(Spec{
		OracleID:     "26e92eb8-7a23-4e7e-be83-244bb6078606",
		Name:         "Captain America, Liberator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Captain America, Liberator — you may search for an Equipment card with mana value 3 or less and put it onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return b09IsEquipmentCard(c) && c.ManaValue() <= 3 },
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Optional:  true,
						Shuffle:   true,
						Reason:    "Captain America, Liberator — an Equipment card with mana value 3 or less, onto the battlefield",
					}.Apply(NewContext(g, item))
				}),
			On(game.EventAttack, msc6ThisAttacks,
				"Captain America, Liberator — for each Equipment attached to him, a 1/1 white Soldier",
				Do(CreateTokensForEach{
					Template: TokenCard("1/1 white Soldier"),
					Name:     "Equipment attached to Captain America",
					Count: func(ctx *Context) int {
						n := 0
						for _, c := range ctx.Game.BattlefieldCardsForEffect() {
							if c.HasSubtype("Equipment") && c.IsAttachedTo(ctx.Source()) {
								n++
							}
						}
						return n
					},
				})),
		},
	})
}
