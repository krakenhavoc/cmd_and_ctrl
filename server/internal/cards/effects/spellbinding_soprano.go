package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spellbinding Soprano — Creature — Human Bard {1}{R}, 2/2 (Streets of New
// Capenna Commander, #5491):
//
//	"Whenever this creature attacks, instant and sorcery spells you cast
//	 this turn cost {1} less to cast.
//	 Encore {3}{R} ({3}{R}, Exile this card from your graveyard: For each
//	 opponent, create a token copy that attacks that opponent this turn if
//	 able. They gain haste. Sacrifice them at the beginning of the next end
//	 step. Activate only as a sorcery.)"
//
// Goblin Maskmaker's promise over instant and sorcery spells: the attack
// leaves a CR 601.2f reduction of generic mana on its controller for every
// matching spell the rest of the turn, which outlives the Soprano. Two
// attacks in one turn make two promises. Encore is the shared keyword
// (encore.go, CR 702.141a); each token copy has the attack trigger too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dffd5f0f-3a7b-47ed-864e-06c752bdcefc",
		Name:         "Spellbinding Soprano",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Spellbinding Soprano — instant and sorcery spells you cast this turn cost {1} less to cast",
				Do(GrantNextSpellPromise{From: "Spellbinding Soprano", Promise: game.NextSpellPromise{
					Filter:             game.PermissionFilter{InstantOrSorceryOnly: true},
					Reduce:             1,
					EveryMatchingSpell: true,
					Text:               "Instant and sorcery spells you cast this turn cost {1} less to cast.",
				}})),
		},
		Activated: []ActivatedAbility{Encore("{3}{R}")},
	})
}
