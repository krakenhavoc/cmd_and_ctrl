package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lord of Atlantis — Creature — Merfolk, {U}{U}, 2/2:
//
//	"Other Merfolk get +1/+1 and have islandwalk."
//
// The first S16 Layer-6 catalog card, and since S26 the canonical
// example of the shared lord builders in tribal.go.
//
// FIXED IN S26 — the anthem used to be restricted to the Lord's own
// controller. The printed card has no "you control" clause: Lord of
// Atlantis buffs every Merfolk on the battlefield, an opponent's
// included, which is exactly the sort of symmetrical drawback a 1993
// lord has and a modern one does not. Goblin King and Elvish Champion
// are in the same family and are written the same way; Goblin
// Chieftain and Elvish Archdruid really do say "you control" and set
// YoursOnly.
//
// The islandwalk grant is left in place. It is inert — CanBlock never
// sees the defending player's lands, so landwalk has nowhere to be
// enforced — and S26's new lords deliberately do NOT grant their own
// landwalk for that reason (see tribal.go). Removing this one would
// be a behaviour change to a shipped card in a sprint about something
// else; it goes when landwalk lands, alongside the others.
func init() {
	otherMerfolk := TribeFilter{Tribes: []string{"Merfolk"}, Others: true}
	Register(Spec{
		OracleID: "cc7f290f-ca00-4285-9bdb-4b4402444f30",
		Name:     "Lord of Atlantis",
		Static: []game.StaticAbility{
			TribalAnthem(otherMerfolk, 1, 1),
			TribalKeywordGrant(otherMerfolk, "islandwalk"),
		},
	})
}
