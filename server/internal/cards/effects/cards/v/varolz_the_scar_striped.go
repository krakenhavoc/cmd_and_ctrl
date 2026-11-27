package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Varolz, the Scar-Striped — Legendary Creature — Troll Warrior {1}{B}{G}, 2/2
// (Dragon's Maze, S82 OLD b406, #5564):
//
//	"Each creature card in your graveyard has scavenge. The scavenge cost is
//	 equal to its mana cost. (Exile a creature card from your graveyard and pay
//	 its mana cost: Put a number of +1/+1 counters equal to that card's power
//	 on target creature. Scavenge only as a sorcery.)
//	 Sacrifice another creature: Regenerate Varolz."
//
// Sedris's graveyard grant with Scavenge's body (the exile cost and the
// counters equal to the exiled card's power) and Cursecloth Wrappings' "pay its
// mana cost" (ItsManaCost): the grant is a bundle row on every creature card
// you own in your graveyard, so a card that loses Varolz's grant, or a Varolz
// that leaves, takes the ability with it. The regeneration is Vampire
// Warlord's.
//
// No simplification.
func init() {
	const grant = "varolz-the-scar-striped/scavenge"
	scavenge := effects.Scavenge("")
	scavenge.Label = "Scavenge — its mana cost, Exile this card from your graveyard: Put a number of +1/+1 counters equal to this card's power on target creature. Activate only as a sorcery."
	scavenge.Cost = effects.Plus(effects.ItsManaCost(), effects.ExileThis())
	effects.Register(effects.Spec{
		OracleID:     "b6b83b88-0660-4227-895e-78ac43d01b66",
		Name:         "Varolz, the Scar-Striped",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       grant,
			Zone:      game.ZoneGraveyard,
			Activated: []effects.ActivatedAbility{scavenge},
			Text:      "Scavenge. The scavenge cost is equal to its mana cost.",
		}},
		Static: []game.StaticAbility{
			effects.Labelled(effects.GrantAbilitiesInZone(game.ZoneGraveyard, effects.CardsYouOwn(effects.CreatureCard), grant),
				"Each creature card in your graveyard has scavenge. The scavenge cost is equal to its mana cost."),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice another creature: Regenerate Varolz.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    effects.SacrificeAnotherN(1, "another creature", effects.Creature()),
			Effect:  effects.RegenerateThisSource,
		}},
	})
}
