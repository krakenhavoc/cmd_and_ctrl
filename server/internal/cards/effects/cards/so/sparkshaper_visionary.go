package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sparkshaper Visionary — Creature — Human Wizard {2}{U}, 0/5 (Commander
// Legends: Battle for Baldur's Gate):
//
//	"At the beginning of combat on your turn, choose any number of target
//	 planeswalkers you control. Until end of turn, they become 3/3 blue Bird
//	 creatures with flying, hexproof, and "Whenever this creature deals combat
//	 damage to a player, scry 1." (They're no longer planeswalkers. Loyalty
//	 abilities can still be activated.)"
//
// One effect on each target, at one timestamp (CR 613.7): the type line is
// set to Creature — Bird (layer 4, so the planeswalker type is gone), blue
// (layer 5), 3/3 (layer 7b), flying and hexproof (layer 6), and the scry
// trigger is a granted ability for the turn (ADR 0093). Their loyalty
// counters stay, and their own loyalty abilities stay on the card.
//
// No simplification.
const sparkshaperVisionaryGrant = "sparkshaper-visionary/scry"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e674793-6586-44a4-b936-a71203c5850a",
		Name:         "Sparkshaper Visionary",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Grants: []effects.AbilityGrant{{
			Key: sparkshaperVisionaryGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Sparkshaper Visionary — scry 1", effects.Do(effects.Scry{N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, scry 1.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat("Sparkshaper Visionary — your target planeswalkers become 3/3 blue Birds with flying, hexproof and a scry trigger until end of turn",
				sparkshaperVisionaryAnimate),
				effects.TargetPermanent("any number of target planeswalkers you control", effects.Planeswalker(), effects.YouControl()).WithCount(0, 0)),
		},
	})
}

func sparkshaperVisionaryAnimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.GrantAbilitiesFor{
			Target: t.ID,
			Keys:   []string{sparkshaperVisionaryGrant},
			Also: []game.Mod{
				game.SetTypesMod([]string{"Creature"}, []string{"Bird"}),
				game.SetColorsMod(ctx.TextColor("U")),
				game.SetBasePowerMod(3),
				game.SetBaseToughnessMod(3),
				game.AddKeywordsMod("flying", "hexproof"),
			},
			Duration: effects.DurationUntilEndOfTurn(ctx),
			Label:    "Sparkshaper Visionary — becomes a 3/3 blue Bird creature with flying, hexproof and a scry trigger",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
