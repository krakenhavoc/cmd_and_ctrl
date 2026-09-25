package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vincent, Vengeful Atoner — Legendary Creature — Assassin {2}{R},
// 3/3 (EDHREC rank 4162):
//
//	"Menace
//	 Whenever one or more creatures you control deal combat damage to
//	 a player, put a +1/+1 counter on Vincent.
//	 Chaos — Whenever Vincent deals combat damage to an opponent, it
//	 deals that much damage to each other opponent if Vincent's power
//	 is 7 or greater."
//
// A three-mana menace creature that grows every combat and, once it
// is big enough, hits the whole table at once. The second ability is
// the win condition: at seven power a connected hit is 7 to the
// defender and 7 to each of the other two opponents.
//
// Three clauses, and the batching on each one is the card:
//
//   - "ONE OR MORE creatures you control deal combat damage TO A
//     PLAYER" is one trigger PER PLAYER damaged, not one per creature
//     and not one for the whole combat (CR 603.2c; Keeper of Fables'
//     2019-10-04 ruling states it for this wording). Three creatures
//     hitting one opponent is one counter; three creatures hitting
//     three different opponents is three. That is
//     OncePerBatchPerPlayer, which keys the dedupe on the damaged
//     player — the package already names that shape
//     (WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer),
//     shared with Keeper of Fables and Professional Face-Breaker.
//   - The counter goes on VINCENT, by name in the printed text, so it
//     lands on the source rather than on whatever connected.
//   - The Chaos ability is Vincent's OWN combat damage to an
//     OPPONENT, so a goaded Vincent forced to swing at a player he
//     does not treat as an opponent — there is no such player at a
//     Commander table — and damage to a planeswalker both fall
//     outside it.
//
// "That much damage" is the amount of the combat damage as it was
// dealt, captured off the event at announce; "each OTHER opponent"
// excludes the player who took the combat damage. The power check
// runs at RESOLUTION, reading Vincent's effective power then: he can
// be pumped in response to the trigger and the redirection happens,
// and a Vincent who has shrunk below seven does nothing. That is
// where the printed text puts the condition — after the effect, not
// as an intervening-if.
//
// A Vincent who has left the battlefield by the time the trigger
// resolves has no power to read, so the damage does not happen.
// Weaker than a reading that remembered his last power, never
// stronger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b8961205-87fa-4cce-aba9-84fa2f91a67f",
		Name:            "Vincent, Vengeful Atoner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Vincent, Vengeful Atoner — put a +1/+1 counter on Vincent", putCounterOnSelf),
			{
				Watches: []game.EventKind{game.EventDealDamage},
				Key:     "Vincent, Vengeful Atoner — Chaos: that much damage to each other opponent",
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ThisDealtCombatDamageToAPlayer(ev, source, game.Characteristic{}, g) &&
						ev.Target != source.Controller
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b40ChaosSpillover(g, item, item.Trigger.Event.Target, item.Trigger.Event.Amount, 7)
				},
			},
		},
	})
}
