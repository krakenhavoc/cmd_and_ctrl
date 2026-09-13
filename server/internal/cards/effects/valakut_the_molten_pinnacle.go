package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valakut, the Molten Pinnacle — Land (EDHREC rank 1323):
//
//	"This land enters tapped.
//	 Whenever a Mountain you control enters, if you control at least
//	 five other Mountains, you may have this land deal 3 damage to
//	 any target.
//	 {T}: Add {R}."
//
// The land that turns a mono-red ramp deck's late land drops into
// Lightning Bolts, and the reason Scapeshift is a win condition.
// Three clauses, all live: the tapped entry is the real CR 614
// self-replacement, the mana is an ordinary tap ability, and the
// trigger watches every battlefield entry for a permanent with the
// Mountain subtype under Valakut's controller — effective subtypes,
// so a land Urborg-style effects have made a Mountain counts, and a
// fetched, Cultivated or reanimated Mountain all count, as printed.
//
// "Five OTHER Mountains" excludes the one that just entered, and the
// intervening-if (CR 603.4) is checked BOTH when the trigger is put
// on the stack and again on resolution: an opponent who Wastelands a
// Mountain in response to the trigger leaves you at four others and
// the damage does not happen. The "you may" is the trigger prompt;
// the target is picked after it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1bc44216-4e06-4f66-89b7-5c327004604e",
		Name:         "Valakut, the Molten Pinnacle",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b12MountainYouControlEntered(ev, source, g) &&
					b12OtherMountainsControlled(g, source.Controller, ev.CardID) >= 5
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Valakut, the Molten Pinnacle — deal 3 damage to any target?"},
			Targets:        TargetAny(),
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				entered := ev.CardID
				return game.NewTriggeredItem(source, "Valakut, the Molten Pinnacle — 3 damage to any target",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 {
							return nil
						}
						// CR 603.4 — the intervening-if is re-checked
						// on resolution.
						if b12OtherMountainsControlled(g, item.Controller, entered) < 5 {
							return nil
						}
						return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: 3}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
