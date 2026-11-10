package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gate Colossus — Artifact Creature — Construct {8}, 8/8 (Foundations):
//
//	"Affinity for Gates
//	 This creature can't be blocked by creatures with power 2 or less.
//	 Whenever a Gate you control enters, you may put this card from your
//	 graveyard on top of your library."
//
// Affinity is the shared self cost modifier over Gates; the evasion is
// the shared block rule. The trigger watches from the graveyard, like
// Bloodghast's landfall: "you control" there is the owner (CR 108.4),
// and the put goes through the shared tuck, so a card that has already
// left the graveyard by the time the ability resolves does nothing
// (CR 400.7).
func init() {
	Register(Spec{
		OracleID:     "386c5895-9713-4c28-9ba8-8ca0caa06c13",
		Name:         "Gate Colossus",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for Gates", HasSubtype("Gate")),
		},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
		Triggered: []game.TriggeredAbility{
			Optional(InGraveyard(On(game.EventETB, gateEnteredUnderYourControl,
				"Gate Colossus — put it from your graveyard on top of your library",
				putThisCardFromYourGraveyardOnTop)),
				"Gate Colossus — put it on top of your library?"),
		},
	})
}

// gateEnteredUnderYourControl — "whenever a Gate you control enters".
func gateEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.HasSubtypeWord("Gate")
}

// putThisCardFromYourGraveyardOnTop is "put this card from your
// graveyard on top of your library": the card is its own payload.
func putThisCardFromYourGraveyardOnTop(g *game.Game, item *game.StackItem) error {
	if sourceIsNewObject(g, item) {
		return nil
	}
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.TuckToLibraryForEffect(item.SourceCardID, false)
}
