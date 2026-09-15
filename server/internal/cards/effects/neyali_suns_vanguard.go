package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Neyali, Suns' Vanguard — Legendary Creature — Human Rebel
// {2}{R}{W}, 3/3 (EDHREC rank 3426):
//
//	"Attacking tokens you control have double strike.
//	 Whenever one or more tokens you control attack a player, exile
//	 the top card of your library. During any turn you attacked with
//	 a token, you may play that card."
//
// The Boros token commander. Both printed abilities ride ONE
// triggered ability, "whenever one or more creature tokens you
// control attack" — deduped per source and label across
// PendingTriggers, the stack and the pick queue, so a wide attack
// fires it once — whose body (b32NeyaliAttack) does both sentences:
//
//   - The static, as an until-end-of-turn grant of double strike to
//     every attacking token the controller controls, snapshotted as
//     the trigger resolves. A static that read "attacking" would not
//     work: an attack declaration does not rebuild the layer cache,
//     so the keyword would not be there when combat damage looked
//     for it. Two abilities on the same event would also queue a
//     CR 603.3b order prompt on every attack, for no decision worth
//     making.
//   - The exile, if a token is attacking a PLAYER as the trigger
//     resolves (a token declared against a planeswalker first and
//     one against a player second still count): the top card of the
//     controller's library, with permission to play it this turn —
//     and every card an earlier resolution of this trigger exiled
//     that is still in exile is granted again for this turn. The
//     turns you attack with a token are exactly the turns this
//     trigger resolves, so re-granting on each resolution is
//     "during any turn you attacked with a token".
//
// Sandbox simplifications, declared, all weaker than printed:
//
//   - The double strike arrives when the trigger resolves, not the
//     instant the token is declared; a token put onto the
//     battlefield attacking after that has none.
//   - The exiled cards' permission depends on Neyali's trigger
//     resolving: a turn on which tokens attacked only a planeswalker
//     or a battle, or on which Neyali was gone, re-grants nothing,
//     where printed the permission outlives her.
//   - "This turn" is the engine's impulse-exile window, which ends at
//     the controller's cleanup.
func init() {
	Register(Spec{
		OracleID:     "4012b400-7dcd-43d6-8806-39a3cb743d8f",
		Name:         "Neyali, Suns' Vanguard",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Attacking tokens get double strike when Neyali's attack trigger resolves, not the moment they're declared.",
			"A card exiled with Neyali can be played on a later turn only when her trigger resolves that turn — tokens must attack a player while she is on the battlefield.",
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b32TokenYouControlAttacked(ev, source, g) &&
					!b12TriggerPendingOrOnStack(g, source, b32NeyaliLabel)
			}, b32NeyaliLabel, b32NeyaliAttack),
		},
	})
}
