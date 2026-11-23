package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "f3d48efa-910a-4872-a5b1-a353c5dbce99",
		Name:            "Pinnacle Monk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordProwess},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Pinnacle Monk — return target instant or sorcery card from your graveyard to your hand",
				effects.ReturnTargetInstantOrSorceryFromYourGraveyard),
				effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
			),
		},
	})
}
