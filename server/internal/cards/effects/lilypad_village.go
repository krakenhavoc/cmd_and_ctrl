package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lilypad Village — Land (EDHREC rank 3766):
//
//	"{T}: Add {C}.
//	 {T}: Add {U}. Spend this mana only to cast a creature spell.
//	 {U}, {T}: Surveil 2. Activate only if a Bird, Frog, Otter, or Rat
//	 entered the battlefield under your control this turn."
//
// Bloomburrow's blue creature-type land. The {U} carries the cast /
// creature restriction tags and is a deliberate click, since the
// auto-tapper leaves restricted abilities alone. The surveil's
// activation condition (CR 602.1b, #743) reads
// Game.EnteredWithSubtypeThisTurn: the turn tally records each
// permanent's subtypes as it enters, under the player it entered
// under. So the question is answered about the creature as it
// entered, not as it is now. A Soldier that entered and later became
// every creature type (a Maskwood Nexus arriving afterwards) does not
// open it, and a Rat token that entered and has since died still does.
//
// One declared simplification, weaker than printed: the tally reads a
// permanent's printed types (with any copy effect it entered as) at
// the moment of entry, before continuous effects are applied to it, so
// a creature that had one of the four types only through a static
// ability as it entered (a Soldier entering while you control Maskwood
// Nexus) does not count. The printed card would count it.
func init() {
	Register(Spec{
		OracleID:     "5bb06e6f-e3af-4caa-b66d-77248ad46b61",
		Name:         "Lilypad Village",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"A creature that was a Bird, Frog, Otter, or Rat only because of another effect (such as Maskwood Nexus) when it entered doesn't enable the surveil ability."},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{U}",
				Label:    "Add {U} (creature spells only)",
				Restrictions: []string{
					ManaRestrictCast,
					ManaRestrictType("Creature"),
				},
			},
		},
		Activated: []ActivatedAbility{{
			Label:     "{U}, {T}: Surveil 2. Activate only if a Bird, Frog, Otter, or Rat entered the battlefield under your control this turn.",
			Cost:      Plus(ManaCost("{U}"), TapCost()),
			Condition: lilypadVillageAnimalEntered,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Surveil{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}

// lilypadVillageAnimalEntered is the Village's surveil condition.
func lilypadVillageAnimalEntered(g *game.Game, controller, _ uuid.UUID) bool {
	for _, t := range []string{"Bird", "Frog", "Otter", "Rat"} {
		if g.EnteredWithSubtypeThisTurn(controller, t) > 0 {
			return true
		}
	}
	return false
}
