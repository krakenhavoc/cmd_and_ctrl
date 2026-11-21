package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const pathOfAnnihilationManaGrant = "path-of-annihilation/any-color-mana"

// Path of Annihilation — Enchantment {3}{G} (Modern Horizons 3):
//
//	"Devoid (This card has no color.)
//	 When this enchantment enters, create two 0/1 colorless Eldrazi Spawn
//	 creature tokens with "Sacrifice this token: Add {C}."
//	 Eldrazi you control have "{T}: Add one mana of any color."
//	 Whenever you cast a creature spell with mana value 7 or greater, you
//	 gain 4 life."
//
// The entry is Deep Forest Hermit's token trigger with the shared Eldrazi
// Spawn. "Eldrazi you control" is every permanent of that type, noncreature
// ones included (a layer 6 grant of the any-colour mana bundle, ADR 0093;
// Ambulatory Reactor's shape), read through the source's text so a changed
// word follows (CR 612.2); a Spawn gets both its own sacrifice ability and
// this one. The cast trigger is Helga's: the spell's mana value read on the
// stack, {X} counted (CR 202.3e), and it resolves above the spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eca6ffeb-463c-40b0-90b4-7e09e5579c10",
		Name:            "Path of Annihilation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Tokens: 2, CreatureTokens: 2, LifeGain: 4},
		Grants:          []AbilityGrant{AnyColorManaGrant(pathOfAnnihilationManaGrant)},
		Static: []game.StaticAbility{
			Labelled(GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Eldrazi"))
			}, pathOfAnnihilationManaGrant), `Eldrazi you control have "{T}: Add one mana of any color."`),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Path of Annihilation — create two 0/1 colorless Eldrazi Spawn",
				Do(CreateToken{Template: EldraziSpawnToken(), N: 2})),
			WheneverYouCast(And(Creature(), ManaValueGE(7)),
				"Path of Annihilation — you gain 4 life", Do(GainLife{Amount: 4})),
		},
	})
}
