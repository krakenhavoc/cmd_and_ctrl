package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ulrich's Kindred — Creature — Wolf {2}{R}, 3/2 (Innistrad Remastered,
// S82 GAP, #5948):
//
//	"Trample
//	 {3}{G}: Target attacking Wolf or Werewolf gains indestructible until
//	 end of turn."
//
// The target is any attacking Wolf or Werewolf, the Kindred itself and
// other players' included (a changeling counts, CR 702.73a), chosen as the
// ability goes on the stack and re-checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "41fd1557-8dba-4bbf-9954-5c4f4f541812",
		Name:            "Ulrich's Kindred",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}: Target attacking Wolf or Werewolf gains indestructible until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")),
			Cost:    effects.ManaCost("{3}{G}"),
			Targets: effects.TargetCreature("target attacking Wolf or Werewolf", effects.AttackingCreature(), effects.Or(effects.HasSubtype("Wolf"), effects.HasSubtype("Werewolf"))),
			Effect:  gainsIndestructibleUntilEndOfTurn,
		}},
	})
}

// gainsIndestructibleUntilEndOfTurn is "target creature gains
// indestructible until end of turn" for each still-legal target.
func gainsIndestructibleUntilEndOfTurn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, id := range effects.LegalTargetCards(item, g) {
		if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
