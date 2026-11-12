package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anowon, the Ruin Thief — Legendary Creature — Vampire Rogue {2}{U}{B},
// 2/4 (Zendikar Rising Commander, #5509):
//
//	"Other Rogues you control get +1/+1.
//	 Whenever one or more Rogues you control deal combat damage to a
//	 player, that player mills a card for each 1 damage dealt to them. If
//	 the player mills at least one creature card this way, you draw a
//	 card."
//
// The anthem is the tribal lord (TribalAnthem, "other", yours only).
//
// The trigger is ONE per damaged player and damage instance, settled once
// the combat damage step's events are all in (AtBatchEnd with
// ThatMuchDamage, CR 510.2, CR 603.2c): Rogues hitting two players at once
// are two triggers (the ruling), and a first-strike step and the regular
// step are two instances. The settled amount is the damage the Rogues you
// controlled dealt to that player in that instance and nothing else: the
// AppliesTo is re-run over the batch, so another creature's damage to the
// same player at the same time is not counted (the ruling). A Rogue is read
// through Anowon's own words (CR 612.2) and as it was when it dealt the
// damage.
//
// The mill is MillToZone with its registered tail, told the cards that
// reached the graveyard; one creature card among them draws one card, however
// many there are (the ruling).
//
// No simplification.
func init() {
	const label = "Anowon, the Ruin Thief — that player mills a card for each 1 damage dealt to them; if a creature card is milled, draw a card"
	Register(Spec{
		OracleID:     "e7e69583-8326-417e-8a9a-6b917c8ba4b6",
		Name:         "Anowon, the Ruin Thief",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Rogue"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{{
			Watches:    []game.EventKind{game.EventDealDamage},
			AppliesTo:  anowonRogueCombatDamageToAPlayer,
			Key:        label,
			AtBatchEnd: ThatMuchDamage(),
			Effect:     anowonMill,
		}},
	})
}

// anowonRogueCombatDamageToAPlayer is combat damage to a player dealt by a
// Rogue creature the source's controller controls.
func anowonRogueCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && OfCreatureType("Rogue")(g, source.Controller, game.ThroughTextOf(source, c))
}

// anowonMill mills the damaged player one card per point of the Rogues'
// damage, then draws through the tail.
func anowonMill(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	return MillToZone{
		Player:   item.Trigger.Event.Target,
		N:        settledEventAmount(item),
		ThenTail: ctx.ItemTail(anowonMilled),
	}.Apply(ctx)
}

// anowonMilled is "If the player mills at least one creature card this way,
// you draw a card": one card, for any number of creature cards milled.
var anowonMilled = CardTail("anowon-the-ruin-thief/after-mill", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			return DrawCards{N: 1}.Apply(ctx)
		}
	}
	return nil
})
