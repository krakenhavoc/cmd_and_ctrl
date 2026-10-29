package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bilbo, Luckwearer // Burglar's Plot — an Adventure card (CR 715, The
// Hobbit, #4454), oracle a73d9109…:
//
//	Bilbo, Luckwearer — Legendary Creature — Halfling Rogue {1}{U}, 1/1
//	  "Bilbo can't be blocked.
//	   Whenever Bilbo deals combat damage to a player, draw a card, then
//	   discard a card."
//	Burglar's Plot — Sorcery — Adventure {4}{U}
//	  "Exchange control of two target nonland permanents that share a card
//	   type. (Then exile this card. You may cast the creature later from
//	   exile.)"
//
// Bilbo is Suspicious Stowaway's front face: a self-only can't-be-blocked
// static and the combat-damage loot (b16DrawThenDiscard).
//
// Burglar's Plot is Switcheroo's exchange over two clauses whose second pick
// is related to the first (SharesACardTypeWithTarget, the Breaking of the
// Fellowship shape): the second target is another nonland permanent that has
// a card type in common with the first, judged as the spell is cast
// (CR 601.2c) and again as it resolves (CR 608.2b), on the types after the
// layers (so an animated artifact shares "creature"). The exchange is all or
// nothing (CR 701.12b): if either target is no longer legal nothing changes
// hands, and one player controlling both is a no-op the engine already
// handles. Neither half states a duration, so both are indefinite.
//
// No simplification.
const bilboLuckwearerOracleID = "a73d9109-36e0-40ea-ab06-5c5f7eb4ac54"

func init() {
	Register(Spec{
		OracleID:     bilboLuckwearerOracleID,
		Name:         "Bilbo, Luckwearer",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Bilbo, Luckwearer — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return b16DrawThenDiscard(g, item, 1) }),
		},
	})
	Register(Spec{
		OracleID:     bilboLuckwearerOracleID + "#1",
		Name:         "Burglar's Plot",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPermanent("target nonland permanent", Nonland()),
			Distinct(SharesACardTypeWithTarget(
				TargetPermanent("another target nonland permanent that shares a card type with it", Nonland()), 0)),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			a, okA := ctx.ClauseTarget(0)
			b, okB := ctx.ClauseTarget(1)
			if !okA || !okB || len(ctx.LegalTargets()) != 2 {
				return nil
			}
			return ExchangeControl{A: a.ID, B: b.ID, Label: "Burglar's Plot — exchange control"}.Apply(ctx)
		},
	})
}
