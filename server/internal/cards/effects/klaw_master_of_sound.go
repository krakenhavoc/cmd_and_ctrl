package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Klaw, Master of Sound — Legendary Creature — Elemental Rogue Villain
// {2}{B}, 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Deathtouch
//	 Whenever you play a card from exile, Klaw gains indestructible until
//	 end of turn.
//	 Whenever Klaw deals combat damage to a player, look at the top card of
//	 that player's library, then exile it face down. You may play it for as
//	 long as it remains exiled. Mana of any type can be spent to cast a
//	 spell this way."
//
// "Play a card from exile" is Prosper's read (b20PlayedACardFromExile): a
// spell cast from exile or a land played from exile, by Klaw's controller,
// from any permission, Klaw's own included. The exile is Gonti's shape
// (gontiExileTopForPlay): face down, played for as long as it stays exiled,
// any mana type; the damaged player's library is the one exiled from and
// Klaw's controller holds the permission.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8eaa8833-ab9e-4bd8-92de-63e195b8d912",
		Name:            "Klaw, Master of Sound",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventCast, game.EventZoneMove}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b20PlayedACardFromExile(ev, source, g)
			}, "Klaw, Master of Sound — Klaw gains indestructible until end of turn",
				thisGainsKeywordUntilEndOfTurn("indestructible", "Klaw, Master of Sound — indestructible until end of turn")),
			WheneverThisDealsCombatDamageToAPlayer("Klaw, Master of Sound — look at the top card of that player's library, then exile it face down; you may play it",
				cmdr2S45ExileTopFaceDownToPlay),
		},
	})
}
