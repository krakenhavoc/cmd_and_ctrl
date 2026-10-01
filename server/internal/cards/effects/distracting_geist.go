package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Distracting Geist // Clever Distraction (#1855, ADR 0107 §4) — a
// disturb card whose back face is an Aura.
//
// Front face, Creature — Spirit {2}{W}, 2/1:
//
//	"Whenever this creature attacks, tap target creature defending
//	 player controls.
//	 Disturb {4}{W} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Enchantment — Aura:
//
//	"Enchant creature
//	 Enchanted creature has "Whenever this creature attacks, tap target
//	 creature defending player controls."
//	 If Clever Distraction would be put into a graveyard from anywhere,
//	 exile it instead."
//
// "Defending player" is a fact about this attack, so the target clause
// is read off the attacker's own attack (TargetCreatureDefendingPlayer
// Controls, Goblin Racketeer's shape). The Aura's quoted ability is a
// granted bundle (ADR 0093) carrying the same trigger, so the enchanted
// creature's attack names its own defending player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         distractingGeistOracleID,
		Name:             "Distracting Geist",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{4}{W}")},
		Triggered:        []game.TriggeredAbility{distractingGeistTap("Distracting Geist — tap target creature defending player controls")},
	})
	Register(Spec{
		OracleID:     distractingGeistOracleID + "#1",
		Name:         "Clever Distraction",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Grants: []AbilityGrant{{
			Key:       cleverDistractionTap,
			Triggered: []game.TriggeredAbility{distractingGeistTap("Clever Distraction — tap target creature defending player controls")},
			Text:      "Whenever this creature attacks, tap target creature defending player controls.",
		}},
		Static:       []game.StaticAbility{GrantAbilitiesToAttached(cleverDistractionTap)},
		Replacements: []game.ReplacementEffect{DisturbedExile("Clever Distraction")},
	})
}

const (
	distractingGeistOracleID = "2fcd5779-7234-49e3-b3c5-ebc07db74462"
	cleverDistractionTap     = "clever-distraction/tap"
)

// distractingGeistTap is "Whenever this creature attacks, tap target
// creature defending player controls", printed on the front face and
// quoted on the back.
func distractingGeistTap(label string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:     []game.EventKind{game.EventAttack},
		AppliesTo:   ThisAttacked,
		TargetsFrom: TargetCreatureDefendingPlayerControls,
		Key:         label,
		Effect:      b36TapChosenCreature,
	}
}
