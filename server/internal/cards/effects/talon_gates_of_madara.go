package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Talon Gates of Madara — Land — Gate:
//
//	"When this land enters, up to one target creature phases out.
//	 {T}: Add {C}.
//	 {1}, {T}: Add one mana of any color.
//	 {4}: Put this card from your hand onto the battlefield."
//
// Four printed lines, three separate primitives:
//
//   - THE ETB is PhaseOut over an "up to one" target — the same
//     (0, 1)-count shape Thassa, Deep-Dwelling's end-step blink
//     declares. Zero legal targets (or a decline that never happens,
//     since this isn't a "may") is a legal outcome: legalTargetCards
//     narrows to whatever is still there at resolution (CR 608.2b) and
//     PhaseOut does nothing when that list is empty.
//   - THE TWO MANA ABILITIES are Hall of Oracles' filter-land shape —
//     a free colourless tap and a second ability that spends floating
//     mana (ManaAbilityCost.Mana) for one mana of any colour. Declared
//     rather than derived, because the engine's synthetic land ability
//     only fires for a permanent with the BASIC supertype.
//   - THE HAND ABILITY is "put THIS card from your hand onto the
//     battlefield" — unlike the put_from_hand.go family (which asks
//     the player to PICK a card), the source here is fixed: the
//     activating player's own copy of Talon Gates in hand. It is an
//     ordinary CR 602 ability that functions from ZoneHand, exactly as
//     Cycling declares, paying {4} and then landing the card through
//     the same CR 614 entry pipeline every other "put onto the
//     battlefield" effect uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c45bf9d-a017-43bf-9e32-67810a8a217b",
		Name:         "Talon Gates of Madara",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Talon Gates of Madara — phase out a creature",
					func(g *game.Game, item *game.StackItem) error {
						return PhaseOut{Targets: legalTargetCards(item, g)}.Apply(NewContext(g, item))
					}),
				TargetCreature("up to one target creature").WithCount(0, 1)),
		},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
		Activated: []ActivatedAbility{{
			Label: "{4}: Put this card from your hand onto the battlefield.",
			Cost:  ManaCost("{4}"),
			Zones: []game.ZoneKind{game.ZoneHand},
			Effect: func(g *game.Game, item *game.StackItem) error {
				_, err := g.PutFromHandOntoBattlefieldForEffect(item.SourceCardID,
					game.HandEntryOptions{Controller: item.Controller})
				return err
			},
		}},
	})
}
