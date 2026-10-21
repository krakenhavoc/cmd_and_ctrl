package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Onakke Oathkeeper — Creature — Ogre Spirit {1}{W}, 0/4:
//
//	"Creatures can't attack planeswalkers you control unless their
//	 controller pays {1} for each creature they control that's attacking
//	 a planeswalker you control.
//	 {4}{W}{W}, Exile this card from your graveyard: Return target
//	 planeswalker card from your graveyard to the battlefield."
//
// The first ability is the CR 508.1h attack tax (ADR 0080) with the
// planeswalker scope (ProtectingPlaneswalkers) narrowed by AppliesTo to
// attacks that name a planeswalker: unlike Sphere of Safety, an attack
// on the Oathkeeper's controller is free. The tax is charged per
// creature attacking one of those planeswalkers, so two creatures on
// one walker pay {2}.
//
// The second works only from the graveyard, and its cost exiles the
// card itself (Ichor Drinker's shape), so the Oathkeeper is gone before
// the planeswalker returns. It has no timing restriction. The returned
// walker enters with its printed loyalty (CR 306.5b).
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "080df9eb-c26b-494a-96b4-d7e47e501866",
		Name:         "Onakke Oathkeeper",
		Completeness: CompletenessFull,
		AttackTaxes:  []game.AttackTax{onakkeOathkeeperTax()},
		Activated: []ActivatedAbility{{
			Label:   "{4}{W}{W}, Exile this card from your graveyard: Return target planeswalker card from your graveyard to the battlefield.",
			Cost:    Plus(ManaCost("{4}{W}{W}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Targets: TargetCardInGraveyard("target planeswalker card from your graveyard", Planeswalker(), YouOwn()),
			Effect:  returnFirstLegalGraveyardTargetToBattlefield,
		}},
	})
}

// onakkeOathkeeperTax is "creatures can't attack planeswalkers you
// control unless …": the planeswalker scope, with attacks on the player
// left free.
func onakkeOathkeeperTax() game.AttackTax {
	t := ProtectingPlaneswalkers(AttackTax("{1}",
		"Creatures can't attack planeswalkers you control unless their controller pays {1} for each creature they control that's attacking a planeswalker you control."))
	t.AppliesTo = func(q game.AttackTaxQuery) bool { return q.TargetKind == game.AttackTargetPlaneswalker }
	return t
}
