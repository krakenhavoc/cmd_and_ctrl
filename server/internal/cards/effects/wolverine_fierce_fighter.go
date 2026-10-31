package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wolverine, Fierce Fighter — {2}{R}{G} Legendary Creature — Mutant
// Berserker Hero, 3/5 (Marvel Super Heroes, #4455, seam #4614):
//
//	"Haste
//	 When Wolverine enters, he fights up to one other target creature.
//	 If damage would be dealt to Wolverine, instead that damage is dealt,
//	 but all other damage already dealt to him is healed."
//
// The fight is the "up to one" enters trigger (Azog's "other", Mawloc's
// fight): it always goes on the stack, may take no target, and "other"
// keeps Wolverine from fighting himself. The last line is
// HealOtherDamageWhenDealtDamage (CR 614.1a, CR 701.69a): lethal damage
// is judged on the newest damage alone, damage dealt together with it
// stays, and wither/infect counters and deathtouch are not damage
// already dealt (game/heal_damage.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8429ba6a-6c06-4a83-a6ee-023b64b825f8",
		Name:            "Wolverine, Fierce Fighter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Replacements:    []game.ReplacementEffect{HealOtherDamageWhenDealtDamage()},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Wolverine — he fights up to one other target creature", itFightsTheFirstTarget),
				Another(TargetCreature("up to one other target creature").WithCount(0, 1))),
		},
	})
}
