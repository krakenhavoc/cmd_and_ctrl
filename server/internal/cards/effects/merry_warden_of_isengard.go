package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merry, Warden of Isengard — Legendary Creature — Halfling Advisor
// {1}{G}{W}, 1/4 (Tales of Middle-earth Commander):
//
//	"Partner with Pippin, Warden of Isengard (When this creature enters,
//	 target player may put Pippin into their hand from their library,
//	 then shuffle.)
//	 Whenever one or more artifacts you control enter, create a 1/1
//	 white Soldier creature token with lifelink. This ability triggers
//	 only once each turn."
//
// Partner with is the shared entry trigger; the artifact clause is Elvish
// Archivist's batched, once-a-turn shape: every artifact entering in one
// batch makes one trigger, and a second batch the same turn makes none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b600345-2b89-45bc-98c8-609fdd08a5fd",
		Name:         "Merry, Warden of Isengard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Merry, Warden of Isengard", "Pippin, Warden of Isengard"),
			TriggersOnlyOnceEachTurn(OncePerBatch(On(game.EventETB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, false)
					return ok && c.IsArtifact()
				}, "Merry, Warden of Isengard — create a 1/1 white Soldier creature token with lifelink",
				Do(CreateToken{Template: TokenCard("1/1 white Soldier with lifelink"), N: 1})))),
		},
	})
}
