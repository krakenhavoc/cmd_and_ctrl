package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pinnacle Monk // Mystic Peak — the FRONT face, Creature — Djinn
// Monk {3}{R}{R}, 2/2:
//
//	"Prowess (Whenever you cast a noncreature spell, this creature
//	 gets +1/+1 until end of turn.)
//	 When this creature enters, return target instant or sorcery card
//	 from your graveyard to your hand."
//
// The land back (pay 3 life or enter tapped; {T}: Add {R}) is the
// mdfc_lands.go row under "<oracle>#1"; this is face 0, which keeps
// the bare oracle ID (game.CatalogKey).
//
// Prowess rides the canonical KeywordProwess token — the engine's
// keywordTriggersFor supplies the ability, so it needs no Triggered
// entry of its own. The ETB is Eternal Witness's shape narrowed to
// instant or sorcery and to the controller's own graveyard: mandatory
// pick, removed if nothing qualifies (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f3d48efa-910a-4872-a5b1-a353c5dbce99",
		Name:            "Pinnacle Monk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordProwess},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Pinnacle Monk — return target instant or sorcery card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					ctx := NewContext(g, item)
					if !ctx.IsTargetLegal(item.Targets[0]) {
						return nil
					}
					return ReturnFromGraveyard{Target: item.Targets[0].ID, Dest: game.ZoneHand}.Apply(ctx)
				}),
				TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					YouOwn(), Or(Instant(), Sorcery())),
			),
		},
	})
}
