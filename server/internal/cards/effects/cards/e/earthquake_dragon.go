package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Earthquake Dragon — {14}{G} Creature — Elemental Dragon 10/10
//
//	"This spell costs {X} less to cast, where X is the total mana value of
//	 Dragons you control.
//	 Flying, trample
//	 {2}{G}, Sacrifice a land: Return this card from your graveyard to your
//	 hand."
//
// The reduction is read as the spell is cast (CR 601.2f): the Dragons the
// caster controls then, each at its mana value (a token is 0, a copy has the
// copied cost). The graveyard ability works from the graveyard only.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1903200d-0350-48a9-ac7c-b851f7df647f",
		Name:            "Earthquake Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.GapB216DragonsTotalManaValue,
				"This spell costs {X} less to cast, where X is the total mana value of Dragons you control."),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, Sacrifice a land: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.B08SacrificeALand()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
