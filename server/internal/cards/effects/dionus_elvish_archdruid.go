package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dionus, Elvish Archdruid — Legendary Creature — Elf Druid {3}{G},
// 3/3 (EDHREC rank 2498):
//
//	"Elves you control have "Whenever this creature becomes tapped
//	 during your turn, untap it and put a +1/+1 counter on it. This
//	 ability triggers only once each turn.""
//
// Every Elf you tap on your turn untaps and grows — a mana dork that
// taps for two and gets bigger doing it. The granted ability has no
// slot of its own in the catalog (a static cannot add a triggered
// ability to another permanent), so it is modelled the way Agent of
// the Iron Throne models its Background grant: as Dionus's own
// trigger, watching every Elf creature its controller controls —
// Dionus himself included, he is an Elf — and firing for the Elf
// that tapped. Observably the same: the ability exists while Dionus
// is on the battlefield and stops when he leaves, applies to Elves
// under HIS controller's control, and one Elf tapping fires it once.
//
// "Becomes tapped" reads two event kinds, because the engine taps an
// attacker without an EventTapCard (Magda, Brazen Outlaw's shape):
// an attacking Elf untaps and grows, a vigilance Elf does not.
// "During your turn" is the active player. "Only once each turn" is
// per Elf, tallied off the event log — see
// b23ElfYouControlBecameTappedFirstTimeThisTurn for the one edge that
// errs weaker. The untap and the counter run in one resolution; an
// Elf that left the battlefield in response gets neither.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5e5cf55-eb9e-4f01-9056-791215411b27",
		Name:         "Dionus, Elvish Archdruid",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventTapCard, game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b23ElfYouControlBecameTappedFirstTimeThisTurn(ev, source, g)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				elf := ev.CardID
				label := "Dionus, Elvish Archdruid — untap the Elf and put a +1/+1 counter on it"
				if c, ok := g.LookupCardForEffect(elf); ok {
					label = "Dionus, Elvish Archdruid — untap " + c.Name + " and put a +1/+1 counter on it"
				}
				return game.NewTriggeredItem(source, label,
					func(g *game.Game, item *game.StackItem) error {
						return b23UntapAndGrow(g, item, elf)
					})
			},
		}},
	})
}
