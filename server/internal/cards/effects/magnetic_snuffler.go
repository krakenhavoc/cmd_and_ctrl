package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magnetic Snuffler — Artifact Creature — Construct {5}, 4/4 (Murders at
// Karlov Manor, #4469):
//
//	"When this creature enters, return target Equipment card from your
//	 graveyard to the battlefield attached to this creature.
//	 Whenever you sacrifice an artifact, put a +1/+1 counter on this
//	 creature."
//
// The return is Unfinished Business's: the Equipment card enters attached
// to the Snuffler (ReturnFromGraveyardTogether.AttachedTo), under its
// owner's control, which is you. A Snuffler that has left the battlefield
// before the trigger resolves, or left and come back as a new object (CR
// 400.7), has nothing to be attached to: the Equipment still returns, and
// enters unattached. The sacrifice trigger is Crime Novelist's
// (b12YouSacrificedAnArtifact), so the Snuffler sacrificing itself is an
// artifact you sacrificed too, and its counter lands on nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f85088c-256d-4925-a4dd-aafcc2632282",
		Name:         "Magnetic Snuffler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Magnetic Snuffler — return target Equipment card from your graveyard attached to this creature",
					magneticSnufflerReturnEquipment),
				TargetCardInGraveyard("target Equipment card from your graveyard", HasSubtype("Equipment"), YouOwn())),
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b12YouSacrificedAnArtifact(ev, source, g)
			}, "Magnetic Snuffler — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

// magneticSnufflerReturnEquipment returns the targeted Equipment card,
// attached to the Snuffler while the Snuffler is still the object whose
// ability this is.
func magneticSnufflerReturnEquipment(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := legalTargetCardIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	ret := ReturnFromGraveyardTogether{Targets: ids}
	host := item.SourceCardID
	if onBattlefield(g, host) && !sourceIsNewObject(g, item) {
		if card, ok := g.LookupCardForEffect(ids[0]); ok && canBeAttachedTo(g, item.Controller, card, host) {
			ret.AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
		}
	}
	return ret.Apply(ctx)
}
