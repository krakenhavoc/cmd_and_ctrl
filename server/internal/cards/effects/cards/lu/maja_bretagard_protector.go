package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0a2075b5-9609-433d-bcd2-e0a637456cf8",
		Name:         "Maja, Bretagard Protector",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Maja, Bretagard Protector — create a 1/1 white Human Warrior", game.LandfallPayoff{Tokens: 1}, effects.B33CreateTokenBody(effects.B33WhiteHumanWarriorToken)),
		},
	})
}
