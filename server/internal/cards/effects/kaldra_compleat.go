package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kaldra Compleat — Legendary Artifact — Equipment for {7}:
//
//	"Living weapon
//	 Indestructible
//	 Equipped creature gets +5/+5 and has first strike, trample,
//	 indestructible, haste, and 'Whenever this creature deals combat
//	 damage to a creature, exile that creature.'
//	 Equip {7}"
//
// Seven mana for an indestructible hasty 5/5 first striker that
// attacks the turn it lands, and re-arms itself every time the Germ
// is answered, because the Equipment's own indestructible survives
// anything that kills the body.
//
// THE LAST CLAUSE IS NOT IMPLEMENTED, and it is the one that makes
// Kaldra terrifying in a creature mirror. "Whenever this creature
// deals combat damage to a creature, exile that creature" grants a
// TRIGGERED ABILITY to another permanent, which
// docs/engine-seams.md lists as an open seam: a layer-6 static can
// grant a KEYWORD the engine honours, and the catalog's trigger
// harvester reads triggers off the source card's own catalog entry,
// so there is nowhere for a granted trigger on a host to live. The
// seam is the PR, not this card — see the Caveat, which is the whole
// of the gap and is published on the catalog page.
//
// Shipping the clause silently omitted would have been the #259
// mistake: the card is strictly weaker than printed without it,
// which is the direction a simplification is allowed to go.
//
// The other four grants are real and all four have live consumers:
// first strike splits the damage step, trample assigns the excess,
// indestructible is read by the destruction path and the damage
// state-based action, haste clears the summoning-sickness gate on a
// Germ created this turn — which is what makes living weapon plus
// haste a seven-mana attack rather than a seven-mana investment.
//
// The Equipment's OWN indestructible is a printed keyword rather
// than a static, the Darksteel Plate distinction: true of the
// Equipment in every zone, and it survives the Germ dying.
func init() {
	Register(Spec{
		OracleID:        "7359e82b-db79-488d-a1d4-75a00f12a4cf",
		Name:            "Kaldra Compleat",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The equipped creature does not exile creatures it deals combat damage to — that part of the card does nothing. Everything else works: +5/+5, first strike, trample, indestructible, haste, and the Germ token."},
		PrintedKeywords: []string{"indestructible"},
		Triggered: []game.TriggeredAbility{
			LivingWeapon("Kaldra Compleat"),
		},
		Static: []game.StaticAbility{
			PumpAttached(5, 5),
			GrantToAttached("first strike", "trample", "indestructible", "haste"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{7}"),
		},
	})
}
