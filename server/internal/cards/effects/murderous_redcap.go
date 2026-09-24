package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Murderous Redcap — Creature — Goblin Assassin {2}{B/R}{B/R}, 2/2
// (EDHREC rank 4215):
//
//	"When this creature enters, it deals damage equal to its power to
//	 any target.
//	 Persist (When this creature dies, if it had no -1/-1 counters on
//	 it, return it to the battlefield under its owner's control with a
//	 -1/-1 counter on it.)"
//
// The combo creature: with a sacrifice outlet and any +1/+1 counter
// effect it is an arbitrarily large amount of damage, and even
// without one it is four mana for four damage spread over two bodies.
//
// The ETB works fully. "Damage equal to ITS power" is read at
// RESOLUTION (CR 608.2h) off the Redcap's power, counters included —
// not Effective().Power alone (#1281: that field excludes counters,
// which is exactly what persist would put on this creature the moment
// it existed) — so an anthem, a counter effect, or persist itself once
// it lands all count. A Redcap that has left the battlefield before the
// trigger resolves still deals the damage, equal to its last-known power
// (#1379).
//
// Declared simplification: a Redcap that has LEFT deals that damage as
// a plain source — the damage tail reads lifelink and deathtouch off
// the battlefield only (ADR 0056 Decision 2, step 4; #1396), so a Redcap
// wearing a Basilisk Collar that is sacrificed in response pings
// without deathtouch or lifelink. Weaker than printed.
//
// "Any target" is the full CR 115.4 slot: a player, a creature, a
// planeswalker or a battle.
//
// DECLARED SIMPLIFICATION — NO PERSIST. The keyword is not
// implemented anywhere in the engine (the `persist.go` file in this
// package is the SPELL of that name, not the mechanic), and it needs
// two things the catalog has no slot for: a dies-trigger that returns
// the card to the battlefield with a counter already on it, and the
// "if it had no -1/-1 counters on it" intervening-if read against the
// permanent as it last existed.
//
// This is the half the card is played for, so the caveat is the
// larger part of the card: a player sleeving this gets a four-mana
// 2/2 that pings for two on arrival and then stays dead. Weaker than
// printed in the only direction we ship (#259); tracked under
// `undying-persist`.
func init() {
	Register(Spec{
		OracleID:     "a498bc70-36e7-4454-bc44-906893df38b8",
		Name:         "Murderous Redcap",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Persist is not implemented — the Redcap does not come back with a -1/-1 counter when it dies, which is the half of the card most decks play it for.",
			"If the Redcap has left the battlefield before its enters trigger resolves, it still deals damage equal to its last power, but without any lifelink or deathtouch it had.",
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Murderous Redcap — damage equal to its power to any target", func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					redcap, ok := ctx.TriggeringPermanent()
					if !ok || len(item.Targets) == 0 {
						return nil
					}
					power := redcap.Power
					return DealDamage{
						Source: item.SourceCardID,
						Target: item.Targets[0].ID,
						Amount: power,
					}.Apply(ctx)
				}),
				TargetAny(),
			),
		},
	})
}
