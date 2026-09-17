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
// activation condition (CR 602.1b, #743) reads this turn's entries:
// an EventETB whose Actor — the player it entered under — is you, of
// a card with one of the four subtypes. The card's types are read as
// it is now, so a Bird card that entered and has since died still
// counts, by its printed type line.
//
// One declared simplification, weaker than printed: a TOKEN that
// entered and has since left the battlefield has ceased to exist (CR
// 704.5d), and nothing records what types it had, so it does not count.
// The printed card would count it.
func init() {
	Register(Spec{
		OracleID:     "5bb06e6f-e3af-4caa-b66d-77248ad46b61",
		Name:         "Lilypad Village",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"A Bird, Frog, Otter, or Rat token that entered this turn and has already left the battlefield doesn't enable the surveil ability."},
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
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventETB || ev.Actor != controller || ev.CardID == uuid.Nil {
			continue
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			continue
		}
		for _, t := range []string{"Bird", "Frog", "Otter", "Rat"} {
			if c.HasSubtype(t) {
				return true
			}
		}
	}
	return false
}
