package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seasinger — Creature — Merfolk (0/1) for {1}{U}{U}:
//
//	"When you control no Islands, sacrifice this creature.
//	 You may choose not to untap this creature during your untap step.
//	 {T}: Gain control of target creature whose controller controls an Island for as long as you control this creature and this creature remains tapped."
//
// ADR 0109 §3 (#1894): a control change (layer 2) for as long as you
// control this creature AND it remains tapped. That is one duration
// with two conditions (Duration.Also, CR 611.2b): it ends the moment
// either stops, so untapping this creature, or losing control of it,
// gives the creature back for good. If either is already false as the
// ability resolves, nothing is taken. "When you control no Islands,
// sacrifice this creature" is a state trigger (ADR 0107 §1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13745b92-1e30-4075-9733-5d00da11408c",
		Name:         "Seasinger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenYouControlNo(effects.QuerySubtype("Island"), "Seasinger — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
		UntapOptOuts: []game.UntapOptOut{effects.MayChooseNotToUntapSelf("Seasinger — you may choose not to untap this creature")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Gain control of target creature whose controller controls an Island for as long as you control this creature and this creature remains tapped.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature whose controller controls an Island", seasingerControllerHasAnIsland()),
			Effect:  effects.GainControlOfTargetFor("Seasinger — control while you control it and it remains tapped", effects.WhileYouControlThisAndItRemainsTapped),
		}},
	})
}

// seasingerControllerHasAnIsland is "creature whose controller controls
// an Island", read as the target is chosen and again as the ability
// resolves (CR 608.2b).
func seasingerControllerHasAnIsland() effects.CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		// "Island" is a word of Seasinger's text, lent to the candidate
		// by the targeting walk (LentSubtype, CR 612.2).
		island := c.LentSubtype("Island")
		for _, p := range g.BattlefieldCardsForEffect() {
			if p.Controller == c.Controller && p.HasSubtype(island) {
				return true
			}
		}
		return false
	}
}
