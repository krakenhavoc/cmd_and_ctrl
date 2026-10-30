package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Thing, Ben Grimm — {5}{G} Legendary Creature — Human Hero 7/7
// (Marvel Super Heroes, #4455):
//
//	"Trample
//	 Whenever one or more Heroes you control deal damage to a player, put
//	 two +1/+1 counters on The Thing."
//
// No "combat" in the printed text, so a Hero's noncombat damage counts too
// (damageToPlayerBy, Breeches's read). It is CR 603.2c's collapse per player
// (OncePerBatchPerPlayer): two Heroes connecting with one player in one
// damage step are one trigger, and a second player hit in the same step is a
// second. The Thing is a Hero himself, so his own damage counts. "Hero" is
// a word of the source's text and creature types are read as they are now.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d566d5f1-5abd-405b-8553-d267ecbe72ab",
		Name:            "The Thing, Ben Grimm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventDealDamage, mshB9HeroDealtDamageToAPlayer,
				"The Thing — put two +1/+1 counters on him", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2}))),
		},
	})
}

// mshB9HeroDealtDamageToAPlayer is the When of "one or more Heroes you
// control deal damage to a player": a Hero the source's controller controls
// dealt damage, of any kind, to a player.
func mshB9HeroDealtDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !damageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && dealer.HasSubtype(source.TextSubtype("Hero"))
}
