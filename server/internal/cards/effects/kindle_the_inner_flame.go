package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kindle the Inner Flame — Kindred Sorcery — Elemental {3}{R}:
//
//	"Create a token that's a copy of target creature you control, except
//	 it has haste and 'At the beginning of the end step, sacrifice this
//	 token.'
//	 Flashback—{1}{R}, Behold three Elementals. (You may cast this card
//	 from your graveyard for its flashback cost. Then exile it. To behold
//	 an Elemental, choose an Elemental you control or reveal an Elemental
//	 card from your hand.)"
//
// The front half is Electroduplicate's (TokenCopyWithHasteAndGrant): haste
// and the quoted trigger are copiable values of the token (CR 707.9a), so
// a copy of the token has the trigger too and sacrifices itself, and the
// trigger fires at the beginning of every end step the token sees.
// The flashback cost is mana plus "behold three Elementals" (CR 702.34a,
// CR 701.4a, #4689): FlashbackBehold, whose three distinct picks ride
// alt_cost_ids, each an Elemental the caster controls or an Elemental card
// in their hand (a changeling counts). The card itself is on the stack as
// the cost is paid (CR 601.2a), so it never beholds itself. With fewer than
// three Elementals the flashback offer is not on the table.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:         "c7ca6371-73bb-4717-a148-0804b8f4c1e6",
		Name:             "Kindle the Inner Flame",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{FlashbackBehold("{1}{R}", 3, "Elemental", "Elementals")},
		Targets:          TargetCreature("target creature you control", YouControl()),
		Grants:           []AbilityGrant{SacrificeThisTokenAtEndStepGrant(kindleTheInnerFlameGrant, "Kindle the Inner Flame — sacrifice this token")},
		OnResolve:        TokenCopyWithHasteAndGrant(kindleTheInnerFlameGrant),
	})
}

// kindleTheInnerFlameGrant is the key of the granted sacrifice trigger, an
// on-disk identity.
const kindleTheInnerFlameGrant = "kindle-the-inner-flame/sacrifice-this-token"
