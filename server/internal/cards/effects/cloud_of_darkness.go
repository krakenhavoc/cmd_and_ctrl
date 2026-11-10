package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloud of Darkness — Legendary Creature — Avatar {2}{B}{G}{G}, 3/3
// (Final Fantasy, #4462):
//
//	"Flying
//	 Particle Beam — When Cloud of Darkness enters, target creature an
//	 opponent controls gets -X/-X until end of turn, where X is the number
//	 of permanent cards in your graveyard."
//
// Gran Pulse Ochu's count of permanent cards in its controller's
// graveyard (CR 110.4), taken once as the ability resolves (the ruling),
// and Grim Bauble's shrink of the still-legal target. Particle Beam is an
// ability word. A target that left in response takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "525f57df-7d4d-4089-9003-4b01005558dc",
		Name:            "Cloud of Darkness",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Cloud of Darkness — target creature an opponent controls gets -X/-X until end of turn, where X is the number of permanent cards in your graveyard",
				cloudOfDarknessParticleBeam),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

func cloudOfDarknessParticleBeam(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	n := permanentCardsInGraveyard(g, item.Controller)
	return BoostUntilEOT{Target: id, Power: -n, Toughness: -n,
		Label: "Cloud of Darkness — -X/-X"}.Apply(ctx)
}
