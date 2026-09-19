package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// True-Name Nemesis — Creature — Merfolk Rogue, {1}{U}{U}, 3/1:
//
//	"As this creature enters, choose a player.
//	 This creature has protection from the chosen player."
//
// The card the choose-a-player seam and half of the protection seam
// were both named after, and it needed two halves that landed a sprint
// apart. #929 shipped the RESOLUTION-time player choice; #662 shipped
// protection against the source OBJECT. This is what was left: a CR
// 614.12 choice made AS the permanent enters and stored on it
// (ChoosePlayerAsEnters → game.Card.ChosenPlayer), and CR 702.16k's
// PLAYER quality in the protection grammar, which is the one quality
// tested against the source's CONTROLLER rather than against any
// characteristic of it. #980.
//
// The protection is a printed keyword, so it rides PrintedKeywords like
// Baneslayer Angel's two — not a static ability, not a grant. The token
// names no seat: it is game.ProtectionFromChosenPlayer verbatim, and
// which player it means is read off this permanent's own stored answer
// by the one grammar reader. That is why a Clone of the Nemesis chooses
// its OWN player (CR 707.2 copies printed values, and a chosen player
// is not one) and why a bounced Nemesis chooses again (CR 400.7).
//
// WHAT THE QUALITY BUYS, all four through the one predicate: the chosen
// player cannot target it, block it, damage it, or attach anything to
// it. In practice that is the whole card in a duel — an unblockable,
// unkillable 3/1 against exactly one seat — and it is correctly a
// blank against everybody else at a four-player table, which is the
// part a colour quality could not express.
//
// "Choose a player" is every seat, the controller included (CR 614.12
// names no restriction). Choosing yourself is legal and useless, and
// the engine offers it rather than inventing a narrowing the card does
// not print.
//
// No simplification. The one thing worth naming is shared with every
// other as-enters choice in the engine and is declared in
// game/choose_player.go: the prompt is queued from the AsEnters hook
// rather than by pausing the CR 614 replacement pipeline, so the
// Nemesis is on the battlefield with nobody chosen for as long as the
// prompt is open. Nothing can act in that window, and an unchosen
// player is nobody — the protection applies to no source at all rather
// than to every source.
func init() {
	Register(Spec{
		OracleID:     "112322ad-8f66-4cd4-98a1-f425d61a69ce",
		Name:         "True-Name Nemesis",
		Completeness: CompletenessFull,
		AsEnters:     ChoosePlayerAsEnters("True-Name Nemesis", Players),
		PrintedKeywords: []string{
			game.ProtectionFromChosenPlayer,
		},
	})
}
