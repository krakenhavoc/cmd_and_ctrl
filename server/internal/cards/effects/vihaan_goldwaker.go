package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vihaan, Goldwaker — Legendary Creature — Dwarf Warlock {R}{W}{B}, 3/3
// (Outlaws of Thunder Junction Commander, #5368):
//
//	"Other outlaws you control have vigilance and haste. (Assassins,
//	 Mercenaries, Pirates, Rogues, and Warlocks are outlaws.)
//	 At the beginning of combat on your turn, you may have Treasures you
//	 control become 3/3 Construct Assassin artifact creatures in addition to
//	 their other types until end of turn."
//
// The static is two layer-6 keyword grants over the five outlaw subtypes
// (At Knifepoint's TribeFilter), "other" and "you control". The trigger is a
// "you may" at the beginning of combat; the Treasures it animates are fixed
// as it resolves (CR 611.2c), and each gains the Artifact and Creature types
// and the Construct and Assassin subtypes while keeping Treasure, with a
// base 3/3 (layer 7b), until end of turn. Being Assassins, the Construct
// Treasures are outlaws: they have vigilance and haste under Vihaan, which is
// what lets a Treasure made this turn attack.
//
// No simplification.
func init() {
	outlaws := TribeFilter{
		Tribes:    []string{"Assassin", "Mercenary", "Pirate", "Rogue", "Warlock"},
		Others:    true,
		YoursOnly: true,
	}
	Register(Spec{
		OracleID:     "0613a0d8-f6bc-41ca-8070-b2506f756fc8",
		Name:         "Vihaan, Goldwaker",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(outlaws, "vigilance"),
			TribalKeywordGrant(outlaws, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			Optional(
				AtBeginningOfYourCombat("Vihaan, Goldwaker — Treasures you control become 3/3 Construct Assassin artifact creatures until end of turn", vihaanTreasuresBecomeCreatures),
				"Vihaan, Goldwaker — have Treasures you control become 3/3 Construct Assassin artifact creatures until end of turn?"),
		},
	})
}

// vihaanTreasuresBecomeCreatures is the animation, one effect over every
// Treasure its controller controls as it resolves.
func vihaanTreasuresBecomeCreatures(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	mods := []game.Mod{
		game.AddTypesMod("Artifact", "Creature"),
		game.AddSubtypesMod("Construct", "Assassin"),
	}
	mods = append(mods, game.SetBasePTMods(3, 3)...)
	return ScopedEffectFor{
		Match:    And(OfSubtype("Treasure"), YouControl()),
		Mods:     mods,
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Vihaan, Goldwaker — Treasures become 3/3 Construct Assassin artifact creatures",
	}.Apply(ctx)
}
