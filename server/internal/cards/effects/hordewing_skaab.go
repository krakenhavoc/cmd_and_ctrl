package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hordewing Skaab — Creature — Zombie Horror {4}{U}, 3/3 (EDHREC rank
// 4265):
//
//	"Flying
//	 Other Zombies you control have flying.
//	 Whenever one or more Zombies you control deal combat damage to
//	 one or more of your opponents, you may draw cards equal to the
//	 number of opponents dealt damage this way. If you do, discard
//	 that many cards."
//
// The Zombie deck's evasion piece: a wide board of ground Zombies all
// gain flying at once, which in a format full of ground stalls is the
// difference between a board and a win. The loot on connection is the
// rider that keeps the deck stocked.
//
// The lord half is TribeFilter{Others, YoursOnly} — the Skaab prints
// its own flying, so "other" costs it nothing, and an opponent's
// Zombies stay on the ground.
//
// The loot is ONE trigger per combat damage step, which is what the
// "one or more … one or more" wording asks for, and it counts the
// opponents dealt damage. The engine emits one damage event per
// creature, so the trigger waits for the step to settle (AtBatchEnd,
// ThatManyEvents) and the filter accepts only the FIRST Zombie event
// against each opponent (b41FirstZombieHitOnThatOpponent): the settled
// total is then the number of distinct opponents dealt damage this way,
// and a five-Zombie alpha strike into one opponent still loots once.
// "You may draw cards equal to that number. If you do, discard that
// many" is one optional draw-then-discard of that size.
func init() {
	Register(Spec{
		OracleID:        "485dea0d-2123-4e5c-91a1-25ba73f4f3cf",
		Name:            "Hordewing Skaab",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Zombie"}, Others: true, YoursOnly: true}, "flying"),
		},
		Triggered: []game.TriggeredAbility{
			Optional(withAtBatchEnd(
				On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b41ZombieYouControlDealtCombatDamageToAnOpponent(ev, source, g) &&
						b41FirstZombieHitOnThatOpponent(ev, source, g)
				}, "Hordewing Skaab — draw cards equal to the opponents dealt damage, then discard that many",
					func(g *game.Game, item *game.StackItem) error {
						return lootOne(g, item, settledEventAmount(item))
					}), ThatManyEvents()),
				"Hordewing Skaab — draw cards equal to the number of opponents dealt damage, then discard that many?"),
		},
	})
}
