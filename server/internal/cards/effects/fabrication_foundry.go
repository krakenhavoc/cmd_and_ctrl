package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fabrication Foundry — Artifact {1}{W} (The Lost Caverns of Ixalan,
// #4470):
//
//	"{T}: Add {W}. Spend this mana only to cast an artifact spell or
//	 activate an ability of an artifact source.
//	 {2}{W}, {T}, Exile one or more other artifacts you control with
//	 total mana value X: Return target artifact card with mana value X
//	 or less from your graveyard to the battlefield. Activate only as a
//	 sorcery."
//
// The mana ability is Isolinear Processor's restriction with {W}: the
// mana may be spent only on an artifact spell or on an ability whose
// source is an artifact (CR 106.6), and the Foundry's own second
// ability is one.
//
// The second ability's X has no {X} symbol: it is the variable the
// exile cost defines, so the activator announces it with the
// activation (CR 601.2b, applied by CR 602.2b), chooses a target
// bounded by it (CR 601.2c), and then pays by exiling other artifacts
// whose mana values add up to exactly that X (CR 601.2h;
// ExileArtifactsWithTotalManaValueX, #5401). An activation whose
// artifacts total anything else is refused with nothing paid. The
// target is re-checked against the same X as the ability resolves (CR
// 608.2b). A permanent's {X} counts as zero (CR 202.3e), and the
// Foundry itself is never one of the artifacts ("other"). Sorcery
// timing is CR 602.5d. The row declares its target a reanimation, so
// the bot prices the pick as the permanent it becomes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0bb20c36-689f-4a7c-80b5-a12bda821847",
		Name:         "Fabrication Foundry",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "{T}: Add {W}. Spend this mana only to cast an artifact spell or activate an ability of an artifact source.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictType("Artifact")},
			)},
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}{W}, {T}, Exile one or more other artifacts you control with total mana value X: Return target artifact card with mana value X or less from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost: Plus(ManaCost("{2}{W}"), TapCost(),
				ExileArtifactsWithTotalManaValueX("one or more other artifacts you control with total mana value X")),
			Targets: TargetCardInGraveyard("target artifact card with mana value X or less from your graveyard", Artifact(), YouOwn()).
				WithManaValueAtMostX(),
			SorcerySpeed: true,
			// The bot prices the pick as the permanent it becomes.
			Purpose: ForTargets(Reanimates(0)),
			Effect:  returnFirstLegalGraveyardTargetToBattlefield,
		}},
	})
}
