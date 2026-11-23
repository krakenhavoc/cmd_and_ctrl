package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const dacksDuplicateDethroneGrant = "dacks-duplicate/dethrone"

// Dack's Duplicate — Creature — Shapeshifter {2}{U}{R}, 0/0 (Conspiracy:
// Take the Crown):
//
//	"You may have this creature enter as a copy of any creature on the
//	 battlefield, except it has haste and dethrone. (Whenever it attacks the
//	 player with the most life or tied for most life, put a +1/+1 counter on
//	 it.)"
//
// Copycrook's shape: EntersAsCopyOf over any creature on the battlefield,
// with the exception written into the copiable values (CR 707.9a): haste is
// a keyword added to the copy, and dethrone is a granted ability (Scourge of
// the Throne's trigger), so a later copy of the Duplicate has both too.
// Declined, or with nothing to copy, it is a 0/0 and dies (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "724767ca-bd84-4137-9133-4ad2e96b8428",
		Name:         "Dack's Duplicate",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       dacksDuplicateDethroneGrant,
			Text:      "Dethrone (Whenever this creature attacks the player with the most life or tied for most life, put a +1/+1 counter on it.)",
			Triggered: []game.TriggeredAbility{effects.GapB201Dethrone("dethrone: put a +1/+1 counter on it")},
		}},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf(
				"Dack's Duplicate",
				effects.AnyCreatureOnBattlefield,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddKeyword("haste")
					v.GrantAbility(dacksDuplicateDethroneGrant)
				},
			),
		},
	})
}
