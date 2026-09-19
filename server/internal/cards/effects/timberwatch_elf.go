package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Timberwatch Elf — 1/2 Creature — Elf for {2}{G} (EDHREC rank 4365):
//
//	"{T}: Target creature gets +X/+X until end of turn, where X is
//	 the number of Elves on the battlefield."
//
// The Elf deck's combat trick and its finisher at once. On a board of
// eight Elves it is a free +8/+8, and because the ability taps rather
// than costs mana it is the reason an Elf board that has already
// spent its mana can still win the fight. Roadmap batch 42 (#449)
// filed it under cost modification, which #93 shipped; what it
// actually needs is an until-end-of-turn pump with a computed amount,
// and #279 shipped that.
//
// "The number of Elves ON THE BATTLEFIELD" — the whole table's, not
// just yours. Playing a second Elf deck across the table makes your
// Timberwatch Elf bigger, and that is the printed card.
//
// X is counted when the ability RESOLVES, not when it is activated
// (CR 608.2f): killing an Elf in response makes the pump smaller, and
// playing one makes it bigger. Timberwatch Elf counts itself — it is
// an Elf on the battlefield, and it stays one while tapped.
//
// EFFECTIVE subtypes, so a changeling counts as an Elf and anything a
// layer-4 effect has turned into an Elf counts too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50cee3ac-cba0-4abb-babf-de1928b1590e",
		Name:         "Timberwatch Elf",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature gets +X/+X until end of turn, where X is the number of Elves on the battlefield.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				x := b42CountCreaturesWithSubtype(g, "Elf")
				return BoostUntilEOT{
					Target: item.Targets[0].ID,
					Power:  x, Toughness: x,
					Label: "Timberwatch Elf — +X/+X until end of turn",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
