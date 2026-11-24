package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ian Chesterton — Legendary Creature — Human Scientist {2}{W}, 2/3 (Doctor
// Who, #5948):
//
//	"Science Teacher — Each Saga spell you cast has replicate. The replicate
//	 cost is equal to its mana cost. (When you cast that Saga, copy it for
//	 each time you paid its replicate cost. Copies of Saga spells become
//	 tokens.)
//	 Doctor's companion (You can have two commanders if the other is the
//	 Doctor.)"
//
// Hatchery Sliver's grant over Sagas: a granted optional cost on every Saga
// spell its controller casts, priced at the spell's own mana cost
// (AdditionalCost.ManaCostIsSpellCost), with the replicate trigger that
// copies the spell once per payment (CR 702.56a). A copy of a permanent
// spell becomes a token as it resolves (CR 707.10, CR 111.13), so each copy
// enters as a Saga token with its lore counter (CR 714.3a). Doctor's
// companion is a deck-construction rule (CR 702.124m) the deck validator
// reads off the oracle text.
//
// No simplification.
func init() {
	spellCost := effects.ReplicateMana("", 20)
	spellCost.ManaCostIsSpellCost = true
	spellCost.Label = "Replicate"
	isSaga := effects.Subtype("Saga")
	effects.Register(effects.Spec{
		OracleID:     "32c9ce8c-a19d-49e4-bc6d-c7182fa00b51",
		Name:         "Ian Chesterton",
		Completeness: effects.CompletenessFull,
		GrantedOptionalCosts: []game.GrantedOptionalCost{effects.SpellsYouCastHave(
			"Each Saga spell you cast has replicate. The replicate cost is equal to its mana cost.", isSaga, spellCost)},
		Triggered: []game.TriggeredAbility{
			effects.S82SncReplicateTrigger("Ian Chesterton — replicate: copy the Saga spell",
				func(g *game.Game, caster uuid.UUID, spell game.Card) bool { return isSaga(g, caster, spell) }),
		},
	})
}
