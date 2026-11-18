package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_dmu_b04_helpers.go — the shared pieces of Dominaria United batch 4
// (S82). One file for the batch so a parallel batch never edits the same one.

// dmuB04TabletTake is "put a land card from among them into your hand and the
// rest on the bottom of your library in a random order", with a draw when no
// card went to the hand (Inscribed Tablet). The take is mandatory when a
// land is there. An on-disk key: never renamed, never reused.
var dmuB04TabletTake = RegisterTake("choose-cards/inscribed-tablet-take",
	TakeDef{Then: func(g *game.Game, res TakeFromLibraryResult) error {
		if err := TakeRestOnBottomInRandomOrder(g, res); err != nil {
			return err
		}
		if len(res.Taken) > 0 {
			return nil
		}
		return g.DrawNForEffect(res.Player, 1)
	}})

// dmuB04VinewallTake is Floriferous Vinewall's take: the rest on the bottom
// in a random order. An on-disk key.
var dmuB04VinewallTake = RegisterTake("choose-cards/floriferous-vinewall-take",
	TakeDef{Reveal: true, Then: TakeRestOnBottomInRandomOrder})

// dmuB04PutCounterOnEachSoldier puts a +1/+1 counter on each Soldier the
// item's controller controls, one placement per Soldier. A Soldier that has
// left by then is skipped.
func dmuB04PutCounterOnEachSoldier(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	soldier := ctx.TextSubtype("Soldier") // a word of the source's text (CR 612.2)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.HasSubtype(soldier) {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// dmuB04NonSkeletonDiedUnderYourControl is "a non-Skeleton creature died
// under your control this turn" (Cult Conscript): the creatures that died
// under the player's control, less those that were Skeletons as they died
// (DiedWithSubtypeThisTurn reads the types the creature had on the
// battlefield, a changeling included, CR 603.10a). The word is read through
// the card whose text it is (CR 612.2).
func dmuB04NonSkeletonDiedUnderYourControl(g *game.Game, controller, source uuid.UUID) bool {
	skeleton := "Skeleton"
	if c, ok := g.LookupCardForEffect(source); ok {
		skeleton = c.TextSubtype("Skeleton")
	}
	return g.TurnTallyFor(controller).CreaturesDied > g.DiedWithSubtypeThisTurn(controller, skeleton)
}

// dmuB04ControlsCreatureWithPowerOverBase is Baird's "you control a creature
// with power greater than its base power": current power against the base
// power the layer pass recorded (CR 208.4b), as Kutzil reads it.
func dmuB04ControlsCreatureWithPowerOverBase(g *game.Game, you uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.IsCreature() && c.PowerForComparison() > c.BasePower() {
			return true
		}
	}
	return false
}
