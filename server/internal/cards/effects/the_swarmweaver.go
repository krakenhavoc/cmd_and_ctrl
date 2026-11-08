package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Swarmweaver — Legendary Artifact Creature — Scarecrow {2}{B}{G},
// 2/3 (Duskmourn: House of Horror, #4466):
//
//	"When The Swarmweaver enters, create two 1/1 black and green Insect
//	 creature tokens with flying.
//	 Delirium — As long as there are four or more card types among cards
//	 in your graveyard, Insects and Spiders you control get +1/+1 and have
//	 deathtouch."
//
// The tokens are a table row (CreateToken). The delirium half is a lord
// (TribalAnthem and TribalKeywordGrant over Insects and Spiders you
// control, read through the card's text, CR 612.2) gated on the
// controller's graveyard (drcDeliriumMet), re-read on every recompute
// (CR 611.3a). Delirium is an ability word (CR 207.2c). The Swarmweaver
// is a Scarecrow, so it pumps only the creatures it names.
//
// No simplification.
func init() {
	const delirium = "Delirium — As long as there are four or more card types among cards in your graveyard, Insects and Spiders you control "
	tribe := TribeFilter{Tribes: []string{"Insect", "Spider"}, YoursOnly: true}
	anthem := TribalAnthem(tribe, 1, 1)
	anthem.AppliesTo = whileYouHave(drcDeliriumMet, tribe.Matches)
	anthem.Label = delirium + "get +1/+1."
	deathtouch := TribalKeywordGrant(tribe, "deathtouch")
	deathtouch.AppliesTo = whileYouHave(drcDeliriumMet, tribe.Matches)
	deathtouch.Label = delirium + "have deathtouch."
	Register(Spec{
		OracleID:     "50120e90-2fdf-43c1-93e2-2230ee4e63fc",
		Name:         "The Swarmweaver",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("The Swarmweaver — create two 1/1 black and green Insect creature tokens with flying",
				Do(CreateToken{Template: TokenCard("1/1 black and green Insect with flying"), N: 2})),
		},
		Static: []game.StaticAbility{anthem, deathtouch},
	})
}
