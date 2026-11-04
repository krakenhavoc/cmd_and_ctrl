package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// sneak.go — #4586: sneak (CR 702.190), on Marvel Super Heroes and TMNT.
//
//	"Sneak {1}{B}{B} (You may cast this spell for {1}{B}{B} if you also
//	 return an unblocked attacker you control to hand during the declare
//	 blockers step. She enters tapped and attacking.)"
//
// An alternative cost (ADR 0025): the printed price, plus Daze's
// ReturnToHand component narrowed to ninjutsu's unblocked attacker
// (UnblockedAttacker, ninjutsu.go). AlternativeCost.Sneak adds the
// declare-blockers window and the tapped-and-attacking entry; the engine
// half is game/sneak.go. Give the offer to a card the way it is printed:
//
//	AlternativeCosts: []game.AlternativeCost{Sneak("{1}{B}{B}")},
//
// An instant or sorcery with sneak takes the same constructor: the
// window and the cost apply, and the entry clause has no permanent to
// apply to.

// Sneak is "Sneak <cost>" (CR 702.190a). `cost` is the printed sneak
// cost in brace notation. The key "sneak" is an on-disk identity: it
// rides StackItem.AltCost into snapshots. The return is
// game.SneakReturnSpec, the one a granted sneak pays too.
func Sneak(cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:          "sneak",
		Label:        "Sneak " + cost,
		ManaCost:     cost,
		ReturnToHand: game.SneakReturnSpec(),
		PayLabel:     game.SneakPayLabel,
		Sneak:        true,
	}
}

// CreatureCardsInYourGraveyardHaveSneak is "Creature cards in your
// graveyard have sneak <cost>. You may cast creature spells from your
// graveyard using their sneak abilities." (Ninja Teen's level 3, #4680).
// One standing permission over its controller's graveyard, Goblin
// Formula's shape (norman_osborn.go): it gives each creature card there a
// sneak offer at `cost` (game.AltCostKeyGrantedSneak) and opens the zone
// to a sneak cast of a creature spell, with that offer or the card's own,
// and to no other cast (game/sneak.go). `label` names the source in the
// picker. Gate it as the card prints it (a Class level goes in
// Spec.GatedCastPermissions).
func CreatureCardsInYourGraveyardHaveSneak(cost, label string) game.CastPermission {
	return game.CastPermission{
		Zone: game.ZoneGraveyard,
		// "Creature CARDS in your graveyard": the card as it is there
		// (CR 712.8a); the spell must also be a creature spell, which the
		// engine asks of the face being cast.
		Filter:     game.PermissionFilter{CardInZone: true, CreatureOnly: true},
		AltCostKey: game.AltCostKeyGrantedSneak,
		SneakCost:  cost,
		CastOnly:   true,
		Label:      label,
	}
}
