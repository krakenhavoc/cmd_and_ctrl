package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Darkness Crystal — Legendary Artifact {2}{B}{B}:
//
//	"Black spells you cast cost {1} less to cast.
//	 If a nontoken creature an opponent controls would die, instead
//	 exile it and you gain 2 life.
//	 {4}{B}{B}, {T}: Put target creature card exiled with The Darkness
//	 Crystal onto the battlefield tapped under your control with two
//	 additional +1/+1 counters on it."
//
// Three lines, three existing shapes (#3213, the Kuja deck):
//
//   - The discount is the Water Crystal's, in black: YourSpell and
//     ColoredSpell("B"). It only ever takes generic mana off.
//   - The replacement is Liesa's "if a creature an opponent controls
//     would die, exile it instead" narrowed to nontoken creatures,
//     with the exile linked to the Crystal (ReplacementEvent.ExiledWith,
//     CR 607.2a). The creature never died (CR 700.4), so its
//     dies-triggers do not fire. A commander is exiled too and its
//     owner then decides about the command zone (CR 903.9a).
//   - "And you gain 2 life" is the second instruction of the modified
//     event (CR 608.2c, 614.6; ADR 0178, #2950): declared with
//     ChangeLifeAfterReplacedMoveForEffect, it lands after the exile,
//     through the life pipeline, and whatever the destroying effect does
//     next waits for it, including a CR 616 question between two life
//     replacements. The player and source are fixed as the replacement
//     applies, so the Crystal leaving afterwards changes nothing.
//   - The ability TARGETS a creature card whose link names this Crystal
//     (#3259): TargetCardInExile narrowed by ExiledWithThis, which is
//     the object, so a second Crystal's exiles are not offered. The card
//     is picked in the zone browser's exile tab when the ability is
//     activated, and re-checked as it resolves (CR 608.2b): a card that
//     left exile in response is an illegal target and the ability does
//     nothing. The counters are an "enters with" on the entry event, so
//     a counter replacement sees them and an enters trigger finds them;
//     the new object is tapped and under YOUR control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "660f370a-b25b-4ea5-b765-7abf96899c1c",
		Name:         "The Darkness Crystal",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Black spells you cast cost {1} less to cast.",
				YourSpell(), ColoredSpell("B")),
		},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if !rfOpponentCreatureWouldDie(ev, g, src) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !c.IsToken()
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				ev.ExiledWith = game.PermissionCardRef{ID: src.InstanceID, Epoch: src.ObjectEpoch}
				return g.ChangeLifeAfterReplacedMoveForEffect(ev, src.InstanceID, src.Controller, 2)
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "The Darkness Crystal: exile instead of dying, and gain 2 life",
		}},
		Activated: []ActivatedAbility{{
			Label: "{4}{B}{B}, {T}: Put target creature card exiled with The Darkness Crystal onto the battlefield tapped under your control with two additional +1/+1 counters on it.",
			Cost:  Plus(ManaCost("{4}{B}{B}"), TapCost()),
			Targets: ExiledWithThis(TargetCardInExile(
				"target creature card exiled with The Darkness Crystal", Creature())),
			Effect: darknessCrystalReturn,
		}},
	})
}

// darknessCrystalReturn puts each still-legal target onto the battlefield
// tapped under the Crystal controller's control with two +1/+1 counters.
// ctx.LegalTargets re-judges the clause (CR 608.2b), so a card that left
// exile in response, or whose link no longer names this Crystal, is skipped.
func darknessCrystalReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if err := (ReturnFromExile{
			Target:     t.ID,
			Controller: item.Controller,
			Tapped:     true,
			Counters:   map[string]int{game.CounterPlusOne: 2},
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
