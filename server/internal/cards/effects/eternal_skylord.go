package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eternal Skylord — Creature — Zombie Wizard {4}{U}, 3/3 (EDHREC rank
// 4666):
//
//	"When this creature enters, amass Zombies 2.
//	 Zombie tokens you control have flying."
//
// The card that makes amass's SUBTYPE half visible. The Army the
// Skylord amasses is a Zombie — the token is created as one, and an
// Army that was already out as some other species becomes one
// (CR 701.47a's last sentence, `grantAmassSubtypeLocked`) — so the
// second ability finds it and the 0/2 Army flies. Nothing in this file
// knows about Armies; the static reads "Zombie token you control" and
// the amass is what makes that true.
//
// The grant is layer 6 through the shared `KeywordGrant`, which reads
// the EFFECTIVE subtypes, which is the whole point: a printed type
// line would never say "Orc Zombie Army".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "324a30cc-bddd-463c-9f28-c94c93a26780",
		Name:         "Eternal Skylord",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Eternal Skylord — amass Zombies 2",
				Do(Amass{Subtype: "Zombie", N: 2})),
		},
		Static: []game.StaticAbility{
			KeywordGrant(zombieTokenYouControl, "flying"),
		},
	})
}

// zombieTokenYouControl is "Zombie tokens you control" — the filter
// Eternal Skylord, Gleaming Overseer, Vizier of the Scorpion and
// Dreadhorde Twins all print.
//
// Not a TribeFilter: that one has no "token" dimension, and adding one
// for a single caller would put a field on a struct sixteen lords read.
func zombieTokenYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.IsToken() &&
		target.Controller == source.Controller && target.HasSubtype("Zombie")
}
