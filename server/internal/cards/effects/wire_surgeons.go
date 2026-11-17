package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wire Surgeons — Creature — Human Artificer {4}{B}{B}, 6/5 (The Brothers'
// War Commander, #5482):
//
//	"Fear
//	 Each artifact creature card in your graveyard has encore. Its encore
//	 cost is equal to its mana cost. (Exile it and pay its mana cost: For
//	 each opponent, create a token copy that attacks that opponent this turn
//	 if able. They gain haste. Sacrifice them at the beginning of the next
//	 end step. Activate only as a sorcery.)"
//
// A graveyard grant (Sedris's seam, ADR 0093) of the ordinary Encore row
// with its mana component read off the card that has the row (ItsManaCost,
// Cursecloth Wrappings' embalm). A card with no mana cost would owe a cost
// it can't pay (CR 118.6), so it is given nothing usable.
//
// No simplification.
func init() {
	encore := Encore("")
	encore.Label = "Encore — its mana cost, Exile this card from your graveyard: For each opponent, " +
		"create a token that's a copy of it that attacks that opponent this turn if able. They gain haste. " +
		"Sacrifice them at the beginning of the next end step. Activate only as a sorcery."
	encore.Cost = Plus(ItsManaCost(), ExileThis())
	Register(Spec{
		OracleID:        "c51078d9-a53a-49ae-9658-2bc00cce2b81",
		Name:            "Wire Surgeons",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Grants: []AbilityGrant{{
			Key:       "wire-surgeons/encore",
			Zone:      game.ZoneGraveyard,
			Activated: []ActivatedAbility{encore},
			Text:      "Encore. The encore cost is equal to its mana cost.",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilitiesInZone(game.ZoneGraveyard, CardsYouOwn(s82Bro13ArtifactCreatureCard), "wire-surgeons/encore"),
				"Each artifact creature card in your graveyard has encore. Its encore cost is equal to its mana cost."),
		},
	})
}
