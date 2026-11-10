package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Timothar, Baron of Bats — Legendary Creature — Vampire Noble {4}{B}{B},
// 4/4 (Commander, #5368):
//
//	"Ward—Discard a card.
//	 Whenever another nontoken Vampire you control dies, you may pay {1}
//	 and exile it. If you do, create a 1/1 black Bat creature token with
//	 flying. It gains "When this token deals combat damage to a player,
//	 sacrifice it and return the exiled card to the battlefield
//	 tapped.""
//
// The ward is Forum Necroscribe's. The trigger fires for each other
// nontoken Vampire (read as it last existed, CR 603.10a) that was under
// your control when it died. {1} is paid as the trigger resolves, then the
// card is exiled; only a card that reached exile makes the Bat. The exiled
// card is linked to that Bat (CR 607.2a, game.RelinkExiledWithForEffect),
// not to Timothar, so several Bats each return their own card. The Bat has
// the granted trigger for as long as it stays on the battlefield; when it
// deals combat damage to a player it is sacrificed and the linked card
// returns tapped. A Bat that dies or is exiled first leaves the card in
// exile for good.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82e03151-3990-451c-99cc-c7d386872df2",
		Name:         "Timothar, Baron of Bats",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: cmdr2B09TimotharBatKey,
			Triggered: []game.TriggeredAbility{
				Ward(WardDiscard(1), "Timothar, Baron of Bats — ward, discard a card"),
				WheneverThisDealsCombatDamageToAPlayer("Timothar's Bat — sacrifice it and return the exiled card to the battlefield tapped",
					cmdr2B09BatReturnsTheCard),
			},
			Text: "When this token deals combat damage to a player, sacrifice it and return the exiled card to the battlefield tapped.",
		}},
		Triggered: []game.TriggeredAbility{
			Ward(WardDiscard(1), "Timothar, Baron of Bats — ward, discard a card"),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && dead.InstanceID != source.InstanceID && !IsToken(dead) &&
					leftUnderControlOf(ev, dead) == source.Controller && leftAsSubtype(ev, dead, "Vampire")
			}, "Timothar, Baron of Bats — you may pay {1} and exile it; if you do, create a 1/1 Bat that returns it",
				cmdr2B09TimotharOffer),
		},
	})
}
