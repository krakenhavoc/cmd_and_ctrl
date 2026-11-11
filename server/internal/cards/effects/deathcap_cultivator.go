package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deathcap Cultivator — Creature — Human Druid {1}{G}, 2/1 (Duskmourn
// Commander):
//
//	"{T}: Add {B} or {G}.
//	 Delirium — This creature has deathtouch as long as there are four or
//	 more card types among cards in your graveyard."
//
// The mana ability is Skull Prophet's. Delirium is an ability word (CR
// 207.2c): a layer 6 grant on the Cultivator alone, gated on the controller's
// own graveyard (drcDeliriumMet) and re-read on every recompute.
//
// No simplification.
func init() {
	deathtouch := KeywordGrant(whileYouHave(drcDeliriumMet, selfCreature), "deathtouch")
	deathtouch.Label = "Delirium — Has deathtouch as long as there are four or more card types among cards in your graveyard."
	Register(Spec{
		OracleID:     "26e555f1-e0ed-4b84-8783-1c0c968f462d",
		Name:         "Deathcap Cultivator",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B|G}",
			Label:    "Add {B} or {G}",
		}},
		Static: []game.StaticAbility{deathtouch},
	})
}
