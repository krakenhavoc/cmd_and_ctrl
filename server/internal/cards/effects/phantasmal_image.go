package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phantasmal Image — "You may have this creature enter as a copy of
// any creature on the battlefield, except it's an Illusion in
// addition to its other types and it has 'When this creature becomes
// the target of a spell or ability, sacrifice it.'"
//
// The card the copy machinery was missing a slot for. Its except
// clause is the one shape ADR 0043 declared out of scope: it ADDS A
// SUBTYPE (CR 707.9b) and GRANTS AN ABILITY (CR 707.9a), and CR
// 707.9a's second sentence makes both of them copiable — a Clone
// that copies a Phantasmal Image is an Illusion with the sacrifice
// trigger too.
//
// That is why the grant is declared as catalog data here (Grants)
// and named from the except clause rather than closed over: the copy
// carries only the bundle's key, in its copiable values, so copying
// the copy copies the grant and a snapshot can carry it. See
// ability_grant.go and server/internal/game/copy_grants.go.
//
// Shipping it without the drawback would have been strictly stronger
// than printed, which AGENTS.md §7 forbids — the whole cost of the
// card is that any removal spell, any Swords, any ward trigger, even
// a friendly Giant Growth kills it outright.
const phantasmalImageIllusionGrant = "phantasmal-image/illusion"

func init() {
	Register(Spec{
		OracleID:     "bde94af8-faea-41ff-8eed-ba642eac9968",
		Name:         "Phantasmal Image",
		Completeness: CompletenessCaveats,
		// The same engine-wide deviation Clone declares: the CR 704.5f
		// state-based action deliberately skips a creature with
		// printed toughness 0 and no counters, because that shape is
		// the placeholder convention for unparseable stats.
		Caveats: []string{"If you decline the copy, the 0/0 Phantasmal Image stays on the battlefield instead of dying."},
		Grants: []AbilityGrant{{
			Key: phantasmalImageIllusionGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventBecomesTarget, Self,
					"Phantasmal Image — sacrifice it",
					SacrificeThisIfStillOnBattlefield),
			},
		}},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Phantasmal Image",
				anyCreatureOnBattlefield,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddSubtype("Illusion")
					v.GrantAbility(phantasmalImageIllusionGrant)
				},
			),
		},
	})
}
