package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hive of the Eye Tyrant — Land (Adventures in the Forgotten Realms):
//
//	"If you control two or more other lands, this land enters tapped.
//	 {T}: Add {B}.
//	 {3}{B}: Until end of turn, this land becomes a 3/3 black Beholder
//	 creature with menace and "Whenever this creature attacks, exile target
//	 card from defending player's graveyard." It's still a land."
//
// The animation is afrB17Manland, with menace as a layer-6 keyword and the
// quoted trigger as a granted bundle (ADR 0093). The trigger's target is
// chosen as it goes on the stack (CR 603.3d) from the graveyard of the
// player the land attacked, read off the attack at trigger time; it is
// re-checked as it resolves (CR 608.2b).
//
// No simplification.
const hiveOfTheEyeTyrantAttackExile = "hive-of-the-eye-tyrant/attack-exile"

func init() {
	attack := effects.WheneverThisAttacks("Hive of the Eye Tyrant — exile target card from defending player's graveyard", effects.OldB238ExileEachLegalTargetCard)
	attack.TargetsFrom = effects.OldB238DefendingGraveyardCardTarget
	effects.Register(effects.Spec{
		OracleID:      "d17163d4-dd43-4de6-b7cf-576448160b7f",
		Name:          "Hive of the Eye Tyrant",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.AfrB17EntersTappedWithTwoOtherLands()},
		ManaAbilities: effects.AfrB17TapForMana("B"),
		Grants: []effects.AbilityGrant{{
			Key:       hiveOfTheEyeTyrantAttackExile,
			Triggered: []game.TriggeredAbility{attack},
			Text:      "Whenever this creature attacks, exile target card from defending player's graveyard.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}: Until end of turn, this land becomes a 3/3 black Beholder creature with menace and \"Whenever this creature attacks, exile target card from defending player's graveyard.\" It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{3}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.AfrB17Manland(g, item, "Hive of the Eye Tyrant", "Beholder", "B", 3, 3, []string{"menace"}, hiveOfTheEyeTyrantAttackExile)
			},
		}},
	})
}

// hiveDefendingGraveyardCard is "target card from defending player's
// graveyard": any card in the graveyard of the player the source attacked.
// No defending player gives no clause (CR 603.3d drops it).
