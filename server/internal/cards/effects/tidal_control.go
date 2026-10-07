package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tidal Control — Enchantment {1}{U}{U}:
//
//	"Cumulative upkeep {2} (At the beginning of your upkeep, put an
//	 age counter on this permanent, then sacrifice it unless you pay
//	 its upkeep cost for each age counter on it.)
//	 Pay 2 life or {2}: Counter target red or green spell. Any player
//	 may activate this ability."
//
// ADR 0106 PR 6 (#1793).
//
//   - Cumulative upkeep {2} is the CumulativeUpkeep constructor
//     (CR 702.24a), a mana cost, which is the form it supports. "You"
//     is Tidal Control's controller: the age counters and the bill are
//     theirs, whoever else has been using the counterspell.
//   - "Pay 2 life or {2}" is one ability with a choice of cost. Heart
//     of Kiran's precedent (docs/adding-cards.md, "Rather than pay")
//     writes it as two rows with the same effect, each with a real
//     cost that can go unpaid, so the activator picks the cost by
//     picking the row. Both rows are any-player (CR 602.2, 602.1b),
//     and the life or the mana is the ACTIVATOR's (CR 602.1a, 119.4).
//   - "Target red or green spell" is the target clause, so a spell of
//     neither colour cannot be pointed at, and one that has stopped
//     being red or green by resolution is no longer a legal target
//     (CR 608.2b). Countering it is CR 701.6a.
//
// No purpose for the bot: who wants a red or green spell countered is
// a judgement no Purpose field makes.
//
// No simplification.
func init() {
	counter := func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			return CounterTarget{StackID: t.ID}.Apply(ctx)
		}
		return nil
	}
	target := func() *game.TargetSpec {
		return TargetSpell("target red or green spell", Or(OfColor("R"), OfColor("G")))
	}
	Register(Spec{
		OracleID:     "855e200e-5375-4aac-b1f7-5113162e7e14",
		Name:         "Tidal Control",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeep("Tidal Control — cumulative upkeep {2}", "{2}"),
		},
		Activated: []ActivatedAbility{
			{
				Label:     "Pay 2 life: Counter target red or green spell. Any player may activate this ability.",
				Cost:      game.AbilityCost{Life: 2},
				Targets:   target(),
				AnyPlayer: true,
				Effect:    counter,
			},
			{
				Label:     "{2}: Counter target red or green spell. Any player may activate this ability.",
				Cost:      game.AbilityCost{Mana: "{2}"},
				Targets:   target(),
				AnyPlayer: true,
				Effect:    counter,
			},
		},
	})
}
