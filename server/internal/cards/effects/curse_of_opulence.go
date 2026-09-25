package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
// "IS ATTACKED" is once per combat in paper (CR 506.3 — one or more
// creatures attacking that player). The engine emits one EventAttack
// per attacker, so both bullets are batched by the OncePerBatch guard
// Professional Face-Breaker uses: one declaration is one batch (see
// AGENTS.md §7). Without that, three attackers would make three
// Golds, which is STRONGER than printed and not shippable.
//
// "EACH OPPONENT ATTACKING THAT PLAYER DOES THE SAME" — closed
// (previously a declared gap). Only the active player declares
// attackers in any one combat (CR 508.1), so "each opponent attacking
// that player" is never more than one player per batch; the second
// ability below fires for THAT attacker (excluding the Curse's own
// controller, since the clause says "opponent") and hands the token
// to them rather than to the Curse's controller, using the same
// item.Controller override WhenYouLoseControlOfThis uses. Two
// creatures from the same attacking player still mint that player
// exactly one Gold (OncePerBatch, keyed by the attacker's controller
// via BatchKey), matching the "IS ATTACKED" batching above.
func init() {
	Register(Spec{
		OracleID:     "ba0d3df2-3acf-46d7-8d64-8d67d1579adc",
		Name:         "Curse of Opulence",
		Completeness: CompletenessFull,
		Targets:      EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetPlayer &&
					source.AttachedTo.ID == ev.Target
			}, "Curse of Opulence — create a Gold token", Do(CreateToken{
				Template: GoldToken(),
				N:        1,
			}))),
			curseOfOpulenceEachAttackingOpponent(),
		},
	})
}

// curseOfOpulenceEachAttackingOpponent is "each opponent attacking
// that player does the same": a Gold token for the ATTACKING player,
// not the Curse's controller. Key is separate from the controller's
// own ability above so the two OncePerBatch guards don't collide, and
// BatchKey groups by the attacker's controller so a multi-creature
// swing from the same player still mints only one Gold.
func curseOfOpulenceEachAttackingOpponent() game.TriggeredAbility {
	const label = "Curse of Opulence — the attacking opponent creates a Gold token"
	t := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventAttack},
		Key:     label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			if source.AttachedTo.Kind != game.TargetPlayer || source.AttachedTo.ID != ev.Target {
				return false
			}
			return ev.Actor != uuid.Nil && ev.Actor != source.Controller
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label, Do(CreateToken{
				Template: GoldToken(),
				N:        1,
			}))
			item.Controller, item.Owner = ev.Actor, ev.Actor
			return item
		},
	}
	t.OncePerBatch = true
	t.BatchKey = func(ev game.Event, _ *game.Card, _ *game.Game) string {
		return ev.Actor.String()
	}
	return t
}
