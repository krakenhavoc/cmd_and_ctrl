package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spawn-Gang Commander — Creature — Eldrazi Goblin {3}{R}{R}, 2/2 (Modern
// Horizons 3):
//
//	"Devoid (This card has no color.)
//	 When you cast this spell, create three 0/1 colorless Eldrazi Spawn
//	 creature tokens with "Sacrifice this token: Add {C}."
//	 {1}{C}, Sacrifice an Eldrazi: This creature deals 2 damage to any
//	 target."
//
// Abundant Maw's cast trigger (it resolves above the spell, so the
// Spawn arrive even if the Commander is countered) and Airdrop Condor's
// sacrifice cost with the Eldrazi subtype. The {C} is colourless mana
// specifically; a Spawn is an Eldrazi, so the three tokens feed it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3641ceef-4f26-4647-ac12-c958524d68ca",
		Name:            "Spawn-Gang Commander",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Spawn-Gang Commander — create three 0/1 colorless Eldrazi Spawn",
				Do(CreateToken{Template: EldraziSpawnToken(), N: 3})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{C}, Sacrifice an Eldrazi: This creature deals 2 damage to any target.",
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Cost:    Plus(ManaCost("{1}{C}"), game.AbilityCost{SacrificeOther: sacrificeSpec("an Eldrazi", OfCreatureType("Eldrazi"))}),
			Targets: TargetAny(),
			Effect:  DamageToFirstTarget(2),
		}},
	})
}
