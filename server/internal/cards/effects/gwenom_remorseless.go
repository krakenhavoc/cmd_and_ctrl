package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gwenom, Remorseless — {3}{B}{B} Legendary Creature — Symbiote Spider
// Hero 4/4 (Marvel's Spider-Man promos, #4460):
//
//	"Deathtouch, lifelink
//	 Whenever Gwenom attacks, until end of turn, you may look at the top
//	 card of your library any time and you may play cards from the top
//	 of your library. If you cast a spell this way, pay life equal to
//	 its mana value rather than pay its mana cost."
//
// Bolas's Citadel's three halves of CR 401.5, granted by a resolution
// for the turn instead of derived from a permanent: one stored standing
// permission over the top of the controller's own library that carries
// the look (SeesLibraryTop, Xanathar's grant) and Citadel's price
// (LifeEqualToManaValue, an alternative cost with a life component, so a
// player without the life is not offered it, CR 119.4). "Play", so a
// land may be played from the top too, using the turn's land play and
// paying no life. The grant's zero duration is "until end of turn", and
// it outlasts Gwenom: once the trigger has resolved, the permission is
// the player's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c48cf135-b280-4e9f-a9d9-225104870741",
		Name:            "Gwenom, Remorseless",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Gwenom, Remorseless — until end of turn, look at and play cards from the top of your library, paying life for spells",
				gwenomGrantTopOfLibrary),
		},
	})
}

// gwenomGrantTopOfLibrary is the attack trigger's body.
func gwenomGrantTopOfLibrary(g *game.Game, item *game.StackItem) error {
	g.GrantCastPermissionForEffect(game.CastPermission{
		Player:               item.Controller,
		Zone:                 game.ZoneLibrary,
		Scope:                game.ScopeStanding,
		TopOfLibraryOnly:     true,
		SeesLibraryTop:       true,
		LifeEqualToManaValue: true,
		AltCostKey:           "gwenom_remorseless",
		Source:               item.SourceCardID,
		Label:                "Pay life equal to its mana value (Gwenom, Remorseless)",
	})
	return nil
}
