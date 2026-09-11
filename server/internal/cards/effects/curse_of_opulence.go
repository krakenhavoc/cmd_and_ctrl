package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curse of Opulence — Enchantment — Aura Curse for {R}:
//
//	"Enchant player
//	 Whenever enchanted player is attacked, create a Gold token. Each
//	 opponent attacking that player does the same."
//
// The Curse is in this batch for a structural reason rather than a
// power one: it is the card that proves the TargetPlayer branch of
// the attachment relation end to end. `Card.AttachedTo` is a
// TargetRef and not a card ID precisely so that "enchant player" has
// somewhere to live — a bare UUID could not say whether it named a
// seat or a permanent — and this is the card that exercises it,
// including the state-based action that puts the Curse in the
// graveyard when the enchanted player is eliminated.
//
// The one-red-mana political card: point it at whoever is ahead and
// the whole table is paid to attack them.
//
// TWO SIMPLIFICATIONS, both strictly weaker than printed:
//
//   - "EACH OPPONENT ATTACKING THAT PLAYER DOES THE SAME" is not
//     implemented. Only the Curse's controller gets a Gold token.
//     The clause is a per-attacking-player fan-out over a batch the
//     engine does not assemble — EventAttack is emitted once per
//     attacking CREATURE, so counting distinct attacking players
//     means reading the whole declared-attackers set, which is a
//     combat-batching problem rather than an attachment one. It is
//     also the half that makes the Curse political rather than
//     merely good, so this is the clause to come back for.
//   - "IS ATTACKED" is once per combat in paper (CR 506.3 — one or
//     more creatures attacking that player). The engine emits one
//     EventAttack per attacker, so the trigger is batched by the
//     same triggerAlreadyPendingFrom guard Professional Face-Breaker
//     uses: the second and later attackers in one declaration find a
//     Curse trigger already waiting and decline. Without it, three
//     attackers would make three Golds, which is STRONGER than
//     printed and not shippable.
//
// Deferred to whichever sprint teaches the engine to see a declared
// attack as one batch.
func init() {
	Register(Spec{
		OracleID: "ba0d3df2-3acf-46d7-8d64-8d67d1579adc",
		Name:     "Curse of Opulence",
		Targets:  EnchantPlayer(),
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetPlayer &&
					source.AttachedTo.ID == ev.Target &&
					!triggerAlreadyPendingFrom(g, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Curse of Opulence — create a Gold token",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   GoldToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
