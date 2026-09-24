package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sephiroth, Fabled SOLDIER — Legendary Creature — Human Avatar
// Soldier {2}{B}, 3/3, the front face of a transforming card (Edea
// steal-and-sac deck, #1565):
//
//	"Whenever Sephiroth enters or attacks, you may sacrifice another
//	 creature. If you do, draw a card.
//	 Whenever another creature dies, target opponent loses 1 life and
//	 you gain 1 life. If this is the fourth time this ability has
//	 resolved this turn, transform Sephiroth."
//
// The back face is sephiroth_one_winged_angel.go, registered under
// "<oracle_id>#1".
//
// **Enters or attacks** is ONE ability watching two events (Sun
// Titan's shape). The sacrifice is chosen on resolution and is
// optional: a resolution-time pick of zero or one of your other
// creatures (ChoosePermanents, floor zero), sacrificed through the one
// sacrifice path, and the draw follows only from a sacrifice. It does
// not target, so a hexproof creature of yours can be sacrificed.
//
// **The drain** counts its own resolutions per object
// (Game.ResolvedThisTurn, keyed on the stack label and the source's
// CR 400.7 epoch). The count includes the resolution in progress, so
// the fourth resolution reads 4 and transforms. A drain countered on
// resolution for want of a legal target did not resolve and does not
// count. A Sephiroth that died and came back this turn is a new object
// and starts again from zero.
//
// **The transform** is the in-place verb (TransformThis, CR 701.27),
// and the back face's Super Nova emblem is made as it happens. See the
// back face for why only this transform makes it.
func init() {
	Register(Spec{
		OracleID:     sephirothOracleID,
		Name:         "Sephiroth, Fabled SOLDIER",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Sephiroth, Fabled SOLDIER — you may sacrifice another creature to draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return sephirothSacrificeOthersThenDraw(g, item, 1)
				}),
			Targeting(On(game.EventLTB, AnotherCreatureDied, sephirothDrainLabel, sephirothDrainAndMaybeTransform),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

const (
	sephirothOracleID   = "70113003-be5e-406a-9aec-cb480468c36d"
	sephirothDrainLabel = "Sephiroth, Fabled SOLDIER — target opponent loses 1 life and you gain 1 life"
)

// sephirothSacrificeOthersThenDraw is "you may sacrifice [up to max]
// other creatures. If you do, draw that many cards." max 0 is "any
// number". Shared by both faces: the front's attack-or-enter trigger
// (max 1) and the back's attack trigger (any number).
func sephirothSacrificeOthersThenDraw(g *game.Game, item *game.StackItem, max int) error {
	self := item.SourceCardID
	controller := item.Controller
	return ChoosePermanents{
		Question:  "Sephiroth — you may sacrifice other creatures to draw that many cards",
		Sacrifice: true,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.InstanceID != self && c.IsCreature() {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, max
		},
		Then: drawOnePerPicked(controller),
	}.Apply(NewContext(g, item))
}

// sephirothDrainAndMaybeTransform is the front face's death trigger:
// the drain, then the fourth-resolution transform.
func sephirothDrainAndMaybeTransform(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := drainTargetOpponentOne(g, item); err != nil {
		return err
	}
	if g.ResolvedThisTurn(item.SourceCardID, sephirothDrainLabel) != 4 {
		return nil
	}
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || TransformedPermanent(src) || ctx.isNewSourceObject(item.SourceCardID) {
		return nil
	}
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	if err := (TransformThis{}).Apply(ctx); err != nil {
		return err
	}
	// Super Nova: "As this creature transforms into Sephiroth,
	// One-Winged Angel, you get an emblem". Made only if the
	// permanent really is on its back face now (CR 701.27c: a
	// permanent that cannot transform does nothing).
	if after, ok := g.LookupCardForEffect(item.SourceCardID); !ok || !TransformedPermanent(after) {
		return nil
	}
	return CreateEmblem{}.Apply(ctx)
}
