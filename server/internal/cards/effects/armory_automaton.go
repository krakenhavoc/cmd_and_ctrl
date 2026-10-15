package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Armory Automaton — Artifact Creature — Construct {3}, 2/2 (#3656):
//
//	"Whenever this creature enters or attacks, you may attach any
//	 number of target Equipment to it. (Control of the Equipment
//	 doesn't change.)"
//
// Gilgamesh's enters-or-attacks shape with a targeted "any number"
// clause over every Equipment on the battlefield (not only yours).
// AttachForEffect writes only the attachment link, so control never
// changes. Each target is re-checked on resolution (CR 608.2b); one
// that left or can't legally be attached stays put (CR 701.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa3335be-80b1-49d7-9809-5bef13be79f4",
		Name:         "Armory Automaton",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(
				Targeting(
					WhenThisEntersOrAttacks("Armory Automaton — you may attach any number of target Equipment to it",
						func(g *game.Game, item *game.StackItem) error {
							for _, t := range NewContext(g, item).LegalTargets() {
								if t.Kind != game.TargetCard {
									continue
								}
								if err := g.AttachForEffect(t.ID, game.TargetRef{Kind: game.TargetCard, ID: item.SourceCardID}); err != nil {
									return err
								}
							}
							return nil
						}),
					TargetPermanent("any number of target Equipment", HasSubtype("Equipment")).WithCount(0, 0)),
				"Armory Automaton — attach any number of target Equipment to it?"),
		},
	})
}
