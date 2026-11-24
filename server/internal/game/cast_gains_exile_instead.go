package game

import "github.com/google/uuid"

// cast_gains_exile_instead.go — #6091: "If you do, it gains 'If this
// permanent would leave the battlefield, exile it instead of putting it
// anywhere else.'" (The Eighth Doctor), a rider of the cast or play
// permission (CastPermission.GainsExileInsteadOfLeaving).
//
// The replacement is ModExileInsteadOfLeaving, the one unearth and Whip of
// Erebos register (CR 614.1a), as its own scoped record pinned to the
// spell's instance for ForThePermanentThisSpellBecomesDuration, the
// duration the keyword and ability gains use: it is in place before the
// permanent can leave, and it ends for good when that object leaves
// (CR 400.7). A land played through the permission gets it as its entry
// finishes, before the entry is announced, as the land's other gains do.

// grantExileInsteadOfLeavingForCastLocked registers the rider for the
// permanent `cardID` is or will become. Nothing when the permission did
// not say it. Caller must hold g.mu (write).
func (g *Game) grantExileInsteadOfLeavingForCastLocked(cardID uuid.UUID, gains permissionCastGains) {
	if !gains.exileInsteadOfLeaving {
		return
	}
	label := "Cast this way — if it would leave the battlefield, exile it instead"
	if gains.label != "" {
		label = gains.label + " — if it would leave the battlefield, exile it instead"
	}
	g.RegisterScopedEffectForEffect(cardID, []AffectedObject{{ID: cardID}},
		[]Mod{{Kind: ModExileInsteadOfLeaving}}, ForThePermanentThisSpellBecomesDuration(cardID), label)
}
