package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hazardroot Herbalist — Creature — Rabbit Druid {2}{G}, 1/4
// (Bloomburrow, #4467):
//
//	"Whenever you attack, target creature you control gets +1/+0 until
//	 end of turn. If that creature is a token, it also gains deathtouch
//	 until end of turn."
//
// Most Valuable Slayer's "whenever you attack" (one trigger per attack
// declaration, OncePerBatch) with a target creature you control, which
// need not be attacking. Whether it is a token is read as the ability
// resolves. A target that left in response, or that you no longer
// control, gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f40cb1af-3db6-4e90-830f-86ff4b60e239",
		Name:         "Hazardroot Herbalist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Hazardroot Herbalist — target creature you control gets +1/+0, and deathtouch if it's a token",
					hazardrootHerbalistPump),
				effects.TargetCreature("target creature you control", effects.YouControl()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0))),
		},
	})
}

// hazardrootHerbalistPump is the +1/+0 on the still-legal target, then
// deathtouch when that creature is a token.
func hazardrootHerbalistPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id := effects.FirstLegalBattlefieldTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if err := (effects.BoostUntilEOT{Target: id, Power: 1, Label: "Hazardroot Herbalist — +1/+0"}).Apply(ctx); err != nil {
		return err
	}
	if !c.IsToken() {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: "Hazardroot Herbalist — deathtouch"}.Apply(ctx)
}
