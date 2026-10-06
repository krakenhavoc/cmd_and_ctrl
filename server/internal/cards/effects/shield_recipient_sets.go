package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// shield_recipient_sets.go — the card-facing half of #2045 (ADR 0108 §7,
// amendment of 2026-10-06): shields that protect a set the other
// ShieldTargets can't name. The engine half is
// game/shield_recipient_filter.go.
//
// Append-only, mechanic-named. Writing a card:
//
//	PreventDamageFromSource{Protect: ShieldCreatures}
//	  — Forfend's "Prevent all damage that would be dealt to creatures
//	    this turn"
//	PreventDamageFromSource{Protect: ShieldCreaturesYouControl}
//	  — Divine Light's "to creatures you control" (not you)
//	PreventDamageFromSource{Protect: ShieldPlayers, CombatOnly: true}
//	  — Defend the Hearth's "all combat damage that would be dealt to
//	    players"
//	PreventDamageFromSource{Protect: ShieldPermanentsYouControl(QuerySubtype("Dog"))}
//	  — Pack Leader's "to Dogs you control"
//	PreventDamageFromSource{Protect: ShieldPermanents(QueryTypes("artifact"), QueryTypes("creature"))}
//	  — Ethersworn Shieldmage's "to artifact creatures" (every query
//	    must match)
//
// Each composes with the source fields: Queries, Filter, Choose. Never
// test the set as the shield resolves: the engine reads it as each
// damage event would be dealt (CR 611.2c, 615.1), so a creature that
// enters later is protected and one you stop controlling is not.

// shieldRecipients is a ShieldTarget protecting the set f names.
func shieldRecipients(f game.DamageRecipientFilter) ShieldTarget {
	return ShieldTarget{kind: shieldRecipientSet, recipients: f}
}

var (
	// ShieldCreatures is "to creatures": every creature, whoever
	// controls it (Forfend, Blinding Fog).
	ShieldCreatures = ShieldPermanents(QueryTypes("creature"))
	// ShieldCreaturesYouControl is "to creatures you control", without
	// you (Divine Light, Sivvi's Ruse).
	ShieldCreaturesYouControl = ShieldPermanentsYouControl(QueryTypes("creature"))
	// ShieldPlayers is "to players": every player, and no permanent
	// (Commencement of Festivities, Chameleon Blur).
	ShieldPlayers = shieldRecipients(game.DamageRecipientFilter{Players: true})
)

// combatDamageToPlayersShield is "Prevent all combat damage that would
// be dealt to players this turn" (Commencement of Festivities, Defend
// the Hearth).
func combatDamageToPlayersShield() PreventDamageFromSource {
	return PreventDamageFromSource{Protect: ShieldPlayers, CombatOnly: true}
}

// ShieldPermanents is "to <permanents>", whoever controls them: a
// permanent matching EVERY query ("artifact creatures" is
// QueryTypes("artifact") and QueryTypes("creature")).
func ShieldPermanents(all ...game.PermanentQuery) ShieldTarget {
	return shieldRecipients(game.DamageRecipientFilter{Permanents: true, Match: all})
}

// ShieldPermanentsYouControl is "to <permanents> you control": a
// permanent you control matching every query ("Dogs you control").
func ShieldPermanentsYouControl(all ...game.PermanentQuery) ShieldTarget {
	return shieldRecipients(game.DamageRecipientFilter{Permanents: true, Match: all,
		Controller: game.RecipientControllerYou})
}
