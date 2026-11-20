package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prowess of the Fair — Kindred Enchantment — Elf {1}{B} (Lorwyn, #5593):
//
//	"Whenever another nontoken Elf is put into your graveyard from the
//	 battlefield, you may create a 1/1 green Elf Warrior creature token."
//
// "Elf" is any permanent of the type, a Kindred one as well as a creature,
// read as it last existed on the battlefield (CR 603.10a), so a changeling
// counts. "Your graveyard" is an OWNER clause: your Elf that an opponent
// stole and sacrificed still goes to your graveyard and triggers, and their
// Elf you stole does not. A token never counts. The token is the "you may"
// (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "31324235-8a2b-4ce4-b0d4-6d947489b930",
		Name:         "Prowess of the Fair",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(
				On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := lrwB1PutIntoAGraveyard(ev, g, source.TextSubtype("Elf"))
					return ok && c.InstanceID != source.InstanceID && !c.IsToken() && c.Owner == source.Controller
				}, "Prowess of the Fair — create a 1/1 green Elf Warrior",
					Do(CreateToken{Template: TokenCard("1/1 green Elf Warrior"), N: 1})),
				"Prowess of the Fair — create a 1/1 green Elf Warrior creature token?"),
		},
	})
}
