package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mishra, Tamer of Mak Fawa — Legendary Creature — Human Artificer {3}{B}{R},
// 4/4 (The Brothers' War, #5482):
//
//	"Permanents you control have 'Ward—Sacrifice a permanent.'
//	 Each artifact card in your graveyard has unearth {1}{B}{R}. (…)"
//
// The ward is Coppercoat Vanguard's WardGranted: ward is a triggered ability
// (CR 702.21a), so the grant is Mishra carrying the trigger for every
// battlefield permanent his controller controls (himself included: the text
// says "permanents", not "other"). The payer sacrifices any permanent they
// control. The unearth is Sedris's graveyard grant of the ordinary Unearth
// row, to artifact cards, at {1}{B}{R}; it goes the pass after Mishra leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3be40d81-9be5-4f39-8ce4-920c26cf2748",
		Name:         "Mishra, Tamer of Mak Fawa",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       "mishra-tamer-of-mak-fawa/unearth",
			Zone:      game.ZoneGraveyard,
			Activated: []ActivatedAbility{Unearth("{1}{B}{R}")},
			Text:      "Unearth {1}{B}{R}",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilitiesInZone(game.ZoneGraveyard, CardsYouOwn(s82Bro13ArtifactCard), "mishra-tamer-of-mak-fawa/unearth"),
				"Each artifact card in your graveyard has unearth {1}{B}{R}."),
		},
		Triggered: []game.TriggeredAbility{
			WardGranted(WardSacrifice("a permanent", Permanent()), "Mishra, Tamer of Mak Fawa — permanents you control have ward—sacrifice a permanent",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller
				}),
		},
	})
}
