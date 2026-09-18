package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ability_grant.go — "except it has '<ability>'" (CR 707.9a).
//
// Phantasmal Image copies a creature "except it's an Illusion in
// addition to its other types and it has 'When this creature becomes
// the target of a spell or ability, sacrifice it.'" Sakashima the
// Impostor grants "{2}{U}{U}: return it at the beginning of the next
// end step". Both are abilities the COPIED card never had, and both
// are copiable in their own right: a Clone of a Phantasmal Image has
// the trigger too.
//
// So a grant cannot be a closure on the replacement that made the
// copy. It is declared HERE, as static catalog data on the card that
// grants it, and the copy carries only its NAME
// (game.PrintedValues.GrantAbility). Register files each grant's own
// *game.CardDef in the same `defs` map cards use, under
// game.GrantKey(Key) — the shape #623 already uses for emblems, for
// the same reason: the object needs abilities the catalog can find,
// and it is not a card.
//
// The engine half is server/internal/game/copy_grants.go, which
// explains how the name becomes a lookup.

// AbilityGrant is one bundle of abilities a copy effect's "except"
// clause can give the copy. A card declares its bundles in
// Spec.Grants and names one from its except clause:
//
//	Grants: []AbilityGrant{{
//	    Key:       "phantasmal-image/illusion",
//	    Triggered: []game.TriggeredAbility{ ... },
//	}},
//	...
//	func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
//	    v.AddSubtype("Illusion")
//	    v.GrantAbility("phantasmal-image/illusion")
//	},
//
// One bundle per printed quoted ability. Splitting a card's two
// granted abilities into two bundles costs nothing and makes the
// except clause read like the card.
type AbilityGrant struct {
	// Key is the bundle's name, unique across the whole catalog.
	// Namespace it with the granting card ("<card>/<what>") — it is
	// the string the copy stores and the string a snapshot carries,
	// so it is as permanent as an oracle ID and a collision would
	// silently give one card another's ability.
	Key string

	// Triggered, Static and Activated are the three ability kinds the
	// engine reads off a permanent's catalog entry, declared exactly
	// as a card declares its own. Mana abilities are deliberately
	// absent: no printed copy effect grants one, and the engine reads
	// a permanent's mana abilities off the card object as well as the
	// catalog, which would be a second seam.
	Triggered []game.TriggeredAbility
	Static    []game.StaticAbility
	Activated []ActivatedAbility
}

// buildGrantDef projects one bundle into the shape the engine reads.
// The twin of buildEmblemDef, and like it a *game.CardDef that is
// not a card's.
func buildGrantDef(gr AbilityGrant) *game.CardDef {
	return &game.CardDef{
		Triggered: gr.Triggered,
		Static:    gr.Static,
		Activated: activatedShapes(gr.Activated),
	}
}

// checkGrants is the registration guard. Both failures are card files
// that are wrong in a way no test of the card would catch — a copy
// naming a bundle nobody registered gets a silent no-ability grant —
// so they fail at boot.
func checkGrants(name string, grants []AbilityGrant) {
	for _, gr := range grants {
		if gr.Key == "" {
			panic("effects.Register: " + name + " declares an ability grant with no Key — the copy has no name to carry")
		}
		if len(gr.Triggered) == 0 && len(gr.Static) == 0 && len(gr.Activated) == 0 {
			panic("effects.Register: " + name + " declares the ability grant " + gr.Key + " with no abilities — CR 707.9a grants an ability or nothing")
		}
		if _, taken := defs[game.GrantKey(gr.Key)]; taken {
			panic("effects.Register: " + name + " declares the ability grant " + gr.Key + ", which is already registered — grant keys are catalog-wide")
		}
	}
}
