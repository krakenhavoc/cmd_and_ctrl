package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alchemist's Talent — Enchantment — Class {3}{R}:
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 When this Class enters, create two tapped Treasure tokens.
//	 {1}{R}: Level 2
//	 Treasures you control have '{T}, Sacrifice this artifact: Add
//	 two mana of any one color.'
//	 {4}{R}: Level 3
//	 Whenever you cast a spell, if mana from a Treasure was spent to
//	 cast it, this Class deals damage equal to that spell's mana
//	 value to each opponent."
//
// Level 1 and level 2 are ordinary ADR 0071 Class machinery: the
// level-1 ETB and a level-2 grant gated `Level(2)` — which, per CR
// 716.2a, stays active at level 3 too, exactly like the printed
// cumulative levels. The grant's own cost has a component ADR 0093's
// bundle already carries: "Sacrifice this artifact" is
// ManaAbilityCost.Sacrifice, the Treasure sacrificing ITSELF, not the
// Class.
//
// Level 3 is the gap: nothing in the engine tracks which MANA SOURCE
// paid for a spell (a mana pool token records its colour and spend
// restrictions, never its origin), so "if mana from a Treasure was
// spent to cast it" has no signal to read. Shipping the trigger
// unconditionally (fire on every cast) would be stronger than
// printed; shipping it never (skip the trigger) is the declared,
// weaker-than-printed gap instead.
func init() {
	const grant = "alchemists-talent/treasure-mana"
	Register(Spec{
		OracleID:     "5a2dfff5-9dba-42f5-bcca-918c16f23807",
		Name:         "Alchemist's Talent",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Level 3's damage trigger never fires — the engine doesn't track whether a spell's mana came from a Treasure.",
		},
		Grants: []AbilityGrant{{
			Key:  grant,
			Text: "{T}, Sacrifice this artifact: Add two mana of any one color.",
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
				Produced: OneColorOfAmount(2),
				Label:    "Add two mana of any one color",
			}},
		}},
		Static: []game.StaticAbility{{
			Layer:          game.Layer6Ability,
			AppliesTo:      treasuresYouControl,
			GrantAbilities: []string{grant},
			ActiveWhen:     Level(2),
		}},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{1}{R}")),
			LevelUp(3, ManaCost("{4}{R}")),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Alchemist's Talent — create two tapped Treasure tokens",
				func(g *game.Game, item *game.StackItem) error {
					return b13CreateTappedTreasures(NewContext(g, item), item.Controller, 2)
				}),
		},
	})
}

// treasuresYouControl is "Treasures you control" — Alchemist's
// Talent's level-2 grant.
func treasuresYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.HasSubtype("Treasure")
}
