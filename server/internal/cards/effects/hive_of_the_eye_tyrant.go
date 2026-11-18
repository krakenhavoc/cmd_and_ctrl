package effects

import (
	"github.com/google/uuid"

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
	attack := WheneverThisAttacks("Hive of the Eye Tyrant — exile target card from defending player's graveyard",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, id := range legalTargetCards(item, g) {
				if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	attack.TargetsFrom = hiveDefendingGraveyardCard
	Register(Spec{
		OracleID:      "d17163d4-dd43-4de6-b7cf-576448160b7f",
		Name:          "Hive of the Eye Tyrant",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{afrB17EntersTappedWithTwoOtherLands()},
		ManaAbilities: afrB17TapForMana("B"),
		Grants: []AbilityGrant{{
			Key:       hiveOfTheEyeTyrantAttackExile,
			Triggered: []game.TriggeredAbility{attack},
			Text:      "Whenever this creature attacks, exile target card from defending player's graveyard.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}{B}: Until end of turn, this land becomes a 3/3 black Beholder creature with menace and \"Whenever this creature attacks, exile target card from defending player's graveyard.\" It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{3}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return afrB17Manland(g, item, "Hive of the Eye Tyrant", "Beholder", "B", 3, 3, []string{"menace"}, hiveOfTheEyeTyrantAttackExile)
			},
		}},
	})
}

// hiveDefendingGraveyardCard is "target card from defending player's
// graveyard": any card in the graveyard of the player the source attacked.
// No defending player gives no clause (CR 603.3d drops it).
func hiveDefendingGraveyardCard(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	defender := g.DefendingPlayerForAttackForEffect(source.AttackingTarget)
	if defender == uuid.Nil {
		return nil
	}
	return &game.TargetSpec{
		Mode:  "card_in_graveyard",
		Label: "target card from defending player's graveyard",
		Zones: []game.ZoneKind{game.ZoneGraveyard},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.Owner == defender
		},
		Min: 1, Max: 1,
	}
}
