package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grilka, Widowed Warrior — Legendary Creature — Klingon Warrior {5}{R},
// 5/5 (Star Trek Commander, #4974):
//
//	"When Grilka enters, gain control of target creature an opponent
//	 controls until end of turn. Untap it. Until end of turn, it gains
//	 trample, haste, and "Whenever this creature deals combat damage to
//	 a player, create that many tapped Treasure tokens.""
//
// The threaten is Hot Pursuit's (gain control until end of turn, untap
// it). The three gains are one duration grant (Herd Heirloom's shape,
// ADR 0093): trample and haste ride the same record as the catalog
// bundle that carries the Treasure trigger, so all three end together in
// the cleanup step. The granted trigger is the CREATURE's, so the player
// who controls it as it deals the damage — you, this turn — creates the
// Treasures, as many as the damage it dealt to that player, tapped. A
// target that became illegal in response is not taken (CR 608.2b).
//
// No simplification.
const grilkaTreasureGrant = "grilka-widowed-warrior/treasure"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8366080-dd0c-40cf-84e9-f628dfe8f45c",
		Name:         "Grilka, Widowed Warrior",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grilkaTreasureGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Grilka, Widowed Warrior — create that many tapped Treasure tokens",
					func(g *game.Game, item *game.StackItem) error {
						if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
							return nil
						}
						return effects.CreateTokenAdvanced{Spec: effects.Token(effects.TreasureToken()).EntersTapped(), N: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
					}),
			},
			Text: "Whenever this creature deals combat damage to a player, create that many tapped Treasure tokens.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Grilka, Widowed Warrior — gain control of target creature until end of turn; untap it; it gains trample, haste and a Treasure trigger", grilkaThreaten),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			),
		},
	})
}

// grilkaThreaten takes the target until end of turn, untaps it, and gives
// it trample, haste and the Treasure trigger until end of turn.
func grilkaThreaten(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.GainControl{
		Target:   target,
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Grilka, Widowed Warrior — gain control until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantAbilitiesFor{
		Target: target,
		Keys:   []string{grilkaTreasureGrant},
		Also:   []game.Mod{game.AddKeywordsMod("trample", "haste")},
		Label:  "Grilka, Widowed Warrior — trample, haste and Treasures on combat damage",
	}.Apply(ctx)
}
