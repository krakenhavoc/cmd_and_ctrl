package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beetle, Legacy Criminal — Legendary Creature — Human Rogue Villain
// {3}{U}, 3/3 (Marvel's Spider-Man):
//
//	"Flying
//	 {1}{U}, Exile this card from your graveyard: Put a +1/+1 counter on
//	 target creature. It gains flying until end of turn. Activate only as
//	 a sorcery."
//
// Campus Crier's graveyard ability (CR 113.6): exiling the card is a cost, so
// the ability resolves with the source gone. The counter and the flying
// both go to the one target; a target that left is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "87972de1-c1d7-408e-8d56-995590f639b8",
		Name:            "Beetle, Legacy Criminal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:        "{1}{U}, Exile this card from your graveyard: Put a +1/+1 counter on target creature. It gains flying until end of turn. Activate only as a sorcery.",
			Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:         Plus(ManaCost("{1}{U}"), ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Beetle — flying until end of turn"}.Apply(ctx)
			},
		}},
	})
}
