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
// "one or more … one or more" wording asks for. The engine emits one
// damage event per creature, so OncePerBatch declines every later
// event of the same batch; without it a five-Zombie alpha strike would
// loot five times, which is the #259 direction.
//
// Declared simplification, weaker than printed (#259): the Skaab draws
// and discards ONE card, not one per opponent dealt damage. "The
// number of opponents dealt damage this way" is a count over the whole
// damage batch, and the per-batch dedup that makes this one trigger
// keeps only the first event — the count of DISTINCT damaged opponents
// is not reconstructible from it. One is the floor and never more than
// printed: a strike that got through to three players loots once
// instead of three times.
func init() {
	Register(Spec{
		OracleID:     "485dea0d-2123-4e5c-91a1-25ba73f4f3cf",
		Name:         "Hordewing Skaab",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"It always loots exactly one card, even when your Zombies connected with several opponents at once — the printed card draws and discards one per opponent damaged.",
		},
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Zombie"}, Others: true, YoursOnly: true}, "flying"),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(Optional(
				On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b41ZombieYouControlDealtCombatDamageToAnOpponent(ev, source, g)
				}, "Hordewing Skaab — draw a card, then discard a card",
					func(g *game.Game, item *game.StackItem) error {
						return lootOne(g, item, 1)
					}),
				"Hordewing Skaab — draw a card, then discard a card?")),
		},
	})
}
