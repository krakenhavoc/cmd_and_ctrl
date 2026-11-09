package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghired, Mirror of the Wilds — Legendary Creature — Human Shaman {R}{G}{W},
// 3/3 (Outlaws of Thunder Junction, #4468):
//
//	"Haste
//	 Nontoken creatures you control have '{T}: Create a token that's a copy
//	 of target token you control that entered this turn.'"
//
// A layer-6 ability grant (ADR 0093) to the controller's nontoken creatures
// (Ghired included). The granted {T} ability names a token its controller
// controls that entered this turn (the object's own entry, CR 400.7), and the
// copy takes the token's copiable values as it exists, or as it last existed
// if it has left (CR 707.2, 608.2h). The tap cost is a creature's, so the
// summoning-sickness rule applies to each recipient.
//
// No simplification.
func init() {
	const grant = "ghired-mirror-of-the-wilds/copy-a-new-token"
	Register(Spec{
		OracleID:        "34b36022-0d93-48cc-a2cc-90ca4d65dc31",
		Name:            "Ghired, Mirror of the Wilds",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Grants: []AbilityGrant{{
			Key: grant,
			Activated: []ActivatedAbility{{
				Label:   "{T}: Create a token that's a copy of target token you control that entered this turn.",
				Cost:    TapCost(),
				Targets: TargetPermanent("target token you control that entered this turn", IsTokenPredicate(), YouControl(), rfCreatureBEnteredThisTurn()),
				Effect:  b35CopyFirstTargetToken,
			}},
			Text: "{T}: Create a token that's a copy of target token you control that entered this turn.",
		}},
		Static: []game.StaticAbility{GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && !target.IsToken() && target.Controller == source.Controller
		}, grant)},
	})
}
