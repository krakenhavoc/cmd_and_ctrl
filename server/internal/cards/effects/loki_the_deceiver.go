package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loki, the Deceiver — Legendary Creature — God Sorcerer Villain {1}{U}{B}{R},
// 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever Loki attacks, create a tapped and attacking token that's a
//	 copy of another target Villain you control, except it isn't legendary
//	 and it's an Illusion in addition to its other types. Sacrifice that
//	 token at the beginning of the next end step.
//	 Whenever one or more Villains you control deal combat damage to a
//	 player, draw a card."
//
// Sandstorm Crasher's copy-and-sacrifice (ADR 0130 §11) with Loki's
// exceptions: the token's copiable values lose Legendary and gain the
// Illusion subtype. The "Villain" is a word of the card's text (CR 612.2).
// If the target is gone as the trigger resolves nothing is made (CR 608.2b).
// The draw is "one or more ... to a player": one trigger for each player
// the Villains hit in a damage step (CR 603.2c, Alela's key).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17e45fd8-c924-4b42-87ad-189469a8dd18",
		Name:         "Loki, the Deceiver",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventAttack, msc6ThisAttacks,
					"Loki, the Deceiver — a tapped and attacking token copy of another target Villain you control; sacrifice it at the next end step",
					lokiTheDeceiverCopy),
				Another(TargetCreature("another target Villain you control", YouControl(), HasSubtype("Villain")))),
			OncePerBatchPerPlayer(On(game.EventDealDamage,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b43CreatureOfSubtypeYouControlDealtCombatDamageToPlayer(ev, source, g, "Villain")
				}, "Loki, the Deceiver — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}

// lokiTheDeceiverCopy creates the token and (through its tail) schedules its
// sacrifice.
func lokiTheDeceiverCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := legalTargetIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	return msc16AttackingNonlegendaryCopies(ctx, item, ids[0], 1, "Illusion",
		"Loki, the Deceiver", lokiTheDeceiverSacrifice)
}

// lokiTheDeceiverSacrifice is the token's tail, keyed: the end step trigger
// that sacrifices it. An on-disk key.
var lokiTheDeceiverSacrifice = msc16EndStepForTokens("attacking-tokens/loki-the-deceiver-sacrifice-the-token",
	"Loki, the Deceiver — sacrifice the token", sacrificeListedCardsBody)
