package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maja, Bretagard Protector — Legendary Creature — Human Warrior
// {2}{G}{W}{W}, 2/3 (EDHREC rank 3483):
//
//	"Other creatures you control get +1/+1.
//	 Landfall — Whenever a land you control enters, create a 1/1 white
//	 Human Warrior creature token."
//
// A landfall token maker under its own anthem. The anthem is a
// layer-7c TribalAnthem over the controller's OTHER creatures — the
// tokens she makes get it, Maja herself does not — and the landfall
// is enteredUnderYourControl narrowed to lands, so a fetched land
// and a played land both trigger, and an opponent's land does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0a2075b5-9609-433d-bcd2-e0a637456cf8",
		Name:         "Maja, Bretagard Protector",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b33LandYouControlEntered(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Maja, Bretagard Protector — create a 1/1 white Human Warrior",
					b33CreateTokenBody(b33WhiteHumanWarriorToken, 1))
			},
		}},
	})
}
