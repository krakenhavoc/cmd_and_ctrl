package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oliphaunt — Creature — Elephant {5}{R}, 6/4 (EDHREC rank 4155):
//
//	"Trample
//	 Whenever this creature attacks, another target creature you
//	 control gets +2/+0 and gains trample until end of turn.
//	 Mountaincycling {1} ({1}, Discard this card: Search your library
//	 for a Mountain card, reveal it, put it into your hand, then
//	 shuffle.)"
//
// A six-mana 6/4 trampler that is really played as a one-mana
// Mountain tutor: the body is a late-game top-deck and the
// mountaincycling is the reason it makes the deck. That is what makes
// the caveat below unusually large for this batch.
//
// The attack trigger is the part that works, and both halves of it
// are until-end-of-turn continuous effects — one of the five
// mechanics the 2026-09-18 re-triage freed (#279):
//
//   - "+2/+0" is a Layer 7c modification, snapshotted at resolution
//     against the one creature the trigger targeted.
//   - "gains trample" is a Layer 6 grant on the same creature.
//
// "ANOTHER target creature you control" is matched by NAME, the
// catalog's convention for "another" on a declared target clause
// (Benevolent Hydra, Dour Port-Mage): a target predicate is not told
// which permanent is the source. Oliphaunt cannot pump itself, as
// printed; it also cannot pump a SECOND Oliphaunt, which the printed
// card can. Noted as a caveat rather than buried.
//
// DECLARED SIMPLIFICATION — NO MOUNTAINCYCLING, the same shape as
// every other cycling card in the catalog and for the same two
// reasons: game.AbilityCost has no discard component, and the CR 602
// activation path only offers abilities on battlefield permanents,
// while cycling is activated from hand. Tracked by #655. Weaker than
// printed in the only direction we ship (#259).
func init() {
	Register(Spec{
		OracleID:        "186b2256-4af3-48cb-96b0-b0e80a7ee6dc",
		Name:            "Oliphaunt",
		Completeness:    CompletenessCaveats,
		PrintedKeywords: []string{"trample"},
		Caveats: []string{
			"Mountaincycling {1} is not implemented — the card can only be cast, never cycled from hand for a Mountain, which is most of the reason it is played.",
			"The attack trigger can't pump a second Oliphaunt, though the printed card can.",
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Oliphaunt — +2/+0 and trample until end of turn", func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (BoostUntilEOT{Target: id, Power: 2, Label: "Oliphaunt — +2/+0"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{
						Target:   id,
						Keywords: []string{"trample"},
						Label:    "Oliphaunt — trample",
					}.Apply(ctx)
				}),
				TargetCreature("another target creature you control", YouControl(), b03NotNamed("Oliphaunt")),
			),
		},
	})
}
