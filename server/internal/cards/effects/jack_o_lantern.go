package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jack-o'-Lantern — Artifact {1} (Innistrad: Midnight Hunt, #5896):
//
//	"{1}, {T}, Sacrifice this artifact: Exile up to one target card from
//	 a graveyard. Draw a card.
//	 {1}, Exile this card from your graveyard: Add one mana of any color."
//
// The first ability is Stonespeaker Crystal's shape on one card: up to
// one target, so with no target it is "draw a card", and with a target
// that left the graveyard in response it does not resolve and draws
// nothing (CR 608.2b).
//
// The second is a mana ability (CR 605.1a: no target, it could add
// mana, not a loyalty ability) that states the zone it functions from,
// so it functions only from the graveyard (CR 113.6b). Off the
// battlefield the card's owner is its "you" (CR 108.4). It pays its {1}
// first and then exiles the card, both while paying for something else
// (CR 605.3a) and on its own, and the auto-tapper's costed pass may
// plan it (game/autotap_zone_costed.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "192b5dd7-f420-4d5b-86d6-66e86fa67f41",
		Name:         "Jack-o'-Lantern",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Exile up to one target card from a graveyard. Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Targets: TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1),
			Effect:  Do(ExileChosenTarget{}, DrawCards{N: 1}),
		}},
		ManaAbilities: []ManaAbility{{
			Label:    "{1}, Exile this card from your graveyard: Add one mana of any color.",
			Produced: "{W|U|B|R|G}",
			Zones:    []game.ZoneKind{game.ZoneGraveyard},
			Cost:     ManaAbilityCost{Mana: "{1}", ExileSelf: true},
		}},
	})
}
