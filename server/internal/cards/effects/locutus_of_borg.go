package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Locutus of Borg — Legendary Artifact Creature — Borg {1}{W}{U}{B}, 3/3
// (Star Trek Commander, #4974):
//
//	"Lifelink
//	 Artifact creatures you control have "Whenever this creature deals
//	 combat damage to a player, create a 1/1 colorless Borg artifact
//	 creature token.""
//
// Lifelink rides PrintedKeywords. The quoted ability is a layer-6 grant
// of a catalog bundle (ADR 0093, Hive Mind Coprocessor's shape), so each
// artifact creature you control has its own copy: "this creature" is the
// one that connected and the token is its controller's. Locutus is an
// artifact creature, so it has the ability too; two Locutus grant two
// instances, which trigger separately (CR 603.2). The token is the
// printed Borg, a plain 1/1 colourless artifact creature.
//
// No simplification.
const locutusOfBorgGrant = "locutus-of-borg/combat-damage-borg-token"

func init() {
	const text = "Whenever this creature deals combat damage to a player, create a 1/1 colorless Borg artifact creature token."
	Register(Spec{
		OracleID:        "d0cff43a-2a22-4e05-9f71-983cb297989f",
		Name:            "Locutus of Borg",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Grants: []AbilityGrant{{
			Key: locutusOfBorgGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Locutus of Borg — create a 1/1 colorless Borg artifact creature token",
					Do(CreateToken{Template: TokenCard("1/1 colorless Borg artifact"), N: 1})),
			},
			Text: text,
		}},
		Static: []game.StaticAbility{Labelled(GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsArtifact() && target.IsCreature() && target.Controller == source.Controller
		}, locutusOfBorgGrant), "Artifact creatures you control have \""+text+"\"")},
	})
}
