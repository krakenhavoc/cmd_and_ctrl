package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kardur, Doomscourge — Legendary Creature — Demon Berserker, {2}{B}{R}, 4/3:
//
//	"When Kardur enters, until your next turn, creatures your
//	 opponents control attack each combat if able and attack a
//	 player other than you if able.
//	 Whenever an attacking creature dies, each opponent loses 1 life
//	 and you gain 1 life."
//
// #1599: the ETB half is The Akroan War's chapter II shape —
// OpponentsCreaturesAttackIfAble with OtherThanYou set (CR 701.15b's
// second requirement, "a player other than you"), for
// DurationUntilYourNextTurn. It is a requirement, not a
// characteristic change, so CR 611.2c does not lock the affected set:
// a creature an opponent casts during that turn cycle has to attack
// too, and the engine judges it with every other requirement at the
// table.
//
// DECLARED SIMPLIFICATION: the drain trigger is not implemented.
// "Whenever an attacking creature dies" needs to know, AFTER the
// creature has already moved to the graveyard, that it was attacking
// at the moment it died — and CR 603.10 last-known-information does
// not carry combat state (Card.AttackingTarget is cleared on every
// zone exit, same as Tapped and Counters, before any watcher's
// AppliesTo runs). Building this properly needs a combat-state LKI
// the engine does not keep for arbitrary watchers today; no second
// card is asking for it yet, so it is not being speculatively built
// here. The requirement half is real and does the card's other job.
func init() {
	Register(Spec{
		OracleID:     "bc14356c-3a1a-47af-9a6e-2b449de0331f",
		Name:         "Kardur, Doomscourge",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"When a creature dies while attacking, opponents don't lose life and you don't gain life — that trigger isn't implemented."},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kardur, Doomscourge — creatures your opponents control attack each combat if able and attack a player other than you if able",
				kardurDoomscourgeRequirement),
		},
	})
}

func kardurDoomscourgeRequirement(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return OpponentsCreaturesAttackIfAble{
		OtherThanYou: true,
		Duration:     DurationUntilYourNextTurn(ctx, item.Controller),
		Label:        "Kardur, Doomscourge — creatures your opponents control attack each combat if able and attack a player other than you if able",
	}.Apply(ctx)
}
