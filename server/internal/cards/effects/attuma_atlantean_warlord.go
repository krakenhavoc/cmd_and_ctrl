package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Attuma, Atlantean Warlord — Legendary Creature — Merfolk Warrior Villain
// {2}{U}{U}, 3/4 (Marvel Super Heroes):
//
//	"Other Merfolk you control get +1/+1.
//	 Whenever one or more Merfolk you control attack a player, draw a card."
//
// The first line is Master of the Pearl Trident's anthem without the
// islandwalk. The attack trigger is Horizon Explorer's "attack a player"
// guard (CR 603.2c): one declaration collapses to a draw per player
// attacked, however many Merfolk go at that player, and a second player
// attacked is a second draw. Attuma himself is a Merfolk, so he counts.
// "Merfolk" is a word of the source's text and creature subtypes are read
// as they are now, so a changeling counts.
//
// No simplification.
func init() {
	yourOtherMerfolk := TribeFilter{Tribes: []string{"Merfolk"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:     "e3b70a85-4296-411f-a673-7992b0c8a9aa",
		Name:         "Attuma, Atlantean Warlord",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(yourOtherMerfolk, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventAttack, mshB1MerfolkAttackAPlayer,
				"Attuma, Atlantean Warlord — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}

// mshB1MerfolkAttackAPlayer is the When of "whenever one or more Merfolk
// you control attack a player": a Merfolk the source's controller controls
// was declared attacking a player (not a planeswalker or battle).
func mshB1MerfolkAttackAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !b16YouAttackedAPlayer(ev, source, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Merfolk"))
}
