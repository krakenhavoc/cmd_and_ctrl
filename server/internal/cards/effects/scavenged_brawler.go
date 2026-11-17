package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scavenged Brawler — Artifact Creature — Construct {6}, 4/4 (The
// Brothers' War):
//
//	"Flying, vigilance, trample, lifelink
//	 {5}, Exile this card from your graveyard: Choose target creature. Put
//	 four +1/+1 counters, a flying counter, a vigilance counter, a trample
//	 counter, and a lifelink counter on that creature. Activate only as a
//	 sorcery."
//
// The graveyard ability is Renew's shape (renew.go): the card is exiled as
// part of the cost, the ability works from the graveyard only, at sorcery
// speed, and the target is chosen on activation. The counters are ordinary
// placements (CR 122.1b for the keyword ones); a target that left by
// resolution makes the whole ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d895863e-a3c5-4e1f-bd28-e47a8717bbd0",
		Name:            "Scavenged Brawler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "trample", "lifelink"},
		Activated: []ActivatedAbility{{
			Label: "{5}, Exile this card from your graveyard: Choose target creature. Put four +1/+1 counters, a flying counter, " +
				"a vigilance counter, a trample counter, and a lifelink counter on that creature. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{5}"), ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				ctx := NewContext(g, item)
				id := item.Targets[0].ID
				for _, c := range []struct {
					kind string
					n    int
				}{
					{game.CounterPlusOne, 4}, {game.CounterFlying, 1}, {game.CounterVigilance, 1},
					{game.CounterTrample, 1}, {game.CounterLifelink, 1},
				} {
					if err := (AddCounter{Target: id, Kind: c.kind, N: c.n}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
